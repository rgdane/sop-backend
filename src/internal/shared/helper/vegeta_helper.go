package helper

import (
	"sync"
	"time"

	vegeta "github.com/tsenart/vegeta/v12/lib"
)

type BenchmarkResult struct {
	TestName          string  `json:"test_name"`
	Target            string  `json:"target"`
	Requests          uint64  `json:"requests"`
	SuccessRate       float64 `json:"success_rate"`
	VegetaMeanLatency string  `json:"vegeta_mean_latency"`
	VegetaP99Latency  string  `json:"vegeta_p99_latency"`
	DBMeanLatencyMs   float64 `json:"db_mean_latency_ms"`
}

var (
	IsBenchmarkingActive bool
	TotalDBDuration      time.Duration
	DBRequestCount       int64
	DBMutex              sync.Mutex
)

func RecordDBLatency(duration time.Duration) {
	DBMutex.Lock()
	defer DBMutex.Unlock()
	if IsBenchmarkingActive {
		TotalDBDuration += duration
		DBRequestCount++
	}
}

func RunVegetaLoadTest(method, targetURL, testName string, durationSec, ratePerSec int) BenchmarkResult {
	DBMutex.Lock()
	IsBenchmarkingActive = true
	TotalDBDuration = 0
	DBRequestCount = 0
	DBMutex.Unlock()

	targeter := vegeta.NewStaticTargeter(vegeta.Target{
		Method: method,
		URL:    targetURL,
	})

	rate := vegeta.Rate{Freq: ratePerSec, Per: time.Second}
	duration := time.Duration(durationSec) * time.Second

	attacker := vegeta.NewAttacker()
	var metrics vegeta.Metrics

	for res := range attacker.Attack(targeter, rate, duration, testName) {
		metrics.Add(res)
	}
	metrics.Close()

	DBMutex.Lock()
	IsBenchmarkingActive = false
	var dbMeanMs float64 = 0
	if DBRequestCount > 0 {
		dbMeanMs = float64(TotalDBDuration.Milliseconds()) / float64(DBRequestCount)
	}
	DBMutex.Unlock()

	return BenchmarkResult{
		TestName:          testName,
		Target:            targetURL,
		Requests:          metrics.Requests,
		SuccessRate:       metrics.Success * 100,
		VegetaMeanLatency: metrics.Latencies.Mean.String(),
		VegetaP99Latency:  metrics.Latencies.P99.String(),
		DBMeanLatencyMs:   dbMeanMs,
	}
}
