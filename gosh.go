package gosh

import (
	"errors"
	"io"
	"math"
	"net/http"
	"runtime"
	"sync"
	"time"
)

type (
	// A Statistics has runtime information.
	Statistics struct {
		Timestamp        int64     `json:"timestamp"`
		GoVersion        string    `json:"go_version"`
		GoOS             string    `json:"go_os"`
		GoArch           string    `json:"go_arch"`
		CPUNum           int       `json:"cpu_num"`
		GoroutineNum     int       `json:"goroutine_num"`
		Gomaxprocs       int       `json:"gomaxprocs"`
		CgoCallNum       int64     `json:"cgo_call_num"`
		MemoryAlloc      uint64    `json:"memory_alloc"`
		MemoryTotalAlloc uint64    `json:"memory_total_alloc"`
		MemorySys        uint64    `json:"memory_sys"`
		MemoryLookups    uint64    `json:"memory_lookups"`
		MemoryMallocs    uint64    `json:"memory_mallocs"`
		MemoryFrees      uint64    `json:"memory_frees"`
		StackInuse       uint64    `json:"stack_inuse"`
		HeapAlloc        uint64    `json:"heap_alloc"`
		HeapSys          uint64    `json:"heap_sys"`
		HeapIdle         uint64    `json:"heap_idle"`
		HeapInuse        uint64    `json:"heap_inuse"`
		HeapReleased     uint64    `json:"heap_released"`
		HeapObjects      uint64    `json:"heap_objects"`
		GCNext           uint64    `json:"gc_next"`
		GCLast           uint64    `json:"gc_last"`
		GCNum            uint32    `json:"gc_num"`
		GCPerSecond      float64   `json:"gc_per_second"`
		GCPausePerSecond float64   `json:"gc_pause_per_second"`
		GCPause          []float64 `json:"gc_pause"`
	}
	// A StatisticsHandler provides runtime information handler.
	StatisticsHandler struct {
		NewJSONEncoder func(w io.Writer) JSONEncoder

		m                sync.Mutex
		lastSampledAt    time.Time
		lastPauseTotalNs uint64
		lastNumGC        uint32
	}

	// A JSONEncoder provides Encode method.
	// The Encode method writes the JSON encoding of v to the stream, followed by a newline character.
	JSONEncoder interface {
		Encode(v interface{}) error
	}
)

// NewStatisticsHandler returns new StatisticsHandler.
func NewStatisticsHandler(fn func(w io.Writer) JSONEncoder) (http.Handler, error) {
	if fn == nil {
		//nolint: goerr113
		return nil, errors.New("gosh: an argument should not be nil")
	}

	sh := &StatisticsHandler{
		NewJSONEncoder:   fn,
		m:                sync.Mutex{},
		lastSampledAt:    time.Time{},
		lastPauseTotalNs: 0,
		lastNumGC:        0,
	}
	sh.MeasureRuntime()

	return sh, nil
}

// ServeHTTP implements http.Handler interface.
func (sh *StatisticsHandler) ServeHTTP(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if err := sh.NewJSONEncoder(w).Encode(sh.MeasureRuntime()); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// MeasureRuntime accesses runtime information.
func (sh *StatisticsHandler) MeasureRuntime() Statistics {
	sh.m.Lock()
	defer sh.m.Unlock()

	var ms runtime.MemStats

	runtime.ReadMemStats(&ms)

	now := time.Now()

	var gcPausePerSec, gcPerSec float64
	if sh.lastPauseTotalNs > 0 {
		gcPausePerSec = time.Duration(ms.PauseTotalNs - sh.lastPauseTotalNs).Seconds()
	}

	gcCount := int(ms.NumGC - sh.lastNumGC)
	if sh.lastNumGC > 0 {
		gcPerSec = float64(gcCount) / now.Sub(sh.lastSampledAt).Seconds()
	}

	const gcCountThreshold = 256
	if gcCount > gcCountThreshold {
		gcCount = gcCountThreshold
	}

	gcPause := make([]float64, gcCount)
	for i := range gcCount {
		gcPause[i] = time.Duration(ms.PauseNs[(int(ms.NumGC)-i+math.MaxUint8)%gcCountThreshold]).Seconds()
	}

	sh.lastSampledAt = now
	sh.lastPauseTotalNs = ms.PauseTotalNs
	sh.lastNumGC = ms.NumGC

	return Statistics{
		Timestamp:        now.Unix(),
		GoVersion:        runtime.Version(),
		GoOS:             runtime.GOOS,
		GoArch:           runtime.GOARCH,
		CPUNum:           runtime.NumCPU(),
		GoroutineNum:     runtime.NumGoroutine(),
		Gomaxprocs:       runtime.GOMAXPROCS(0),
		CgoCallNum:       runtime.NumCgoCall(),
		MemoryAlloc:      ms.Alloc,
		MemoryTotalAlloc: ms.TotalAlloc,
		MemorySys:        ms.Sys,
		MemoryLookups:    ms.Lookups,
		MemoryMallocs:    ms.Mallocs,
		MemoryFrees:      ms.Frees,
		StackInuse:       ms.StackInuse,
		HeapAlloc:        ms.HeapAlloc,
		HeapSys:          ms.HeapSys,
		HeapIdle:         ms.HeapIdle,
		HeapInuse:        ms.HeapInuse,
		HeapReleased:     ms.HeapReleased,
		HeapObjects:      ms.HeapObjects,
		GCNext:           ms.NextGC,
		GCLast:           ms.LastGC,
		GCNum:            ms.NumGC,
		GCPerSecond:      gcPerSec,
		GCPausePerSecond: gcPausePerSec,
		GCPause:          gcPause,
	}
}
