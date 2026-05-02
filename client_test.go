package uptime_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tickstem/uptime"
)

func serverFunc(fn func(w http.ResponseWriter, r *http.Request)) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(fn))
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(body) //nolint:errcheck
}

var fixtureMonitor = &uptime.Monitor{
	ID:           "m1",
	Name:         "Production API",
	URL:          "https://api.example.com/health",
	IntervalSecs: 60,
	TimeoutSecs:  10,
	Status:       uptime.MonitorStatusActive,
	CreatedAt:    time.Now(),
	UpdatedAt:    time.Now(),
}

func TestList(t *testing.T) {
	t.Run("given monitors exist when listing then returns slice", func(t *testing.T) {
		srv := serverFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodGet, r.Method)
			assert.Equal(t, "/v1/monitors", r.URL.Path)
			writeJSON(w, http.StatusOK, map[string]any{"monitors": []*uptime.Monitor{fixtureMonitor}})
		})
		defer srv.Close()

		client := uptime.New("key", uptime.WithBaseURL(srv.URL+"/v1"))
		got, err := client.List(context.Background())

		require.NoError(t, err)
		require.Len(t, got, 1)
		assert.Equal(t, "m1", got[0].ID)
	})

	t.Run("given null monitors when listing then returns empty slice", func(t *testing.T) {
		srv := serverFunc(func(w http.ResponseWriter, r *http.Request) {
			writeJSON(w, http.StatusOK, map[string]any{"monitors": nil})
		})
		defer srv.Close()

		client := uptime.New("key", uptime.WithBaseURL(srv.URL+"/v1"))
		got, err := client.List(context.Background())

		require.NoError(t, err)
		assert.Empty(t, got)
	})
}

func TestGet(t *testing.T) {
	t.Run("given valid id when getting then returns monitor", func(t *testing.T) {
		srv := serverFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodGet, r.Method)
			assert.Equal(t, "/v1/monitors/m1", r.URL.Path)
			writeJSON(w, http.StatusOK, fixtureMonitor)
		})
		defer srv.Close()

		client := uptime.New("key", uptime.WithBaseURL(srv.URL+"/v1"))
		got, err := client.Get(context.Background(), "m1")

		require.NoError(t, err)
		assert.Equal(t, "m1", got.ID)
		assert.Equal(t, "Production API", got.Name)
	})

	t.Run("given unknown id when getting then returns not found error", func(t *testing.T) {
		srv := serverFunc(func(w http.ResponseWriter, r *http.Request) {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "not found"})
		})
		defer srv.Close()

		client := uptime.New("key", uptime.WithBaseURL(srv.URL+"/v1"))
		_, err := client.Get(context.Background(), "missing")

		require.Error(t, err)
		assert.True(t, uptime.IsNotFound(err))
	})
}

func TestCreate(t *testing.T) {
	t.Run("given valid params when creating then returns monitor", func(t *testing.T) {
		srv := serverFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodPost, r.Method)
			assert.Equal(t, "/v1/monitors", r.URL.Path)
			assert.Equal(t, "Bearer key", r.Header.Get("Authorization"))

			var body map[string]any
			require.NoError(t, json.NewDecoder(r.Body).Decode(&body))
			assert.Equal(t, "Production API", body["name"])
			assert.Equal(t, "https://api.example.com/health", body["url"])

			writeJSON(w, http.StatusCreated, fixtureMonitor)
		})
		defer srv.Close()

		client := uptime.New("key", uptime.WithBaseURL(srv.URL+"/v1"))
		got, err := client.Create(context.Background(), uptime.CreateParams{
			Name:         "Production API",
			URL:          "https://api.example.com/health",
			IntervalSecs: 60,
		})

		require.NoError(t, err)
		assert.Equal(t, "m1", got.ID)
	})

	t.Run("given plan limit reached when creating then returns quota error", func(t *testing.T) {
		srv := serverFunc(func(w http.ResponseWriter, r *http.Request) {
			writeJSON(w, http.StatusPaymentRequired, map[string]string{"error": "monitor quota reached"})
		})
		defer srv.Close()

		client := uptime.New("key", uptime.WithBaseURL(srv.URL+"/v1"))
		_, err := client.Create(context.Background(), uptime.CreateParams{
			Name: "x", URL: "https://x.com",
		})

		require.Error(t, err)
		assert.True(t, uptime.IsQuotaExceeded(err))
	})

	t.Run("given interval below plan minimum when creating then returns quota error", func(t *testing.T) {
		srv := serverFunc(func(w http.ResponseWriter, r *http.Request) {
			writeJSON(w, http.StatusPaymentRequired, map[string]string{"error": "interval_secs below plan minimum"})
		})
		defer srv.Close()

		client := uptime.New("key", uptime.WithBaseURL(srv.URL+"/v1"))
		_, err := client.Create(context.Background(), uptime.CreateParams{
			Name: "x", URL: "https://x.com", IntervalSecs: 10,
		})

		require.Error(t, err)
		assert.True(t, uptime.IsQuotaExceeded(err))
	})
}

func TestUpdate(t *testing.T) {
	t.Run("given valid params when updating then returns updated monitor", func(t *testing.T) {
		newName := "Updated"
		updated := *fixtureMonitor
		updated.Name = newName

		srv := serverFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodPatch, r.Method)
			assert.Equal(t, "/v1/monitors/m1", r.URL.Path)
			writeJSON(w, http.StatusOK, &updated)
		})
		defer srv.Close()

		client := uptime.New("key", uptime.WithBaseURL(srv.URL+"/v1"))
		got, err := client.Update(context.Background(), "m1", uptime.UpdateParams{Name: &newName})

		require.NoError(t, err)
		assert.Equal(t, "Updated", got.Name)
	})
}

func TestPause(t *testing.T) {
	t.Run("given active monitor when pausing then returns paused monitor", func(t *testing.T) {
		paused := *fixtureMonitor
		paused.Status = uptime.MonitorStatusPaused

		srv := serverFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodPatch, r.Method)
			assert.Equal(t, "/v1/monitors/m1/pause", r.URL.Path)
			writeJSON(w, http.StatusOK, &paused)
		})
		defer srv.Close()

		client := uptime.New("key", uptime.WithBaseURL(srv.URL+"/v1"))
		got, err := client.Pause(context.Background(), "m1")

		require.NoError(t, err)
		assert.Equal(t, uptime.MonitorStatusPaused, got.Status)
	})
}

func TestResume(t *testing.T) {
	t.Run("given paused monitor when resuming then returns active monitor", func(t *testing.T) {
		srv := serverFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodPatch, r.Method)
			assert.Equal(t, "/v1/monitors/m1/resume", r.URL.Path)
			writeJSON(w, http.StatusOK, fixtureMonitor)
		})
		defer srv.Close()

		client := uptime.New("key", uptime.WithBaseURL(srv.URL+"/v1"))
		got, err := client.Resume(context.Background(), "m1")

		require.NoError(t, err)
		assert.Equal(t, uptime.MonitorStatusActive, got.Status)
	})
}

func TestDelete(t *testing.T) {
	t.Run("given existing monitor when deleting then resolves without error", func(t *testing.T) {
		srv := serverFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodDelete, r.Method)
			assert.Equal(t, "/v1/monitors/m1", r.URL.Path)
			w.WriteHeader(http.StatusNoContent)
		})
		defer srv.Close()

		client := uptime.New("key", uptime.WithBaseURL(srv.URL+"/v1"))
		err := client.Delete(context.Background(), "m1")

		require.NoError(t, err)
	})
}

func TestChecks(t *testing.T) {
	statusCode := 200
	fixtureCheck := &uptime.MonitorCheck{
		ID:         "c1",
		MonitorID:  "m1",
		Status:     uptime.CheckStatusUp,
		StatusCode: &statusCode,
		DurationMs: 45,
		CheckedAt:  time.Now(),
	}

	t.Run("given checks exist when listing then returns slice", func(t *testing.T) {
		srv := serverFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodGet, r.Method)
			assert.Equal(t, "/v1/monitors/m1/checks", r.URL.Path)
			assert.Equal(t, "50", r.URL.Query().Get("limit"))
			writeJSON(w, http.StatusOK, map[string]any{"checks": []*uptime.MonitorCheck{fixtureCheck}})
		})
		defer srv.Close()

		client := uptime.New("key", uptime.WithBaseURL(srv.URL+"/v1"))
		got, err := client.Checks(context.Background(), "m1", uptime.ChecksParams{})

		require.NoError(t, err)
		require.Len(t, got, 1)
		assert.Equal(t, uptime.CheckStatusUp, got[0].Status)
		assert.Equal(t, int64(45), got[0].DurationMs)
	})

	t.Run("given null checks when listing then returns empty slice", func(t *testing.T) {
		srv := serverFunc(func(w http.ResponseWriter, r *http.Request) {
			writeJSON(w, http.StatusOK, map[string]any{"checks": nil})
		})
		defer srv.Close()

		client := uptime.New("key", uptime.WithBaseURL(srv.URL+"/v1"))
		got, err := client.Checks(context.Background(), "m1", uptime.ChecksParams{})

		require.NoError(t, err)
		assert.Empty(t, got)
	})

	t.Run("given explicit limit when listing then sends limit param", func(t *testing.T) {
		srv := serverFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "25", r.URL.Query().Get("limit"))
			writeJSON(w, http.StatusOK, map[string]any{"checks": []*uptime.MonitorCheck{}})
		})
		defer srv.Close()

		client := uptime.New("key", uptime.WithBaseURL(srv.URL+"/v1"))
		_, err := client.Checks(context.Background(), "m1", uptime.ChecksParams{Limit: 25})
		require.NoError(t, err)
	})
}

func TestErrorHelpers(t *testing.T) {
	t.Run("given 401 response when requesting then IsUnauthorized returns true", func(t *testing.T) {
		srv := serverFunc(func(w http.ResponseWriter, r *http.Request) {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		})
		defer srv.Close()

		client := uptime.New("bad-key", uptime.WithBaseURL(srv.URL+"/v1"))
		_, err := client.List(context.Background())

		require.Error(t, err)
		assert.True(t, uptime.IsUnauthorized(err))
		assert.False(t, uptime.IsQuotaExceeded(err))
		assert.False(t, uptime.IsNotFound(err))
	})

	t.Run("given API error when formatting then includes status and message", func(t *testing.T) {
		err := &uptime.APIError{StatusCode: 500, Message: "internal error"}
		assert.Equal(t, "tickstem/uptime: API error 500: internal error", err.Error())
	})

	t.Run("given non-JSON error body when parsing then uses raw body as message", func(t *testing.T) {
		srv := serverFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusServiceUnavailable)
			w.Write([]byte("service unavailable")) //nolint:errcheck
		})
		defer srv.Close()

		client := uptime.New("key", uptime.WithBaseURL(srv.URL+"/v1"))
		_, err := client.List(context.Background())

		require.Error(t, err)
		var apiErr *uptime.APIError
		require.ErrorAs(t, err, &apiErr)
		assert.Equal(t, 503, apiErr.StatusCode)
		assert.Equal(t, "service unavailable", apiErr.Message)
	})
}

func TestWithHTTPClient(t *testing.T) {
	t.Run("given custom http client when requesting then uses it", func(t *testing.T) {
		called := false
		transport := &recordingTransport{called: &called, inner: http.DefaultTransport}

		srv := serverFunc(func(w http.ResponseWriter, r *http.Request) {
			writeJSON(w, http.StatusOK, map[string]any{"monitors": nil})
		})
		defer srv.Close()

		client := uptime.New("key",
			uptime.WithBaseURL(srv.URL+"/v1"),
			uptime.WithHTTPClient(&http.Client{Transport: transport}),
		)
		_, err := client.List(context.Background())

		require.NoError(t, err)
		assert.True(t, called)
	})
}

type recordingTransport struct {
	called *bool
	inner  http.RoundTripper
}

func (rt *recordingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	*rt.called = true
	return rt.inner.RoundTrip(req)
}
