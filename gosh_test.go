package gosh_test

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/osamingo/gosh"
)

func TestNewStatisticsHandler(t *testing.T) {
	h := gosh.NewStatisticsHandler()
	if h == nil {
		t.Fatal("value is nil")
	}

	if _, ok := h.(*gosh.StatisticsHandler); !ok {
		t.Fatal("failed to cast to StatisticsHandler")
	}
}

func TestStatisticsHandler_ServeHTTP(t *testing.T) {

	srv := httptest.NewServer(gosh.NewStatisticsHandler())
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
	h := gosh.NewStatisticsHandler().(*gosh.StatisticsHandler)
	ss := make([]*gosh.Statistics, 100000)
	for i := 0; i < len(ss); i++ {
		s := h.MeasureRuntime()
		ss[i] = &s
	}
}

func BenchmarkStatisticsHandler_MeasureRuntime(b *testing.B) {
	h := gosh.NewStatisticsHandler().(*gosh.StatisticsHandler)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		h.MeasureRuntime()
	}
}

func ExampleNewStatisticsHandler() {

	const path = "/healthz"

	mux := http.NewServeMux()
	mux.Handle(path, gosh.NewStatisticsHandler())

	srv := httptest.NewServer(mux)
	defer srv.Close()

	resp, err := http.Get(srv.URL + path)
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
