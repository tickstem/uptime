// Smoke test for the tickstem/uptime SDK.
// Runs against the real API — requires TICKSTEM_API_KEY to be set.
//
// Usage:
//
//	TICKSTEM_API_KEY=tsk_live_... go run ./cmd/smoke
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/tickstem/uptime"
)

func main() {
	apiKey := os.Getenv("TICKSTEM_API_KEY")
	if apiKey == "" {
		log.Fatal("TICKSTEM_API_KEY not set")
	}

	ctx := context.Background()
	client := uptime.New(apiKey)

	step("listing monitors")
	monitors, err := client.List(ctx)
	must(err)
	fmt.Printf("  found %d monitors\n", len(monitors))

	step("creating smoke monitor")
	monitor, err := client.Create(ctx, uptime.CreateParams{
		Name:         "tickstem-uptime-smoke-test",
		URL:          "https://tickstem.dev/healthz",
		IntervalSecs: 300,
	})
	must(err)
	fmt.Printf("  created: id=%s status=%s\n", monitor.ID, monitor.Status)

	step("getting monitor by id")
	got, err := client.Get(ctx, monitor.ID)
	must(err)
	fmt.Printf("  got: id=%s name=%s\n", got.ID, got.Name)

	step("pausing monitor")
	paused, err := client.Pause(ctx, monitor.ID)
	must(err)
	if paused.Status != uptime.MonitorStatusPaused {
		log.Fatalf("pause: expected status=paused, got %s", paused.Status)
	}
	fmt.Printf("  status=%s\n", paused.Status)

	step("resuming monitor")
	resumed, err := client.Resume(ctx, monitor.ID)
	must(err)
	if resumed.Status != uptime.MonitorStatusActive {
		log.Fatalf("resume: expected status=active, got %s", resumed.Status)
	}
	fmt.Printf("  status=%s\n", resumed.Status)

	step("listing checks (may be empty)")
	checks, err := client.Checks(ctx, monitor.ID, uptime.ChecksParams{Limit: 10})
	must(err)
	fmt.Printf("  found %d checks\n", len(checks))

	step("deleting monitor")
	must(client.Delete(ctx, monitor.ID))
	fmt.Println("  deleted")

	step("verifying deletion")
	time.Sleep(200 * time.Millisecond)
	_, err = client.Get(ctx, monitor.ID)
	if err == nil {
		log.Fatal("expected not-found error after deletion, got nil")
	}
	if !uptime.IsNotFound(err) {
		log.Fatalf("expected not-found error, got: %v", err)
	}
	fmt.Println("  confirmed not found")

	fmt.Println("\nsmoke test passed")
}

func step(name string) {
	fmt.Printf("\n%s...\n", name)
}

func must(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
