package main

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"runtime"
	"sort"
	"sync"
	"time"

	"chorus/internal/gitstore"
	"chorus/internal/thread"
)

type Result struct {
	Duration time.Duration
	Err      error
}

func runLoadScenario(concurrency int, totalRequests int) {
	fmt.Printf("\n=======================================================\n")
	fmt.Printf(" Running Load Scenario: %d Virtual Concurrent Users (%d Total Req)\n", concurrency, totalRequests)
	fmt.Printf("=======================================================\n")

	tempDir, err := os.MkdirTemp("", "chorus_loadtest_*")
	if err != nil {
		fmt.Printf("Error creating temp dir: %v\n", err)
		return
	}
	defer os.RemoveAll(tempDir)

	store := gitstore.NewGitStore(tempDir)
	repo, err := gitstore.NewThreadRepository(store)
	if err != nil {
		fmt.Printf("Error creating repo: %v\n", err)
		return
	}

	ctx := context.Background()
	now := time.Now().UTC()

	// Seed 10 base threads for reading
	seedIDs := make([]string, 10)
	for i := 0; i < 10; i++ {
		id := fmt.Sprintf("thd_seed_%d", i)
		seedIDs[i] = id
		_ = repo.SaveThread(ctx, &thread.Thread{
			ID:        id,
			Title:     fmt.Sprintf("Load Seed Thread %d", i),
			AuthorID:  "usr_seed",
			CreatedAt: now,
			UpdatedAt: now,
		})
	}

	var mStart, mEnd runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&mStart)

	workChan := make(chan int, totalRequests)
	for i := 0; i < totalRequests; i++ {
		workChan <- i
	}
	close(workChan)

	resultsChan := make(chan Result, totalRequests)
	var wg sync.WaitGroup

	startTime := time.Now()

	for w := 0; w < concurrency; w++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for reqIdx := range workChan {
				reqStart := time.Now()

				// Workload Distribution: 80% Reads, 20% Writes (Realistic Production Ratio)
				opType := reqIdx % 10
				var reqErr error

				switch {
				case opType < 4:
					// 40% ListThreads (index read)
					_, reqErr = repo.ListThreads(ctx)
				case opType < 8:
					// 40% GetThread (single thread read)
					targetID := seedIDs[rand.Intn(len(seedIDs))]
					_, reqErr = repo.FindThreadByID(ctx, targetID)
				case opType == 8:
					// 10% CreateThread (write + git commit)
					tID := fmt.Sprintf("thd_w%d_r%d", workerID, reqIdx)
					reqErr = repo.SaveThread(ctx, &thread.Thread{
						ID:        tID,
						Title:     fmt.Sprintf("Concurrent Thread %s", tID),
						AuthorID:  "usr_worker",
						CreatedAt: now,
						UpdatedAt: now,
					})
				case opType == 9:
					// 10% AppendMessage (write + git commit)
					targetID := seedIDs[rand.Intn(len(seedIDs))]
					mID := fmt.Sprintf("msg_w%d_r%d", workerID, reqIdx)
					reqErr = repo.SaveMessage(ctx, &thread.Message{
						ID:        mID,
						ThreadID:  targetID,
						AuthorID:  "usr_worker",
						Content:   "Concurrent load message body",
						CreatedAt: now,
					})
				}

				dur := time.Since(reqStart)
				resultsChan <- Result{Duration: dur, Err: reqErr}
			}
		}(w)
	}

	wg.Wait()
	totalDuration := time.Since(startTime)
	close(resultsChan)

	runtime.ReadMemStats(&mEnd)

	var durations []time.Duration
	errorCount := 0

	for res := range resultsChan {
		if res.Err != nil {
			errorCount++
		} else {
			durations = append(durations, res.Duration)
		}
	}

	if len(durations) == 0 {
		fmt.Printf("All requests failed! Error count: %d\n", errorCount)
		return
	}

	sort.Slice(durations, func(i, j int) bool {
		return durations[i] < durations[j]
	})

	p50 := durations[len(durations)*50/100]
	p95 := durations[len(durations)*95/100]
	p99 := durations[len(durations)*99/100]
	maxLat := durations[len(durations)-1]
	rps := float64(totalRequests) / totalDuration.Seconds()
	memAllocatedMB := float64(mEnd.TotalAlloc-mStart.TotalAlloc) / 1024 / 1024

	fmt.Printf("Total Elapsed Time   : %v\n", totalDuration)
	fmt.Printf("Completed Requests   : %d (Errors: %d)\n", len(durations), errorCount)
	fmt.Printf("Throughput (RPS)     : %.2f req/sec\n", rps)
	fmt.Printf("Latency p50 (Median) : %v\n", p50)
	fmt.Printf("Latency p95          : %v\n", p95)
	fmt.Printf("Latency p99          : %v\n", p99)
	fmt.Printf("Latency Max          : %v\n", maxLat)
	fmt.Printf("Memory Allocated     : %.2f MB\n", memAllocatedMB)
	fmt.Printf("Heap Objects         : %d allocs\n", mEnd.Mallocs-mStart.Mallocs)
}

func main() {
	rand.Seed(time.Now().UnixNano())
	fmt.Println("Starting Chorus Performance & Stress Testing Suite...")
	fmt.Printf("CPU Core Count       : %d\n", runtime.NumCPU())
	fmt.Printf("Go Version           : %s\n", runtime.Version())

	// 1. 100 Virtual Users
	runLoadScenario(100, 1000)

	// 2. 1,000 Virtual Users
	runLoadScenario(1000, 2000)

	// 3. 10,000 Virtual Users
	runLoadScenario(10000, 5000)
}
