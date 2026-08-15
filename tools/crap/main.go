// crap 関数ごとの CRAP 値（Change Risk Anti-Patterns）を計算し、閾値を超える関数があれば失敗する。
//
// CRAP(m) = comp(m)^2 * (1 - cov(m))^3 + comp(m)
//
//	comp: 循環的複雑度、cov: 文カバレッジ（0〜1）
//
// 使い方:
//
//	go test -coverprofile=coverage.out ./...
//	go run ./tools/crap -profile coverage.out -threshold 15
//
// 生成コード（sqlcgen）、テストファイル、main 関数、tools/ 配下は対象外にする。
package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

type funcCov struct {
	file string
	name string
	cov  float64
}

type result struct {
	file string
	name string
	comp int
	cov  float64
	crap float64
}

func main() {
	profile := flag.String("profile", "coverage.out", "go test -coverprofile の出力")
	threshold := flag.Float64("threshold", 15, "この値以上の CRAP を持つ関数があれば失敗する")
	top := flag.Int("top", 15, "表示する上位件数")
	flag.Parse()

	covs, err := loadFuncCoverage(*profile)
	if err != nil {
		fmt.Fprintln(os.Stderr, "coverage:", err)
		os.Exit(2)
	}
	results, err := compute(covs)
	if err != nil {
		fmt.Fprintln(os.Stderr, "compute:", err)
		os.Exit(2)
	}
	sort.Slice(results, func(i, j int) bool { return results[i].crap > results[j].crap })

	failed := 0
	fmt.Printf("%-8s %-5s %-7s %s\n", "CRAP", "CC", "COV", "FUNC")
	for i, r := range results {
		if i < *top || r.crap >= *threshold {
			fmt.Printf("%-8.2f %-5d %-6.1f%% %s:%s\n", r.crap, r.comp, r.cov*100, r.file, r.name)
		}
		if r.crap >= *threshold {
			failed++
		}
	}
	if failed > 0 {
		fmt.Printf("\nCRAP >= %.0f の関数が %d 件あります\n", *threshold, failed)
		os.Exit(1)
	}
	fmt.Printf("\n全 %d 関数が CRAP < %.0f を満たします\n", len(results), *threshold)
}

// loadFuncCoverage go tool cover -func の出力を関数ごとのカバレッジに変換する
func loadFuncCoverage(profile string) (map[string]funcCov, error) {
	cmd := exec.CommandContext(context.Background(), "go", "tool", "cover", "-func="+profile) //nolint:gosec // 固定引数
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("go tool cover: %w", err)
	}
	covs := make(map[string]funcCov)
	sc := bufio.NewScanner(strings.NewReader(string(out)))
	for sc.Scan() {
		fields := strings.Fields(sc.Text())
		if len(fields) < 3 || fields[0] == "total:" {
			continue
		}
		// fields[0] = path/file.go:line:, fields[1] = name, fields[2] = 12.3%
		loc := strings.TrimSuffix(fields[0], ":")
		parts := strings.Split(loc, ":")
		if len(parts) < 2 {
			continue
		}
		file := parts[0]
		if excluded(file) {
			continue
		}
		pct, err := strconv.ParseFloat(strings.TrimSuffix(fields[2], "%"), 64)
		if err != nil {
			continue
		}
		key := file + ":" + parts[1]
		covs[key] = funcCov{file: file, name: fields[1], cov: pct / 100}
	}
	return covs, sc.Err()
}

func excluded(file string) bool {
	return strings.Contains(file, "/sqlcgen/") ||
		strings.HasSuffix(file, "_test.go") ||
		strings.Contains(file, "/tools/") ||
		strings.Contains(file, "/cmd/")
}

// compute カバレッジのある各関数について循環的複雑度を求め CRAP を計算する
func compute(covs map[string]funcCov) ([]result, error) {
	// ファイルごとにまとめてパースする
	byFile := make(map[string][]string)
	for key, c := range covs {
		byFile[c.file] = append(byFile[c.file], key)
	}
	modRoot, err := moduleRoot()
	if err != nil {
		return nil, err
	}
	modPath, err := modulePath(modRoot)
	if err != nil {
		return nil, err
	}

	var results []result
	fset := token.NewFileSet()
	for file, keys := range byFile {
		rel := strings.TrimPrefix(file, modPath+"/")
		path := filepath.Join(modRoot, rel)
		f, err := parser.ParseFile(fset, path, nil, 0)
		if err != nil {
			return nil, fmt.Errorf("parse %s: %w", path, err)
		}
		lineToKey := make(map[int]string)
		for _, k := range keys {
			line, err := strconv.Atoi(strings.Split(k, ":")[1])
			if err != nil {
				continue
			}
			lineToKey[line] = k
		}
		ast.Inspect(f, func(n ast.Node) bool {
			fn, ok := n.(*ast.FuncDecl)
			if !ok || fn.Body == nil {
				return true
			}
			line := fset.Position(fn.Pos()).Line
			key, ok := lineToKey[line]
			if !ok {
				return true
			}
			if fn.Name.Name == "main" {
				return true
			}
			comp := complexity(fn)
			c := covs[key]
			crap := float64(comp*comp)*pow3(1-c.cov) + float64(comp)
			results = append(results, result{file: rel, name: c.name, comp: comp, cov: c.cov, crap: crap})
			return true
		})
	}
	return results, nil
}

func pow3(x float64) float64 { return x * x * x }

// complexity 循環的複雑度を数える（gocyclo と同じ規則: 1 + 分岐 + && / ||）
func complexity(fn *ast.FuncDecl) int {
	comp := 1
	ast.Inspect(fn, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.IfStmt, *ast.ForStmt, *ast.RangeStmt, *ast.CaseClause, *ast.CommClause:
			comp++
		case *ast.BinaryExpr:
			if x.Op == token.LAND || x.Op == token.LOR {
				comp++
			}
		}
		return true
	})
	return comp
}

func moduleRoot() (string, error) {
	out, err := exec.CommandContext(context.Background(), "go", "env", "GOMOD").Output()
	if err != nil {
		return "", err
	}
	return filepath.Dir(strings.TrimSpace(string(out))), nil
}

func modulePath(root string) (string, error) {
	data, err := os.ReadFile(filepath.Join(root, "go.mod")) //nolint:gosec // 固定パス
	if err != nil {
		return "", err
	}
	for _, line := range strings.Split(string(data), "\n") {
		if strings.HasPrefix(line, "module ") {
			return strings.TrimSpace(strings.TrimPrefix(line, "module ")), nil
		}
	}
	return "", fmt.Errorf("module 行が見つかりません")
}
