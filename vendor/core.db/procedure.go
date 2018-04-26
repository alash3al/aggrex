package db

// Procedure .
type Procedure struct {
	Key  string   `json:"key"`
	Code string   `json:"code"`
	Tags []string `json:"tags"`
}
