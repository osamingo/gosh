package main

import (
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"

	"github.com/osamingo/gosh"
)

func main() {

	const path = "/healthz"

	mux := http.NewServeMux()
	mux.Handle(path, gosh.NewStatisticsHandler())

	srv := httptest.NewServer(mux)
	defer srv.Close()

	resp, err := http.Get(srv.URL + path)
	if err != nil {
		log.Fatalln(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Fatalln("unexpect status code")
	}

	io.Copy(os.Stderr, resp.Body)
}
