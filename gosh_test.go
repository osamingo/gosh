package gosh

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewStatisticsHandler(t *testing.T) {
	h := NewStatisticsHandler()
	if h == nil {
		t.Fatal("value is nil")
	}
	if h.lastSampledAt.IsZero() {
		t.Fatal("lastSampledAt shound not be zero")
	}
}

func TestStatisticsHandler_ServeHTTP(t *testing.T) {

	srv := httptest.NewServer(NewStatisticsHandler())
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

func TestStatisticsHandler_MeasureRuntime(t *testing.T) {
	defer func() {
		if err := recover(); err != nil {
			t.Fatal("panic occurred")
		}
	}()
	h := NewStatisticsHandler()
	ss := make([]*Statistics, 100000)
	for i := 0; i < len(ss); i++ {
		s := h.MeasureRuntime()
		ss[i] = &s
	}
}

func BenchmarkStatisticsHandler_MeasureRuntime(b *testing.B) {
	h := NewStatisticsHandler()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		h.MeasureRuntime()
	}
}
