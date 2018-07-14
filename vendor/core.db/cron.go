package db

// Cron a cron job
type Cron struct {
	Interval string `json:"interval"`
	Job      string `json:"job"`
}
