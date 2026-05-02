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

// AssertionSource identifies what part of the HTTP response an assertion checks.
type AssertionSource string

const (
	AssertionSourceStatusCode   AssertionSource = "status_code"
	AssertionSourceResponseTime AssertionSource = "response_time"
	AssertionSourceBody         AssertionSource = "body"
)

// AssertionComparison is the operator applied between the source value and the target.
type AssertionComparison string

const (
	AssertionComparisonEq          AssertionComparison = "eq"
	AssertionComparisonNe          AssertionComparison = "ne"
	AssertionComparisonLt          AssertionComparison = "lt"
	AssertionComparisonLte         AssertionComparison = "lte"
	AssertionComparisonGt          AssertionComparison = "gt"
	AssertionComparisonGte         AssertionComparison = "gte"
	AssertionComparisonContains    AssertionComparison = "contains"
	AssertionComparisonNotContains AssertionComparison = "not_contains"
)

// Assertion defines a condition that must hold for a check to be considered up.
// When a monitor has assertions they replace the default 2xx/3xx success logic.
type Assertion struct {
	Source     AssertionSource     `json:"source"`
	Comparison AssertionComparison `json:"comparison"`
	Target     string              `json:"target"`
}

// Monitor is a registered uptime monitor.
type Monitor struct {
	ID          string        `json:"id"`
	Name        string        `json:"name"`
	URL         string        `json:"url"`
	IntervalSecs int          `json:"interval_secs"`
	TimeoutSecs  int          `json:"timeout_secs"`
	Status       MonitorStatus `json:"status"`
	SSLExpiresAt *time.Time   `json:"ssl_expires_at,omitempty"`
	Assertions   []Assertion  `json:"assertions"`
	NextCheckAt  *time.Time   `json:"next_check_at,omitempty"`
	CreatedAt    time.Time    `json:"created_at"`
	UpdatedAt    time.Time    `json:"updated_at"`
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
	Name         string      `json:"name"`
	URL          string      `json:"url"`
	IntervalSecs int         `json:"interval_secs,omitempty"`
	TimeoutSecs  int         `json:"timeout_secs,omitempty"`
	Assertions   []Assertion `json:"assertions,omitempty"`
}

// UpdateParams are the fields that can be changed on an existing monitor.
// Only non-nil fields are sent to the API.
type UpdateParams struct {
	Name         *string     `json:"name,omitempty"`
	URL          *string     `json:"url,omitempty"`
	IntervalSecs *int        `json:"interval_secs,omitempty"`
	TimeoutSecs  *int        `json:"timeout_secs,omitempty"`
	Assertions   []Assertion `json:"assertions,omitempty"`
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
