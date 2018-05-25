package gosh

import (
	"encoding/json"
	"net/http"
	"runtime"
	"sync"
	"time"
)

type (
	// A Stats has runtime, memory statistics information.
	Stats struct {
		Timestamp        int64     `json:"timestamp"`
		GoVersion        string    `json:"go_version"`
		GoOs             string    `json:"go_os"`
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
		StackInUse       uint64    `json:"memory_stack"`
		HeapAlloc        uint64    `json:"heap_alloc"`
		HeapSys          uint64    `json:"heap_sys"`
		HeapIdle         uint64    `json:"heap_idle"`
		HeapInuse        uint64    `json:"heap_inuse"`
		HeapReleased     uint64    `json:"heap_released"`
		HeapObjects      uint64    `json:"heap_objects"`
		GcNext           uint64    `json:"gc_next"`
		GcLast           uint64    `json:"gc_last"`
		GcNum            uint32    `json:"gc_num"`
		GcPerSecond      float64   `json:"gc_per_second"`
		GcPausePerSecond float64   `json:"gc_pause_per_second"`
		GcPause          []float64 `json:"gc_pause"`
	}
	// A StatsHandler provides runtime, memory statistics handler.
	StatsHandler struct {
		m                sync.Mutex
		lastSampledAt    time.Time
		lastPauseTotalNs uint64
		lastNumGC        uint32
	}
)

// NewStatsHandler returns gosh.StatsHandler.
func NewStatsHandler() *StatsHandler {
	h := &StatsHandler{}
	h.MeasureStats()
	return h
}

// ServeHTTP implements http.Handler interface.
func (sh *StatsHandler) ServeHTTP(w http.ResponseWriter, _ *http.Request) {
	if err := json.NewEncoder(w).Encode(sh.MeasureStats()); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
}

// MeasureStats accesses runtime, memory statistics information.
func (sh *StatsHandler) MeasureStats() Stats {

	sh.m.Lock()
	defer sh.m.Unlock()

	var ms runtime.MemStats
	runtime.ReadMemStats(&ms)
	now := time.Now()

	var gcPausePerSec float64
	if sh.lastPauseTotalNs > 0 {
		gcPausePerSec = time.Duration(ms.PauseTotalNs - sh.lastPauseTotalNs).Seconds()
	}
	sh.lastPauseTotalNs = ms.PauseTotalNs

	gcCount := int(ms.NumGC - sh.lastNumGC)
	var gcPerSec float64
	if sh.lastNumGC > 0 {
		gcPerSec = float64(gcCount) / now.Sub(sh.lastSampledAt).Seconds()
	}

	if gcCount > 256 {
		gcCount = 256
	}

	gcPause := make([]float64, gcCount)
	for i := 0; i < gcCount; i++ {
		gcPause[i] = time.Duration(ms.PauseNs[(int(ms.NumGC)-i+255)%256]).Seconds()
	}

	sh.lastNumGC = ms.NumGC
	sh.lastSampledAt = now

	return Stats{
		Timestamp:        now.UnixNano(),
		GoVersion:        runtime.Version(),
		GoOs:             runtime.GOOS,
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
		StackInUse:       ms.StackInuse,
		HeapAlloc:        ms.HeapAlloc,
		HeapSys:          ms.HeapSys,
		HeapIdle:         ms.HeapIdle,
		HeapInuse:        ms.HeapInuse,
		HeapReleased:     ms.HeapReleased,
		HeapObjects:      ms.HeapObjects,
		GcNext:           ms.NextGC,
		GcLast:           ms.LastGC,
		GcNum:            ms.NumGC,
		GcPerSecond:      gcPerSec,
		GcPausePerSecond: gcPausePerSec,
		GcPause:          gcPause,
	}
}
