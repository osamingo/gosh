package gosh_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"runtime"
	"testing"

	"github.com/osamingo/gosh"
)

type wrongJSONEncoder struct{}

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
	_, err := gosh.NewStatisticsHandler(nil)
	if err == nil {
		t.Fatal("expect occur an error")
	}
	h, err := gosh.NewStatisticsHandler(newJSONEncoder)
	if err != nil {
		t.Fatal(err)
	}
	if h == nil {
		t.Fatal("value is nil")
	}

	if _, ok := h.(*gosh.StatisticsHandler); !ok {
		t.Fatal("failed to cast to StatisticsHandler")
	}
}

func TestStatisticsHandler_ServeHTTP(t *testing.T) {
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
	}
	if resp.ContentLength == 0 {
		t.Fatal("response body should not be empty")
	}
}

func TestStatisticsHandler_ServeHTTPWithError(t *testing.T) {
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
	h, err := gosh.NewStatisticsHandler(newJSONEncoder)
	if err != nil {
		t.Fatal(err)
	}
	hh := h.(*gosh.StatisticsHandler)
	ss := make([]*gosh.Statistics, 100000)
	for i := 0; i < len(ss); i++ {
		s := hh.MeasureRuntime()
		ss[i] = &s
	}
}

func TestStatisticsHandler_MeasureRuntimeWithGC(t *testing.T) {
	defer func() {
		if err := recover(); err != nil {
			t.Fatal("panic occurred")
		}
	}()
	h, err := gosh.NewStatisticsHandler(newJSONEncoder)
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 256; i++ {
		runtime.GC()
	}
	hh := h.(*gosh.StatisticsHandler)
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
	hh := h.(*gosh.StatisticsHandler)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		hh.MeasureRuntime()
	}
}

func ExampleNewStatisticsHandler() {
	const path = "/healthz"

	h, err := gosh.NewStatisticsHandler(newJSONEncoder)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	mux := http.NewServeMux()
	mux.Handle(path, h)

	srv := httptest.NewServer(mux)
	defer srv.Close()

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, srv.URL+path, nil)
	if err != nil {
		log.Fatalln("failed to generate request")
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatalln(err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		fmt.Println("unexpect status code:", resp.StatusCode)
		os.Exit(1)
	}

	var s gosh.Statistics
	if err := json.NewDecoder(resp.Body).Decode(&s); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Printf("status_code: %d, has_gorutine: %t", resp.StatusCode, s.GoroutineNum > 0)

	// Output:
	// status_code: 200, has_gorutine: true
}
