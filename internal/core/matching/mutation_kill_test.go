package matching

import (
	"context"
	"math"
	"testing"
)

// 変異体（境界・算術・否定）を確実に殺すための厳密な数値・境界テスト

func almost(t *testing.T, name string, got, want float64) {
	t.Helper()
	if math.Abs(got-want) > 1e-9 {
		t.Errorf("%s = %.12f, want %.12f", name, got, want)
	}
}

func TestComputeTimeSeriesStats_ExactValues(t *testing.T) {
	s := ComputeTimeSeriesStats([]float64{1, 2, 3, 6})
	almost(t, "mean", s.Mean, 3.0)
	almost(t, "std", s.Std, 1.8708286933869707)
	almost(t, "min", s.Min, 1)
	almost(t, "max", s.Max, 6)
	almost(t, "trend", s.Trend, 1.6)
	almost(t, "volatility", s.Volatility, 0.2357022603955158)

	// 1点だけならトレンドとボラティリティは0、min=max
	one := ComputeTimeSeriesStats([]float64{5})
	almost(t, "one.mean", one.Mean, 5)
	almost(t, "one.std", one.Std, 0)
	almost(t, "one.trend", one.Trend, 0)
	almost(t, "one.volatility", one.Volatility, 0)
	almost(t, "one.min", one.Min, 5)
	almost(t, "one.max", one.Max, 5)

	// 前の値が0のリターンは0として扱う
	z := ComputeTimeSeriesStats([]float64{0, 2, 4})
	almost(t, "z.volatility", z.Volatility, 0.5) // returns = [0, 1] → mean 0.5, 分散 0.25
	almost(t, "z.trend", z.Trend, 2)
	almost(t, "z.min", z.Min, 0)
	almost(t, "z.max", z.Max, 4)

	// 減少系列は負のトレンド、min/max の更新方向を確認
	d := ComputeTimeSeriesStats([]float64{9, 4, 1})
	almost(t, "d.trend", d.Trend, -4)
	almost(t, "d.min", d.Min, 1)
	almost(t, "d.max", d.Max, 9)

	if e := ComputeTimeSeriesStats(nil); e.Mean != 0 || e.Max != 0 {
		t.Errorf("空入力はゼロ値: %+v", e)
	}
}

func TestNormalizeValue_Boundaries(t *testing.T) {
	almost(t, "min", NormalizeValue(10, 10, 20), 0)
	almost(t, "max", NormalizeValue(20, 10, 20), 1)
	almost(t, "mid", NormalizeValue(15, 10, 20), 0.5)
	almost(t, "below", NormalizeValue(5, 10, 20), 0)
	almost(t, "above", NormalizeValue(25, 10, 20), 1)
	almost(t, "degenerate", NormalizeValue(3, 7, 7), 0.5)
	almost(t, "quarter", NormalizeValue(12.5, 10, 20), 0.25)
}

func TestNewRanker_ClampsRandomFactor(t *testing.T) {
	if r := NewRanker(&RankingConfig{RandomFactor: -0.5}); r.Config.RandomFactor != 0 {
		t.Errorf("負の RandomFactor は 0 に: %v", r.Config.RandomFactor)
	}
	if r := NewRanker(&RankingConfig{RandomFactor: 1.5}); r.Config.RandomFactor != 1 {
		t.Errorf("1 超の RandomFactor は 1 に: %v", r.Config.RandomFactor)
	}
	if r := NewRanker(&RankingConfig{RandomFactor: 1}); r.Config.RandomFactor != 1 {
		t.Errorf("境界 1 はそのまま: %v", r.Config.RandomFactor)
	}
	if r := NewRanker(&RankingConfig{}); r.Config.Limit != 20 || r.Config.SortOrder != "desc" {
		t.Errorf("既定値: %+v", r.Config)
	}
}

func makeMatches(scores ...float64) []ScoredMatch {
	out := make([]ScoredMatch, 0, len(scores))
	for i, s := range scores {
		fv := NewFeatureVector(string(rune('a'+i)), "t")
		fv.SetCategorical("g", map[bool]string{true: "x", false: "y"}[i%2 == 0], 1)
		out = append(out, ScoredMatch{Candidate: fv, Score: s})
	}
	return out
}

func TestRanker_PaginationBoundaries(t *testing.T) {
	// offset == len → 空、offset+limit == len → ちょうど末尾まで
	r := NewRanker(&RankingConfig{Offset: 3, Limit: 2})
	if got := r.Rank(makeMatches(0.9, 0.8, 0.7)); len(got) != 0 {
		t.Errorf("offset==len は空のはず: %d", len(got))
	}
	r = NewRanker(&RankingConfig{Offset: 1, Limit: 2})
	got := r.Rank(makeMatches(0.9, 0.8, 0.7))
	if len(got) != 2 || got[0].Score != 0.8 || got[1].Score != 0.7 || got[0].Rank != 1 || got[1].Rank != 2 {
		t.Errorf("offset 1 limit 2: %+v", got)
	}
	r = NewRanker(&RankingConfig{Offset: 1, Limit: 5})
	if got := r.Rank(makeMatches(0.9, 0.8, 0.7)); len(got) != 2 {
		t.Errorf("limit 超過は末尾まで: %d", len(got))
	}
	// asc ソート
	r = NewRanker(&RankingConfig{SortOrder: "asc", Limit: 3})
	got = r.Rank(makeMatches(0.9, 0.7, 0.8))
	if got[0].Score != 0.7 || got[2].Score != 0.9 {
		t.Errorf("asc: %+v", got)
	}
}

func TestRanker_DiversityMaxPerGroupBoundary(t *testing.T) {
	// グループ x: a, c, e / y: b, d。max_per_group=2 → x は2件まで
	r := NewRanker(&RankingConfig{Limit: 10, Diversity: &DiversityConfig{Enabled: true, GroupByField: "g", MaxPerGroup: 2}})
	got := r.Rank(makeMatches(0.9, 0.8, 0.7, 0.6, 0.5))
	if len(got) != 4 {
		t.Fatalf("多様性で 4 件になるべき: %d", len(got))
	}
	// max_per_group=0 は無効
	r = NewRanker(&RankingConfig{Limit: 10, Diversity: &DiversityConfig{Enabled: true, GroupByField: "g", MaxPerGroup: 0}})
	if got := r.Rank(makeMatches(0.9, 0.8, 0.7)); len(got) != 3 {
		t.Errorf("MaxPerGroup 0 は無効: %d", len(got))
	}
	// グループ値がメタデータにある場合
	m := makeMatches(0.9, 0.8, 0.7)
	for i := range m {
		m[i].Candidate.Categorical = nil
		m[i].Candidate.SetMetadata("g", "same")
	}
	r = NewRanker(&RankingConfig{Limit: 10, Diversity: &DiversityConfig{Enabled: true, GroupByField: "g", MaxPerGroup: 1}})
	if got := r.Rank(m); len(got) != 1 {
		t.Errorf("メタデータのグループで 1 件: %d", len(got))
	}
	// カテゴリ値が全て0なら空文字グループとしてまとめて制限される
	m = makeMatches(0.9, 0.8)
	for i := range m {
		m[i].Candidate.Categorical["g"] = map[string]float64{"x": 0}
	}
	if got := r.Rank(m); len(got) != 1 {
		t.Errorf("空グループも MaxPerGroup で制限される: %d", len(got))
	}
}

func TestRanker_RandomnessKeepsSetAndRange(t *testing.T) {
	r := NewRanker(&RankingConfig{Limit: 10, RandomFactor: 0.1})
	in := makeMatches(0.5, 0.5, 0.5)
	got := r.Rank(in)
	if len(got) != 3 {
		t.Fatalf("件数が変わってはいけない: %d", len(got))
	}
	for _, g := range got {
		if g.Score < 0.45-1e-9 || g.Score > 0.55+1e-9 {
			t.Errorf("摂動は ±10%% 以内: %v", g.Score)
		}
	}
	// RandomFactor 0 なら不変
	r0 := NewRanker(&RankingConfig{Limit: 10, RandomFactor: 0})
	if got := r0.Rank(makeMatches(0.5)); got[0].Score != 0.5 {
		t.Errorf("RandomFactor 0 で変化: %v", got[0].Score)
	}
}

func TestValidateConfig_WeightBoundaries(t *testing.T) {
	c := &MatchingConfig{Version: "1", Domain: "d", Scoring: ScoringConfig{Components: []ComponentConfig{{Type: "categorical", Field: "f", Weight: 0}}}}
	if err := ValidateConfig(c); err != nil {
		t.Errorf("weight 0 は許容: %v", err)
	}
	c.Scoring.Components[0].Weight = 1
	if err := ValidateConfig(c); err != nil {
		t.Errorf("weight 1 は許容: %v", err)
	}
	if _, err := NewCompositeScorer([]ScoringComponent{{Similarity: &CategoricalSimilarity{Field: "f"}, Weight: 1}}); err != nil {
		t.Errorf("scorer weight 1 は許容: %v", err)
	}
	if _, err := NewCompositeScorer([]ScoringComponent{{Similarity: &CategoricalSimilarity{Field: "f"}, Weight: 1.0000001}}); err == nil {
		t.Error("scorer weight 1 超は拒否")
	}
}

func TestEngine_MinScoreBoundaryInclusive(t *testing.T) {
	// categorical 一致で 1.0、min_score 1.0 は含む（>= で採用）
	cfg := &MatchingConfig{Version: "1", Domain: "d",
		Scoring: ScoringConfig{MinScore: 1.0, Components: []ComponentConfig{{Name: "c", Type: "categorical", Field: "f", Weight: 1}}},
		Ranking: RankingConfig{Limit: 10}}
	eng, err := NewConfigurableEngine(cfg)
	if err != nil {
		t.Fatal(err)
	}
	src := NewFeatureVector("s", "t")
	src.SetCategorical("f", "v", 1)
	same := NewFeatureVector("a", "t")
	same.SetCategorical("f", "v", 1)
	diff := NewFeatureVector("b", "t")
	diff.SetCategorical("f", "w", 1)
	got, err := eng.FindMatches(context.Background(), src, []*FeatureVector{same, diff})
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].Candidate.ID != "a" {
		t.Errorf("min_score と等しいスコアは採用: %+v", got)
	}
}

func TestCompositeScorer_WeightedAverageExact(t *testing.T) {
	a := NewFeatureVector("a", "t")
	a.SetCategorical("f", "v", 1)
	a.SetCategorical("g", "v", 1)
	b := NewFeatureVector("b", "t")
	b.SetCategorical("f", "v", 1)
	b.SetCategorical("g", "w", 1)
	// f 一致(1.0, 重み0.75), g 不一致(0, 重み0.25) → 0.75
	scorer, _ := NewCompositeScorer([]ScoringComponent{
		{Name: "f", Similarity: &CategoricalSimilarity{Field: "f"}, Weight: 0.75},
		{Name: "g", Similarity: &CategoricalSimilarity{Field: "g"}, Weight: 0.25},
	})
	score, breakdown, err := scorer.Score(context.Background(), a, b)
	if err != nil {
		t.Fatal(err)
	}
	almost(t, "weighted", score, 0.75)
	almost(t, "breakdown.f", breakdown["f"], 1)
	almost(t, "breakdown.g", breakdown["g"], 0)

	// 変換で 1 を超える／負になる場合はクランプ
	up, _ := NewCompositeScorer([]ScoringComponent{{Name: "f", Similarity: &CategoricalSimilarity{Field: "f"}, Weight: 1, Transform: LinearTransform(2, 0.5)}})
	score, _, _ = up.Score(context.Background(), a, b)
	almost(t, "clamp high", score, 1)
	down, _ := NewCompositeScorer([]ScoringComponent{{Name: "f", Similarity: &CategoricalSimilarity{Field: "f"}, Weight: 1, Transform: LinearTransform(-2, 0)}})
	score, _, _ = down.Score(context.Background(), a, b)
	almost(t, "clamp low", score, 0)
}

func TestTransforms_ExactValues(t *testing.T) {
	almost(t, "gaussian mu", GaussianTransform(0.5, 0.1)(0.5), 1)
	almost(t, "gaussian 1sigma", GaussianTransform(0, 1)(1), math.Exp(-0.5))
	almost(t, "gaussian 2sigma", GaussianTransform(0, 2)(2), math.Exp(-0.5))
	almost(t, "inverse", InverseTransform()(3), 0.25)
	almost(t, "linear", LinearTransform(2, 3)(4), 11)
	almost(t, "sigmoid center", SigmoidTransform(2, 1)(1), 0.5)
	almost(t, "step below", StepTransform(0.5)(0.49), 0)
	almost(t, "step at", StepTransform(0.5)(0.5), 1)
}

func TestCosineSimilarity_ExactAndClamp(t *testing.T) {
	a := NewFeatureVector("a", "t")
	a.SetEmbedding("v", []float64{1, 0})
	b := NewFeatureVector("b", "t")
	b.SetEmbedding("v", []float64{1, 1})
	s, err := (&CosineSimilarity{VectorField: "v"}).Compute(a, b)
	if err != nil {
		t.Fatal(err)
	}
	almost(t, "cos45", s, 1/math.Sqrt2)
	c := NewFeatureVector("c", "t")
	c.SetEmbedding("v", []float64{-1, 0})
	s, _ = (&CosineSimilarity{VectorField: "v"}).Compute(a, c)
	almost(t, "opposite", s, -1)
	z := NewFeatureVector("z", "t")
	z.SetEmbedding("v", []float64{0, 0})
	s, _ = (&CosineSimilarity{VectorField: "v"}).Compute(a, z)
	almost(t, "zero vector", s, 0)
}

func TestCategoricalSimilarity_UsesMaxCategory(t *testing.T) {
	a := NewFeatureVector("a", "t")
	a.SetCategorical("f", "x", 0.2)
	a.SetCategorical("f", "y", 0.9)
	b := NewFeatureVector("b", "t")
	b.SetCategorical("f", "y", 0.4)
	b.SetCategorical("f", "x", 0.3)
	s, err := (&CategoricalSimilarity{Field: "f"}).Compute(a, b)
	if err != nil {
		t.Fatal(err)
	}
	almost(t, "max category y==y", s, 1)
	// 値が全て0のカテゴリは空扱いで不一致
	c := NewFeatureVector("c", "t")
	c.SetCategorical("f", "y", 0)
	s, _ = (&CategoricalSimilarity{Field: "f"}).Compute(a, c)
	almost(t, "zero-valued category", s, 0)
}

func TestFilters_ValueMustBePositive(t *testing.T) {
	fv := NewFeatureVector("a", "t")
	fv.SetCategorical("cat", "x", 0)
	other := NewFeatureVector("b", "t")
	eq, err := CreateFilter(FilterConfig{Field: "cat", Operator: "eq", Value: "x"})
	if err != nil {
		t.Fatal(err)
	}
	if eq(other, fv) {
		t.Error("値 0 のカテゴリは eq で一致しない")
	}
	fv.SetCategorical("cat", "x", 0.1)
	if !eq(other, fv) {
		t.Error("値が正なら一致")
	}
	in, err := CreateFilter(FilterConfig{Field: "cat", Operator: "in", Value: []any{"x", "z"}})
	if err != nil {
		t.Fatal(err)
	}
	fv.SetCategorical("cat", "x", 0)
	if in(other, fv) {
		t.Error("値 0 のカテゴリは in で一致しない")
	}
	fv.SetCategorical("cat", "x", 1)
	if !in(other, fv) {
		t.Error("値が正なら in で一致")
	}
}

func TestFeatureVector_SettersInitializeMaps(t *testing.T) {
	fv := &FeatureVector{ID: "x", Type: "t"}
	fv.SetEmbedding("e", []float64{1})
	fv.SetTimeSeries("ts", &TimeSeriesStats{Mean: 1})
	fv.SetMetadata("m", "v")
	fv.SetSparse("s", "i", 1)
	if fv.Embeddings["e"][0] != 1 || fv.TimeSeries["ts"].Mean != 1 || fv.Metadata["m"] != "v" || fv.Sparse["s"]["i"] != 1 {
		t.Errorf("セッターがマップを初期化していない: %+v", fv)
	}
	// 二回目の呼び出しで既存要素が残る
	fv.SetEmbedding("e2", []float64{2})
	fv.SetMetadata("m2", "w")
	fv.SetTimeSeries("ts2", &TimeSeriesStats{})
	if len(fv.Embeddings) != 2 || len(fv.Metadata) != 2 || len(fv.TimeSeries) != 2 {
		t.Errorf("既存要素が消えた: %+v", fv)
	}
}
