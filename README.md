# Go Statistics Handler

[![Build Status](https://travis-ci.org/osamingo/gosh.svg?branch=master)](https://travis-ci.org/osamingo/gosh)
[![codecov](https://codecov.io/gh/osamingo/gosh/branch/master/graph/badge.svg)](https://codecov.io/gh/osamingo/gosh)
[![Go Report Card](https://goreportcard.com/badge/github.com/osamingo/gosh)](https://goreportcard.com/report/github.com/osamingo/gosh)
[![GoDoc](https://godoc.org/github.com/osamingo/gosh?status.svg)](https://godoc.org/github.com/osamingo/gosh)
[![GitHub license](https://img.shields.io/badge/license-MIT-blue.svg)](https://raw.githubusercontent.com/osamingo/gosh/master/LICENSE)

## About

- This Repository is provided following functions.
  - Go runtime statistics struct.
  - Go runtime statistics API handler.
  - Go runtime measure method.

## Install

```bash
$ go get -u github.com/osamingo/gosh
```

## Usage

### Example

```go
package main

import (
	"log"
	"net/http"

	"github.com/osamingo/gosh"
)

func main() {

	mux := http.NewServeMux()
	mux.Handle("/healthz", gosh.NewStatisticsHandler())

	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatalln(err)
	}
}
```

### Output

```json
$ curl "localhost:8080/healthz" | jq .
{
  "timestamp": 1527317620,
  "go_version": "go1.10.2",
  "go_os": "darwin",
  "go_arch": "amd64",
  "cpu_num": 8,
  "goroutine_num": 6,
  "gomaxprocs": 8,
  "cgo_call_num": 1,
  "memory_alloc": 422272,
  "memory_total_alloc": 422272,
  "memory_sys": 3084288,
  "memory_lookups": 6,
  "memory_mallocs": 4720,
  "memory_frees": 71,
  "stack_inuse": 491520,
  "heap_alloc": 422272,
  "heap_sys": 1605632,
  "heap_idle": 401408,
  "heap_inuse": 1204224,
  "heap_released": 0,
  "heap_objects": 4649,
  "gc_next": 4473924,
  "gc_last": 0,
  "gc_num": 0,
  "gc_per_second": 0,
  "gc_pause_per_second": 0,
  "gc_pause": []
}
```


## License

Released under the [MIT License](https://github.com/osamingo/gosh/blob/master/LICENSE).
