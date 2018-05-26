# Gosh

## About

Go statistics handler

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
	mux.Handle("/healthz", gosh.NewStatsHandler())

	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatalln(err)
	}
}
```

### Output

```json
$ curl "localhost:8080/healthz" | jq .
{
  "timestamp": 1527313320320714500,
  "go_version": "go1.10.2",
  "go_os": "darwin",
  "go_arch": "amd64",
  "cpu_num": 8,
  "goroutine_num": 3,
  "gomaxprocs": 8,
  "cgo_call_num": 1,
  "memory_alloc": 450808,
  "memory_total_alloc": 450808,
  "memory_sys": 4262136,
  "memory_lookups": 15,
  "memory_mallocs": 5045,
  "memory_frees": 121,
  "memory_stack": 524288,
  "heap_alloc": 450808,
  "heap_sys": 1572864,
  "heap_idle": 557056,
  "heap_inuse": 1015808,
  "heap_released": 0,
  "heap_objects": 4924,
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
