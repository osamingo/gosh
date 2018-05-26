package gosh

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

var _ http.Handler = (*StatsHandler)(nil)

func TestNewStatsHandler(t *testing.T) {
	h := NewStatsHandler()
	if h == nil {
		t.Fatal("value is nil")
	}
	if h.lastSampledAt.IsZero() {
		t.Fatal("lastSampledAt shound not be zero")
	}
}

func TestStatsHandler_ServeHTTP(t *testing.T) {

	srv := httptest.NewServer(NewStatsHandler())
	defer srv.Close()

	resp, err := http.Get(srv.URL)
	if err != nil {
		t.Fatal("failed to request")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatal("unexpect status code")
	}
	if resp.ContentLength == 0 {
		t.Fatal("response body should not be empty")
	}
}

func TestStatsHandler_MeasureStats(t *testing.T) {
	defer func() {
		if err := recover(); err != nil {
			t.Fatal("panic occurred")
		}
	}()
	NewStatsHandler().MeasureStats()
}

func BenchmarkStatsHandler_MeasureStats(b *testing.B) {
	h := NewStatsHandler()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		h.MeasureStats()
	}
}
