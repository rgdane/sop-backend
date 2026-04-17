package helper

import (
	"time"

	vegeta "github.com/tsenart/vegeta/v12/lib"
)

// Pindahkan struct ke sini
type BenchmarkResult struct {
	TestName    string  `json:"test_name"`
	Target      string  `json:"target"`
	Requests    uint64  `json:"requests"`
	SuccessRate float64 `json:"success_rate"`
	MeanLatency string  `json:"mean_latency"`
	P99Latency  string  `json:"p99_latency"`
}

// Helper mandiri yang tidak bergantung pada Fiber Ctx
func RunVegetaLoadTest(method, targetURL, testName string, durationSec, ratePerSec int) BenchmarkResult {
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

	return BenchmarkResult{
		TestName:    testName,
		Target:      targetURL,
		Requests:    metrics.Requests,
		SuccessRate: metrics.Success * 100,
		MeanLatency: metrics.Latencies.Mean.String(),
		P99Latency:  metrics.Latencies.P99.String(),
	}
}
