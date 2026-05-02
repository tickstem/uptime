package uptime

import "time"

// MonitorStatus represents the current state of a monitor.
type MonitorStatus string

const (
	MonitorStatusActive  MonitorStatus = "active"
	MonitorStatusPaused  MonitorStatus = "paused"
	MonitorStatusFailing MonitorStatus = "failing"
)

// CheckStatus represents the outcome of a single HTTP check.
type CheckStatus string

const (
	CheckStatusUp      CheckStatus = "up"
	CheckStatusDown    CheckStatus = "down"
	CheckStatusTimeout CheckStatus = "timeout"
)

// Monitor is a registered uptime monitor.
type Monitor struct {
	ID          string        `json:"id"`
	Name        string        `json:"name"`
	URL         string        `json:"url"`
	IntervalSecs int          `json:"interval_secs"`
	TimeoutSecs  int          `json:"timeout_secs"`
	Status      MonitorStatus `json:"status"`
	NextCheckAt *time.Time    `json:"next_check_at,omitempty"`
	CreatedAt   time.Time     `json:"created_at"`
	UpdatedAt   time.Time     `json:"updated_at"`
}

// MonitorCheck is the result of a single HTTP check against a monitor's URL.
type MonitorCheck struct {
	ID           string      `json:"id"`
	MonitorID    string      `json:"monitor_id"`
	Status       CheckStatus `json:"status"`
	StatusCode   *int        `json:"status_code,omitempty"`
	DurationMs   int64       `json:"duration_ms"`
	Error        string      `json:"error,omitempty"`
	SSLExpiresAt *time.Time  `json:"ssl_expires_at,omitempty"`
	CheckedAt    time.Time   `json:"checked_at"`
}

// CreateParams are the parameters for creating a new monitor.
type CreateParams struct {
	Name         string `json:"name"`
	URL          string `json:"url"`
	IntervalSecs int    `json:"interval_secs,omitempty"`
	TimeoutSecs  int    `json:"timeout_secs,omitempty"`
}

// UpdateParams are the fields that can be changed on an existing monitor.
// Only non-nil fields are sent to the API.
type UpdateParams struct {
	Name         *string `json:"name,omitempty"`
	URL          *string `json:"url,omitempty"`
	IntervalSecs *int    `json:"interval_secs,omitempty"`
	TimeoutSecs  *int    `json:"timeout_secs,omitempty"`
}

// ChecksParams configures the Checks request.
type ChecksParams struct {
	Limit  int
	Offset int
}

type listMonitorsResponse struct {
	Monitors []*Monitor `json:"monitors"`
}

type listChecksResponse struct {
	Checks []*MonitorCheck `json:"checks"`
}
