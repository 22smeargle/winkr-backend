package testutils

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// PerformanceTestHelper provides utilities for performance testing
type PerformanceTestHelper struct {
	t            *testing.T
	metrics      *PerformanceMetrics
	startTime    time.Time
	endTime      time.Time
	isRunning    bool
	mu           sync.Mutex
}

// PerformanceMetrics holds performance metrics
type PerformanceMetrics struct {
	TotalRequests     int64
	SuccessfulReqs   int64
	FailedReqs       int64
	TotalDuration    time.Duration
	MinDuration      time.Duration
	MaxDuration      time.Duration
	MemoryUsage      runtime.MemStats
	CPUUsage        float64
	ConcurrentUsers  int64
	ThroughputQPS    float64
	ErrorRate        float64
	AvgResponseTime  time.Duration
	P95ResponseTime  time.Duration
	P99ResponseTime  time.Duration
}

// NewPerformanceTestHelper creates a new performance test helper
func NewPerformanceTestHelper(t *testing.T) *PerformanceTestHelper {
	return &PerformanceTestHelper{
		t:       t,
		metrics: &PerformanceMetrics{},
	}
}

// StartTest starts the performance test
func (pth *PerformanceTestHelper) StartTest() {
	pth.mu.Lock()
	defer pth.mu.Unlock()
	
	if pth.isRunning {
		return
	}
	
	pth.isRunning = true
	pth.startTime = time.Now()
	pth.metrics = &PerformanceMetrics{
		MinDuration: time.Hour, // Initialize to a large value
	}
	
	// Start memory monitoring
	runtime.ReadMemStats(&pth.metrics.MemoryUsage)
}

// StopTest stops the performance test and calculates metrics
func (pth *PerformanceTestHelper) StopTest() {
	pth.mu.Lock()
	defer pth.mu.Unlock()
	
	if !pth.isRunning {
		return
	}
	
	pth.endTime = time.Now()
	pth.isRunning = false
	
	// Calculate final metrics
	pth.calculateMetrics()
}

// RecordRequest records a single request
func (pth *PerformanceTestHelper) RecordRequest(duration time.Duration, success bool) {
	pth.mu.Lock()
	defer pth.mu.Unlock()
	
	if !pth.isRunning {
		return
	}
	
	atomic.AddInt64(&pth.metrics.TotalRequests, 1)
	atomic.AddInt64(&pth.metrics.TotalDuration, int64(duration))
	
	if success {
		atomic.AddInt64(&pth.metrics.SuccessfulReqs, 1)
	} else {
		atomic.AddInt64(&pth.metrics.FailedReqs, 1)
	}
	
	// Update min/max duration
	if duration < pth.metrics.MinDuration {
		pth.metrics.MinDuration = duration
	}
	if duration > pth.metrics.MaxDuration {
		pth.metrics.MaxDuration = duration
	}
}

// RecordConcurrentUser records a concurrent user
func (pth *PerformanceTestHelper) RecordConcurrentUser() {
	atomic.AddInt64(&pth.metrics.ConcurrentUsers, 1)
}

// RemoveConcurrentUser removes a concurrent user
func (pth *PerformanceTestHelper) RemoveConcurrentUser() {
	atomic.AddInt64(&pth.metrics.ConcurrentUsers, -1)
}

// calculateMetrics calculates final performance metrics
func (pth *PerformanceTestHelper) calculateMetrics() {
	totalReqs := atomic.LoadInt64(&pth.metrics.TotalRequests)
	successfulReqs := atomic.LoadInt64(&pth.metrics.SuccessfulReqs)
	failedReqs := atomic.LoadInt64(&pth.metrics.FailedReqs)
	totalDuration := atomic.LoadInt64(&pth.metrics.TotalDuration)
	
	if totalReqs > 0 {
		testDuration := pth.endTime.Sub(pth.startTime)
		
		pth.metrics.ThroughputQPS = float64(totalReqs) / testDuration.Seconds()
		pth.metrics.ErrorRate = float64(failedReqs) / float64(totalReqs) * 100
		pth.metrics.AvgResponseTime = time.Duration(totalDuration) / time.Duration(totalReqs)
	}
	
	// Read final memory stats
	runtime.ReadMemStats(&pth.metrics.MemoryUsage)
}

// GetMetrics returns the performance metrics
func (pth *PerformanceTestHelper) GetMetrics() *PerformanceMetrics {
	pth.mu.Lock()
	defer pth.mu.Unlock()
	
	return pth.metrics
}

// AssertThroughput asserts that throughput meets minimum requirements
func (pth *PerformanceTestHelper) AssertThroughput(minQPS float64) {
	metrics := pth.GetMetrics()
	require.GreaterOrEqual(pth.t, metrics.ThroughputQPS, minQPS, 
		"Throughput %.2f QPS is below minimum %.2f QPS", metrics.ThroughputQPS, minQPS)
}

// AssertErrorRate asserts that error rate is below maximum threshold
func (pth *PerformanceTestHelper) AssertErrorRate(maxErrorRate float64) {
	metrics := pth.GetMetrics()
	require.LessOrEqual(pth.t, metrics.ErrorRate, maxErrorRate, 
		"Error rate %.2f%% is above maximum %.2f%%", metrics.ErrorRate, maxErrorRate)
}

// AssertAvgResponseTime asserts that average response time is below maximum
func (pth *PerformanceTestHelper) AssertAvgResponseTime(maxTime time.Duration) {
	metrics := pth.GetMetrics()
	require.LessOrEqual(pth.t, metrics.AvgResponseTime, maxTime, 
		"Average response time %v is above maximum %v", metrics.AvgResponseTime, maxTime)
}

// AssertMaxResponseTime asserts that maximum response time is below maximum
func (pth *PerformanceTestHelper) AssertMaxResponseTime(maxTime time.Duration) {
	metrics := pth.GetMetrics()
	require.LessOrEqual(pth.t, metrics.MaxDuration, maxTime, 
		"Maximum response time %v is above maximum %v", metrics.MaxDuration, maxTime)
}

// AssertMemoryUsage asserts that memory usage is below maximum
func (pth *PerformanceTestHelper) AssertMemoryUsage(maxMB uint64) {
	metrics := pth.GetMetrics()
	memoryMB := metrics.MemoryUsage.Alloc / 1024 / 1024
	require.LessOrEqual(pth.t, memoryMB, maxMB, 
		"Memory usage %d MB is above maximum %d MB", memoryMB, maxMB)
}

// LoadTestRunner runs load tests with different configurations
type LoadTestRunner struct {
	t           *testing.T
	helper      *PerformanceTestHelper
	concurrency int
	duration    time.Duration
	rampUp      time.Duration
}

// NewLoadTestRunner creates a new load test runner
func NewLoadTestRunner(t *testing.T, concurrency int, duration time.Duration) *LoadTestRunner {
	return &LoadTestRunner{
		t:           t,
		helper:      NewPerformanceTestHelper(t),
		concurrency: concurrency,
		duration:    duration,
		rampUp:      time.Second * 5, // Default 5 second ramp-up
	}
}

// WithRampUp sets the ramp-up duration
func (ltr *LoadTestRunner) WithRampUp(rampUp time.Duration) *LoadTestRunner {
	ltr.rampUp = rampUp
	return ltr
}

// Run runs the load test with the provided function
func (ltr *LoadTestRunner) Run(testFunc func() (time.Duration, bool)) {
	ltr.helper.StartTest()
	defer ltr.helper.StopTest()
	
	var wg sync.WaitGroup
	ctx, cancel := context.WithTimeout(context.Background(), ltr.duration)
	defer cancel()
	
	// Calculate ramp-up delay between goroutines
	rampDelay := ltr.rampUp / time.Duration(ltr.concurrency)
	
	// Start concurrent workers
	for i := 0; i < ltr.concurrency; i++ {
		wg.Add(1)
		
		go func(workerID int) {
			defer wg.Done()
			
			// Ramp-up delay
			time.Sleep(time.Duration(workerID) * rampDelay)
			
			ltr.helper.RecordConcurrentUser()
			defer ltr.helper.RemoveConcurrentUser()
			
			for {
				select {
				case <-ctx.Done():
					return
				default:
					duration, success := testFunc()
					ltr.helper.RecordRequest(duration, success)
				}
			}
		}(i)
	}
	
	wg.Wait()
}

// GetMetrics returns the load test metrics
func (ltr *LoadTestRunner) GetMetrics() *PerformanceMetrics {
	return ltr.helper.GetMetrics()
}

// StressTestRunner runs stress tests
type StressTestRunner struct {
	t           *testing.T
	helper      *PerformanceTestHelper
	maxUsers    int
	stepSize    int
	stepTime    time.Duration
	maxDuration time.Duration
}

// NewStressTestRunner creates a new stress test runner
func NewStressTestRunner(t *testing.T, maxUsers int, stepSize int, stepTime time.Duration) *StressTestRunner {
	return &StressTestRunner{
		t:           t,
		helper:      NewPerformanceTestHelper(t),
		maxUsers:    maxUsers,
		stepSize:    stepSize,
		stepTime:    stepTime,
		maxDuration: time.Hour, // Default 1 hour max
	}
}

// WithMaxDuration sets the maximum test duration
func (str *StressTestRunner) WithMaxDuration(duration time.Duration) *StressTestRunner {
	str.maxDuration = duration
	return str
}

// Run runs the stress test
func (str *StressTestRunner) Run(testFunc func() (time.Duration, bool)) {
	str.helper.StartTest()
	defer str.helper.StopTest()
	
	ctx, cancel := context.WithTimeout(context.Background(), str.maxDuration)
	defer cancel()
	
	currentUsers := 0
	stepCtx, stepCancel := context.WithCancel(context.Background())
	
	for currentUsers < str.maxUsers {
		select {
		case <-ctx.Done():
			return
		default:
			// Add more users
			nextUsers := currentUsers + str.stepSize
			if nextUsers > str.maxUsers {
				nextUsers = str.maxUsers
			}
			
			str.runStep(stepCtx, nextUsers-currentUsers, testFunc)
			currentUsers = nextUsers
			
			// Wait for step time
			select {
			case <-ctx.Done():
				return
			case <-time.After(str.stepTime):
				// Continue to next step
			}
		}
	}
	
	stepCancel()
}

// runStep runs a single step of the stress test
func (str *StressTestRunner) runStep(ctx context.Context, newUsers int, testFunc func() (time.Duration, bool)) {
	var wg sync.WaitGroup
	
	for i := 0; i < newUsers; i++ {
		wg.Add(1)
		
		go func() {
			defer wg.Done()
			
			str.helper.RecordConcurrentUser()
			defer str.helper.RemoveConcurrentUser()
			
			for {
				select {
				case <-ctx.Done():
					return
				default:
					duration, success := testFunc()
					str.helper.RecordRequest(duration, success)
				}
			}
		}()
	}
	
	wg.Wait()
}

// GetMetrics returns the stress test metrics
func (str *StressTestRunner) GetMetrics() *PerformanceMetrics {
	return str.helper.GetMetrics()
}

// BenchmarkHelper provides utilities for benchmark testing
type BenchmarkHelper struct {
	t         *testing.T
	iterations int
	warmup    int
}

// NewBenchmarkHelper creates a new benchmark helper
func NewBenchmarkHelper(t *testing.T, iterations int) *BenchmarkHelper {
	return &BenchmarkHelper{
		t:         t,
		iterations: iterations,
		warmup:    iterations / 10, // 10% warmup by default
	}
}

// WithWarmup sets the number of warmup iterations
func (bh *BenchmarkHelper) WithWarmup(warmup int) *BenchmarkHelper {
	bh.warmup = warmup
	return bh
}

// Run runs the benchmark
func (bh *BenchmarkHelper) Run(testFunc func() time.Duration) BenchmarkResult {
	// Warmup
	for i := 0; i < bh.warmup; i++ {
		testFunc()
	}
	
	// Actual benchmark
	durations := make([]time.Duration, bh.iterations)
	
	for i := 0; i < bh.iterations; i++ {
		durations[i] = testFunc()
	}
	
	return bh.calculateResults(durations)
}

// calculateResults calculates benchmark results
func (bh *BenchmarkHelper) calculateResults(durations []time.Duration) BenchmarkResult {
	if len(durations) == 0 {
		return BenchmarkResult{}
	}
	
	// Sort durations for percentile calculations
	sorted := make([]time.Duration, len(durations))
	copy(sorted, durations)
	
	// Simple bubble sort for small datasets
	for i := 0; i < len(sorted); i++ {
		for j := 0; j < len(sorted)-1-i; j++ {
			if sorted[j] > sorted[j+1] {
				sorted[j], sorted[j+1] = sorted[j+1], sorted[j]
			}
		}
	}
	
	var total time.Duration
	min := sorted[0]
	max := sorted[len(sorted)-1]
	
	for _, d := range sorted {
		total += d
	}
	
	avg := total / time.Duration(len(sorted))
	
	p50 := sorted[len(sorted)/2]
	p95 := sorted[int(float64(len(sorted))*0.95)]
	p99 := sorted[int(float64(len(sorted))*0.99)]
	
	return BenchmarkResult{
		Iterations: len(durations),
		Min:        min,
		Max:        max,
		Avg:        avg,
		P50:        p50,
		P95:        p95,
		P99:        p99,
		Total:      total,
		Throughput: float64(len(durations)) / total.Seconds(),
	}
}

// BenchmarkResult holds benchmark results
type BenchmarkResult struct {
	Iterations int
	Min        time.Duration
	Max        time.Duration
	Avg        time.Duration
	P50        time.Duration
	P95        time.Duration
	P99        time.Duration
	Total      time.Duration
	Throughput float64
}

// AssertThroughput asserts that throughput meets minimum requirements
func (br *BenchmarkResult) AssertThroughput(t *testing.T, minThroughput float64) {
	require.GreaterOrEqual(t, br.Throughput, minThroughput, 
		"Throughput %.2f ops/sec is below minimum %.2f ops/sec", br.Throughput, minThroughput)
}

// AssertAvgTime asserts that average time is below maximum
func (br *BenchmarkResult) AssertAvgTime(t *testing.T, maxTime time.Duration) {
	require.LessOrEqual(t, br.Avg, maxTime, 
		"Average time %v is above maximum %v", br.Avg, maxTime)
}

// AssertP95Time asserts that P95 time is below maximum
func (br *BenchmarkResult) AssertP95Time(t *testing.T, maxTime time.Duration) {
	require.LessOrEqual(t, br.P95, maxTime, 
		"P95 time %v is above maximum %v", br.P95, maxTime)
}

// MemoryProfiler provides memory profiling utilities
type MemoryProfiler struct {
	t      *testing.T
	before runtime.MemStats
	after  runtime.MemStats
}

// NewMemoryProfiler creates a new memory profiler
func NewMemoryProfiler(t *testing.T) *MemoryProfiler {
	return &MemoryProfiler{t: t}
}

// Start starts memory profiling
func (mp *MemoryProfiler) Start() {
	runtime.ReadMemStats(&mp.before)
}

// Stop stops memory profiling and returns the difference
func (mp *MemoryProfiler) Stop() MemoryDiff {
	runtime.ReadMemStats(&mp.after)
	
	return MemoryDiff{
		AllocDiff:      mp.after.Alloc - mp.before.Alloc,
		TotalAllocDiff: mp.after.TotalAlloc - mp.before.TotalAlloc,
		MallocsDiff:    mp.after.Mallocs - mp.before.Mallocs,
		FreesDiff:      mp.after.Frees - mp.before.Frees,
		HeapDiff:       mp.after.HeapAlloc - mp.before.HeapAlloc,
		GCDiff:         mp.after.NumGC - mp.before.NumGC,
	}
}

// MemoryDiff represents memory usage difference
type MemoryDiff struct {
	AllocDiff      uint64
	TotalAllocDiff uint64
	MallocsDiff    uint64
	FreesDiff      uint64
	HeapDiff       uint64
	GCDiff         uint32
}

// AssertNoMemoryLeak asserts that there's no significant memory leak
func (md *MemoryDiff) AssertNoMemoryLeak(t *testing.T, threshold uint64) {
	require.LessOrEqual(t, md.AllocDiff, threshold, 
		"Memory allocation increased by %d bytes, threshold is %d bytes", md.AllocDiff, threshold)
}

// AssertNoGoroutineLeak asserts that there's no goroutine leak
func AssertNoGoroutineLeak(t *testing.T, before, after int) {
	require.Equal(t, before, after, 
		"Goroutine count changed from %d to %d, possible leak detected", before, after)
}

// CPUMeasurer provides CPU measurement utilities
type CPUMeasurer struct {
	t      *testing.T
	before float64
	after  float64
}

// NewCPUMeasurer creates a new CPU measurer
func NewCPUMeasurer(t *testing.T) *CPUMeasurer {
	return &CPUMeasurer{t: t}
}

// Start starts CPU measurement
func (cm *CPUMeasurer) Start() {
	cm.before = getCPUUsage()
}

// Stop stops CPU measurement and returns the difference
func (cm *CPUMeasurer) Stop() CPUDiff {
	cm.after = getCPUUsage()
	
	return CPUDiff{
		UsageDiff: cm.after - cm.before,
	}
}

// CPUDiff represents CPU usage difference
type CPUDiff struct {
	UsageDiff float64
}

// getCPUUsage gets current CPU usage (simplified implementation)
func getCPUUsage() float64 {
	// This is a simplified implementation
	// In a real scenario, you'd use system-specific calls or libraries
	// to get accurate CPU usage
	return float64(runtime.NumGoroutine())
}

// AssertCPUUsage asserts that CPU usage is within acceptable range
func (cd *CPUDiff) AssertCPUUsage(t *testing.T, maxUsage float64) {
	require.LessOrEqual(t, cd.UsageDiff, maxUsage, 
		"CPU usage increased by %.2f, maximum allowed is %.2f", cd.UsageDiff, maxUsage)
}

// LatencyProfiler provides latency profiling utilities
type LatencyProfiler struct {
	t         *testing.T
	durations []time.Duration
	mu        sync.Mutex
}

// NewLatencyProfiler creates a new latency profiler
func NewLatencyProfiler(t *testing.T) *LatencyProfiler {
	return &LatencyProfiler{
		t:         t,
		durations: make([]time.Duration, 0),
	}
}

// Record records a latency measurement
func (lp *LatencyProfiler) Record(duration time.Duration) {
	lp.mu.Lock()
	defer lp.mu.Unlock()
	
	lp.durations = append(lp.durations, duration)
}

// GetPercentiles returns latency percentiles
func (lp *LatencyProfiler) GetPercentiles() LatencyPercentiles {
	lp.mu.Lock()
	defer lp.mu.Unlock()
	
	if len(lp.durations) == 0 {
		return LatencyPercentiles{}
	}
	
	// Sort durations
	sorted := make([]time.Duration, len(lp.durations))
	copy(sorted, lp.durations)
	
	// Simple bubble sort
	for i := 0; i < len(sorted); i++ {
		for j := 0; j < len(sorted)-1-i; j++ {
			if sorted[j] > sorted[j+1] {
				sorted[j], sorted[j+1] = sorted[j+1], sorted[j]
			}
		}
	}
	
	return LatencyPercentiles{
		P50: sorted[len(sorted)/2],
		P90: sorted[int(float64(len(sorted))*0.90)],
		P95: sorted[int(float64(len(sorted))*0.95)],
		P99: sorted[int(float64(len(sorted))*0.99)],
		P999: sorted[int(float64(len(sorted))*0.999)],
		Min:  sorted[0],
		Max:  sorted[len(sorted)-1],
	}
}

// LatencyPercentiles holds latency percentile measurements
type LatencyPercentiles struct {
	P50  time.Duration
	P90  time.Duration
	P95  time.Duration
	P99  time.Duration
	P999 time.Duration
	Min  time.Duration
	Max  time.Duration
}

// AssertP99Latency asserts that P99 latency is below maximum
func (lp *LatencyPercentiles) AssertP99Latency(t *testing.T, maxLatency time.Duration) {
	require.LessOrEqual(t, lp.P99, maxLatency, 
		"P99 latency %v is above maximum %v", lp.P99, maxLatency)
}

// AssertP95Latency asserts that P95 latency is below maximum
func (lp *LatencyPercentiles) AssertP95Latency(t *testing.T, maxLatency time.Duration) {
	require.LessOrEqual(t, lp.P95, maxLatency, 
		"P95 latency %v is above maximum %v", lp.P95, maxLatency)
}

// PrintResults prints latency results
func (lp *LatencyPercentiles) PrintResults() {
	fmt.Printf("Latency Percentiles:\n")
	fmt.Printf("  Min:  %v\n", lp.Min)
	fmt.Printf("  P50:  %v\n", lp.P50)
	fmt.Printf("  P90:  %v\n", lp.P90)
	fmt.Printf("  P95:  %v\n", lp.P95)
	fmt.Printf("  P99:  %v\n", lp.P99)
	fmt.Printf("  P999: %v\n", lp.P999)
	fmt.Printf("  Max:  %v\n", lp.Max)
}