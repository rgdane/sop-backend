package helper

import (
	"time"

	vegeta "github.com/tsenart/vegeta/v12/lib"
)

// Tambahkan field untuk metrik Database murni
type BenchmarkResult struct {
	TestName          string  `json:"test_name"`
	Target            string  `json:"target"`
	Requests          uint64  `json:"requests"`
	SuccessRate       float64 `json:"success_rate"`
	VegetaMeanLatency string  `json:"vegeta_mean_latency"` // Total API Latency
	VegetaP99Latency  string  `json:"vegeta_p99_latency"`  // Total API P99
	DBMeanLatencyMs   float64 `json:"db_mean_latency_ms"`  // Waktu Murni DB (dalam Milidetik)
}

// Tambahkan parameter `dbQueryFunc` berupa fungsi callback
func RunVegetaLoadTest(method, targetURL, testName string, durationSec, ratePerSec int, dbQueryFunc func() time.Duration) BenchmarkResult {
	
	// ==========================================
	// 1. PENGUJIAN API LATENCY (DENGAN VEGETA)
	// ==========================================
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

	// ==========================================
	// 2. PENGUJIAN MURNI DATABASE LATENCY
	// ==========================================
	var dbMeanMs float64 = 0
	if dbQueryFunc != nil {
		var totalDBTime time.Duration
		dbIterations := 100 // Kita test query DB 100x beruntun untuk akurasi rata-rata
		
		for i := 0; i < dbIterations; i++ {
			totalDBTime += dbQueryFunc() // Eksekusi fungsi DB dari Handler
		}
		
		// Konversi hasil akumulasi ke milidetik (ms)
		dbMeanMs = float64(totalDBTime.Microseconds()) / float64(dbIterations) / 1000.0
	}

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