# tickstem/uptime

[![Go Reference](https://pkg.go.dev/badge/github.com/tickstem/uptime.svg)](https://pkg.go.dev/github.com/tickstem/uptime)
[![Go Report Card](https://goreportcard.com/badge/github.com/tickstem/uptime)](https://goreportcard.com/report/github.com/tickstem/uptime)
[![codecov](https://codecov.io/gh/tickstem/uptime/badge.svg)](https://codecov.io/gh/tickstem/uptime)

Go SDK for [Tickstem](https://tickstem.dev) — HTTP uptime monitoring with SSL expiry alerts and response assertions.

## Install

```bash
go get github.com/tickstem/uptime
```

## Quick start

```go
package main

import (
    "context"
    "fmt"
    "log"
    "os"

    "github.com/tickstem/uptime"
)

func main() {
    client := uptime.New(os.Getenv("TICKSTEM_API_KEY"))

    monitor, err := client.Create(context.Background(), uptime.CreateParams{
        Name: "Production API",
        URL:  "https://api.yourapp.com/health",
    })
    if err != nil {
        log.Fatal(err)
    }

    fmt.Println(monitor.ID)     // use to retrieve checks later
    fmt.Println(monitor.Status) // "active"
}
```

Get your API key at [app.tickstem.dev](https://app.tickstem.dev).

## Usage

### Create a client

```go
// Minimal — uses https://api.tickstem.dev/v1
client := uptime.New(os.Getenv("TICKSTEM_API_KEY"))

// With options
client := uptime.New(apiKey,
    uptime.WithBaseURL("http://localhost:8080/v1"),
)
```

### Create a monitor

```go
monitor, err := client.Create(ctx, uptime.CreateParams{
    Name:         "Production API",
    URL:          "https://api.yourapp.com/health",
    IntervalSecs: 60,   // default: 60, min: 60, max: 86400
    TimeoutSecs:  10,   // default: 10, min: 5, max: 30
})
```

### Add response assertions

Assertions replace the default 2xx/3xx success logic. All must pass for a check to be considered up.

```go
monitor, err := client.Create(ctx, uptime.CreateParams{
    Name: "API health check",
    URL:  "https://api.yourapp.com/health",
    Assertions: []uptime.Assertion{
        {Source: uptime.AssertionSourceStatusCode,   Comparison: uptime.AssertionComparisonEq,  Target: "200"},
        {Source: uptime.AssertionSourceResponseTime, Comparison: uptime.AssertionComparisonLt,  Target: "2000"},
        {Source: uptime.AssertionSourceBody,         Comparison: uptime.AssertionComparisonContains, Target: `"status":"ok"`},
    },
})
```

| Source          | Valid comparisons                          | Target type |
|-----------------|--------------------------------------------|-------------|
| `status_code`   | `eq` `ne` `lt` `lte` `gt` `gte`           | integer     |
| `response_time` | `eq` `ne` `lt` `lte` `gt` `gte`           | integer (ms)|
| `body`          | `eq` `ne` `contains` `not_contains`        | string      |

### Manage monitors

```go
// List all monitors
monitors, err := client.List(ctx)

// Get a single monitor
monitor, err := client.Get(ctx, monitorID)

// Update — only supplied fields are changed
err = client.Update(ctx, monitorID, uptime.UpdateParams{
    IntervalSecs: ptr(300),
})

// Pause / resume
_, err = client.Pause(ctx, monitorID)
_, err = client.Resume(ctx, monitorID)

// Delete
err = client.Delete(ctx, monitorID)
```

### Retrieve check history

```go
checks, err := client.Checks(ctx, monitorID, uptime.ChecksParams{
    Limit:  50,
    Offset: 0,
})
for _, c := range checks {
    fmt.Printf("%s  status=%s  duration=%dms\n", c.CheckedAt, c.Status, c.DurationMs)
    if c.SSLExpiresAt != nil {
        fmt.Printf("  SSL expires: %s\n", *c.SSLExpiresAt)
    }
}
```

### SSL certificate expiry

Tickstem automatically captures SSL certificate expiry on every HTTPS check. When a certificate is within 30 days of expiry you receive an email alert. The `Monitor.SSLExpiresAt` and `MonitorCheck.SSLExpiresAt` fields expose the raw expiry timestamp.

## Error handling

```go
monitor, err := client.Create(ctx, params)
if err != nil {
    if uptime.IsUnauthorized(err) {
        // invalid or revoked API key
    }
    if uptime.IsQuotaExceeded(err) {
        // monitor quota reached — upgrade at app.tickstem.dev/dashboard/billing
    }
    if uptime.IsNotFound(err) {
        // monitor does not exist or belongs to another account
    }
    var apiErr *uptime.APIError
    if errors.As(err, &apiErr) {
        fmt.Println(apiErr.StatusCode, apiErr.Message)
    }
}
```

## Pricing

| Plan     | Monitors  | Min interval | Price  |
|----------|-----------|--------------|--------|
| Free     | 5         | 60 s         | $0     |
| Starter  | 20        | 60 s         | $12/mo |
| Pro      | 100       | 60 s         | $29/mo |
| Business | unlimited | 60 s         | $79/mo |

[View full pricing →](https://tickstem.dev/#pricing)

## License

MIT
