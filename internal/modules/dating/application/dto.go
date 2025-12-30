package application

// MatchResult マッチング結果DTO
type MatchResult struct {
	UserID    string             `json:"user_id"`
	Score     float64            `json:"score"`
	Rank      int                `json:"rank"`
	Breakdown map[string]float64 `json:"breakdown"`
}

// MutualMatchResult 相互マッチ結果DTO
type MutualMatchResult struct {
	MatchID   string             `json:"match_id"`
	UserIDA   string             `json:"user_id_a"`
	UserIDB   string             `json:"user_id_b"`
	Score     float64            `json:"score"`
	Breakdown map[string]float64 `json:"breakdown"`
	MatchedAt string             `json:"matched_at"`
}

// LikeRequest いいねリクエストDTO
type LikeRequest struct {
	TargetUserID string `json:"target_user_id"`
}

// LikeResponse いいねレスポンスDTO
type LikeResponse struct {
	Matched bool   `json:"matched"`
	MatchID string `json:"match_id,omitempty"`
}
