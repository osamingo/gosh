package gosh_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"runtime"
	"testing"

	"github.com/osamingo/gosh"
)

type wrongJSONEncoder struct{}

//nolint: goerr113
func (e *wrongJSONEncoder) Encode(v interface{}) error {
	return fmt.Errorf("wrong json encoder")
}

func newJSONEncoder(w io.Writer) gosh.JSONEncoder {
	return json.NewEncoder(w)
}

func newWrongJSONEncoder(w io.Writer) gosh.JSONEncoder {
	return &wrongJSONEncoder{}
}

func TestNewStatisticsHandler(t *testing.T) {
	t.Parallel()

	_, err := gosh.NewStatisticsHandler(nil)
	if err == nil {
		t.Fatal("expect occur an error")
	}

	sh, err := gosh.NewStatisticsHandler(newJSONEncoder)
	if err != nil {
		t.Fatal(err)
	} else if sh == nil {
		t.Fatal("value is nil")
	}

	if _, ok := sh.(*gosh.StatisticsHandler); !ok {
		t.Fatal("failed to cast to StatisticsHandler")
	}
}

func TestStatisticsHandler_ServeHTTP(t *testing.T) {
	t.Parallel()

	h, err := gosh.NewStatisticsHandler(newJSONEncoder)
	if err != nil {
		t.Fatal(err)
	}

	srv := httptest.NewServer(h)
	defer srv.Close()

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, srv.URL, nil)
	if err != nil {
		t.Fatal("failed to generate request")
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal("failed to request")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatal("unexpect status code")
	} else if resp.ContentLength == 0 {
		t.Fatal("response body should not be empty")
	}
}

func TestStatisticsHandler_ServeHTTPWithError(t *testing.T) {
	t.Parallel()

	h, err := gosh.NewStatisticsHandler(newWrongJSONEncoder)
	if err != nil {
		t.Fatal(err)
	}

	srv := httptest.NewServer(h)
	defer srv.Close()

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, srv.URL, nil)
	if err != nil {
		t.Fatal("failed to generate request")
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal("failed to request")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatal("unexpect status code")
	} else if resp.ContentLength == 0 {
		t.Fatal("response body should not be empty")
	}
}

func TestStatisticsHandler_MeasureRuntime(t *testing.T) {
	t.Parallel()

	defer func() {
		if err := recover(); err != nil {
			t.Fatal("panic occurred")
		}
	}()

	h, err := gosh.NewStatisticsHandler(newJSONEncoder)
	if err != nil {
		t.Fatal(err)
	}

	hh, ok := h.(*gosh.StatisticsHandler)
	if !ok {
		t.Fatal("failed to cast *gosh.StatisticsHandler")
	}

	ss := make([]*gosh.Statistics, 100000)
	for i := 0; i < len(ss); i++ {
		s := hh.MeasureRuntime()
		ss[i] = &s
	}
}

func TestStatisticsHandler_MeasureRuntimeWithGC(t *testing.T) {
	t.Parallel()

	defer func() {
		if err := recover(); err != nil {
			t.Fatal("panic occurred")
		}
	}()

	sh, err := gosh.NewStatisticsHandler(newJSONEncoder)
	if err != nil {
		t.Fatal(err)
	}

	for i := 0; i < 256; i++ {
		runtime.GC()
	}

	hh, ok := sh.(*gosh.StatisticsHandler)
	if !ok {
		t.Fatal("failed to cast to *gosh.StatisticsHandler")
	}

	ss := make([]*gosh.Statistics, 100000)
	for i := 0; i < len(ss); i++ {
		s := hh.MeasureRuntime()
		ss[i] = &s
	}
}

func BenchmarkStatisticsHandler_MeasureRuntime(b *testing.B) {
	h, err := gosh.NewStatisticsHandler(newJSONEncoder)
	if err != nil {
		b.Fatal(err)
	}

	hh, ok := h.(*gosh.StatisticsHandler)
	if !ok {
		b.Fatal("failed to cast *gosh.StatisticsHandler")
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		hh.MeasureRuntime()
	}
}

func ExampleNewStatisticsHandler() {
	sh, err := gosh.NewStatisticsHandler(newJSONEncoder)
	if err != nil {
		fmt.Println(err)

		return
	}

	srv := httptest.NewServer(sh)

	defer srv.Close()

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, srv.URL, nil)
	if err != nil {
		fmt.Println(err)

		return
	}

	resp, err := srv.Client().Do(req)
	if err != nil {
		fmt.Println(err)

		return
	}

	defer func() {
		if err := resp.Body.Close(); err != nil {
			fmt.Println(err)

			return
		}
	}()

	if resp.StatusCode != http.StatusOK {
		fmt.Println("unexpect status code:", resp.StatusCode)

		return
	}

	var stats gosh.Statistics
	if err := json.NewDecoder(resp.Body).Decode(&stats); err != nil {
		fmt.Println(err)

		return
	}

	fmt.Printf("status_code: %d, has_gorutine: %t", resp.StatusCode, stats.GoroutineNum > 0)

	// Output:
	// status_code: 200, has_gorutine: true
}
