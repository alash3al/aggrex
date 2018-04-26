package db

import "time"

// Result the search result
type Result struct {
	Totals   uint64                   `json:"totals"`
	Hits     []map[string]interface{} `json:"hits"`
	MaxScore float64                  `json:"max_score"`
	Took     time.Duration            `json:"took"`
}
