package matching

import (
	"context"
	"errors"
	"testing"
)

// 設定の検証・構築の分岐を網羅する（CRAP 低減のための境界・異常系テスト）

func validComponent() ComponentConfig {
	return ComponentConfig{Name: "age", Type: "euclidean", Fields: []string{"age"}, Weight: 0.5}
}

func TestValidateConfig_Branches(t *testing.T) {
	base := func() *MatchingConfig {
		return &MatchingConfig{Version: "1.0", Domain: "d", Scoring: ScoringConfig{Components: []ComponentConfig{validComponent()}}}
	}
	cases := []struct {
		name   string
		mutate func(c *MatchingConfig)
		want   error
	}{
		{"version 欠落", func(c *MatchingConfig) { c.Version = "" }, ErrInvalidConfig},
		{"domain 欠落", func(c *MatchingConfig) { c.Domain = "" }, ErrInvalidConfig},
		{"components 空", func(c *MatchingConfig) { c.Scoring.Components = nil }, ErrNoComponents},
		{"type 欠落", func(c *MatchingConfig) { c.Scoring.Components[0].Type = "" }, ErrInvalidConfig},
		{"weight 負", func(c *MatchingConfig) { c.Scoring.Components[0].Weight = -0.1 }, ErrInvalidWeight},
		{"weight 超過", func(c *MatchingConfig) { c.Scoring.Components[0].Weight = 1.5 }, ErrInvalidWeight},
		{"euclidean fields 欠落", func(c *MatchingConfig) { c.Scoring.Components[0].Fields = nil }, ErrInvalidConfig},
		{"manhattan fields 欠落", func(c *MatchingConfig) {
			c.Scoring.Components[0].Type = "manhattan"
			c.Scoring.Components[0].Fields = nil
		}, ErrInvalidConfig},
		{"jaccard field 欠落", func(c *MatchingConfig) { c.Scoring.Components[0].Type = "jaccard" }, ErrInvalidConfig},
		{"categorical field 欠落", func(c *MatchingConfig) { c.Scoring.Components[0].Type = "categorical" }, ErrInvalidConfig},
		{"cosine field 欠落", func(c *MatchingConfig) { c.Scoring.Components[0].Type = "cosine" }, ErrInvalidConfig},
		{"未知の type", func(c *MatchingConfig) { c.Scoring.Components[0].Type = "time_series" }, ErrInvalidConfig},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			c := base()
			tc.mutate(c)
			if err := ValidateConfig(c); !errors.Is(err, tc.want) {
				t.Errorf("err = %v, want %v", err, tc.want)
			}
		})
	}
	valid := base()
	valid.Scoring.Components = append(valid.Scoring.Components,
		ComponentConfig{Name: "m", Type: "manhattan", Fields: []string{"x"}, Weight: 0.1},
		ComponentConfig{Name: "c", Type: "cosine", Field: "v", Weight: 0.1},
		ComponentConfig{Name: "j", Type: "jaccard", Field: "s", Weight: 0.1},
		ComponentConfig{Name: "k", Type: "categorical", Field: "cat", Weight: 0.1},
	)
	if err := ValidateConfig(valid); err != nil {
		t.Errorf("正常設定でエラー: %v", err)
	}
}

func TestBuildScorer_Branches(t *testing.T) {
	// 全タイプを構築できる
	cfg := &ScoringConfig{Components: []ComponentConfig{
		{Name: "e", Type: "euclidean", Fields: []string{"a"}, Weight: 0.2},
		{Name: "m", Type: "manhattan", Fields: []string{"a"}, Weight: 0.2},
		{Name: "c", Type: "cosine", Field: "v", Weight: 0.2},
		{Name: "j", Type: "jaccard", Field: "s", Weight: 0.2},
		{Name: "k", Type: "categorical", Field: "cat", Weight: 0.2, Transform: &TransformConfig{Type: "inverse"},
			Filter: &FilterConfig{Field: "a", Operator: "gt", Value: 0.0}},
	}}
	scorer, err := BuildScorer(cfg)
	if err != nil {
		t.Fatalf("BuildScorer: %v", err)
	}
	if len(scorer.Components) != 5 {
		t.Errorf("components = %d, want 5", len(scorer.Components))
	}
	if scorer.Components[4].Transform == nil || scorer.Components[4].Filter == nil {
		t.Error("transform/filter が設定されるべき")
	}

	// 未知タイプ
	if _, err := BuildScorer(&ScoringConfig{Components: []ComponentConfig{{Type: "nope", Weight: 0.5}}}); !errors.Is(err, ErrInvalidConfig) {
		t.Errorf("未知タイプ err = %v", err)
	}
	// 不正な transform
	if _, err := BuildScorer(&ScoringConfig{Components: []ComponentConfig{{Type: "categorical", Field: "c", Weight: 0.5, Transform: &TransformConfig{Type: "gaussian"}}}}); !errors.Is(err, ErrInvalidTransform) {
		t.Errorf("不正 transform err = %v", err)
	}
	// 不正な filter
	if _, err := BuildScorer(&ScoringConfig{Components: []ComponentConfig{{Type: "categorical", Field: "c", Weight: 0.5, Filter: &FilterConfig{Field: "a", Operator: "unknown_op"}}}}); err == nil {
		t.Error("不正 filter でエラーになるべき")
	}
	// 空 components は NewCompositeScorer が拒否する
	if _, err := BuildScorer(&ScoringConfig{}); !errors.Is(err, ErrNoComponents) {
		t.Errorf("空 components err = %v", err)
	}
}

func TestBuildTransform_AllTypes(t *testing.T) {
	ok := []TransformConfig{
		{Type: "linear", Params: map[string]float64{"a": 2, "b": 1}},
		{Type: "inverse"},
		{Type: "gaussian", Params: map[string]float64{"mu": 0, "sigma": 1}},
		{Type: "sigmoid", Params: map[string]float64{"k": 1, "x0": 0}},
		{Type: "step", Params: map[string]float64{"threshold": 0.5}},
	}
	for _, c := range ok {
		fn, err := buildTransform(&c)
		if err != nil || fn == nil {
			t.Errorf("%s: err=%v fn=nil?%v", c.Type, err, fn == nil)
		}
	}
	bad := []TransformConfig{
		{Type: "linear", Params: map[string]float64{"a": 1}},
		{Type: "gaussian", Params: map[string]float64{"mu": 0}},
		{Type: "sigmoid", Params: map[string]float64{"k": 1}},
		{Type: "step"},
		{Type: "unknown"},
	}
	for _, c := range bad {
		if _, err := buildTransform(&c); !errors.Is(err, ErrInvalidTransform) {
			t.Errorf("%s: err = %v, want ErrInvalidTransform", c.Type, err)
		}
	}
}

func TestCompositeScorer_Score_FilterAndCancel(t *testing.T) {
	a := NewFeatureVector("a", "t")
	a.SetNumerical("x", 1)
	a.SetCategorical("cat", "v", 1)
	b := NewFeatureVector("b", "t")
	b.SetNumerical("x", 1)
	b.SetCategorical("cat", "v", 1)

	// フィルタで全コンポーネントが除外されるとスコア0
	filterOut := func(*FeatureVector, *FeatureVector) bool { return false }
	scorer, err := NewCompositeScorer([]ScoringComponent{{Name: "c", Similarity: &CategoricalSimilarity{Field: "cat"}, Weight: 1, Filter: filterOut}})
	if err != nil {
		t.Fatal(err)
	}
	score, breakdown, err := scorer.Score(context.Background(), a, b)
	if err != nil || score != 0 || len(breakdown) != 0 {
		t.Errorf("filtered: score=%v breakdown=%v err=%v", score, breakdown, err)
	}

	// 距離関数 + 変換
	scorer, err = NewCompositeScorer([]ScoringComponent{{Name: "d", Distance: &EuclideanDistance{Fields: []string{"x"}}, Weight: 1, Transform: InverseTransform()}})
	if err != nil {
		t.Fatal(err)
	}
	score, _, err = scorer.Score(context.Background(), a, b)
	if err != nil || score <= 0 {
		t.Errorf("distance: score=%v err=%v", score, err)
	}

	// コンテキストキャンセル
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, _, err := scorer.Score(ctx, a, b); !errors.Is(err, context.Canceled) {
		t.Errorf("cancel: err = %v", err)
	}

	// 検証エラー（ID 無し）
	if _, _, err := scorer.Score(context.Background(), NewFeatureVector("", "t"), b); err == nil {
		t.Error("不正ベクトルAでエラーになるべき")
	}
	if _, _, err := scorer.Score(context.Background(), a, NewFeatureVector("", "t")); err == nil {
		t.Error("不正ベクトルBでエラーになるべき")
	}
	// 距離計算エラー（フィールド欠落）
	c := NewFeatureVector("c", "t")
	if _, _, err := scorer.Score(context.Background(), a, c); err == nil {
		t.Error("フィールド欠落でエラーになるべき")
	}
}

func TestFeatureVector_EnsureSparse(t *testing.T) {
	fv := NewFeatureVector("x", "t")
	fv.EnsureSparse("tags")
	if fv.Sparse["tags"] == nil || len(fv.Sparse["tags"]) != 0 {
		t.Fatalf("空集合として確保されるべき: %+v", fv.Sparse)
	}
	fv.SetSparse("tags", "a", 1)
	fv.EnsureSparse("tags")
	if len(fv.Sparse["tags"]) != 1 {
		t.Error("既存の要素を消してはいけない")
	}
}
