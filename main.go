package main

import (
	"flag"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/codahale/hdrhistogram"
)

const timeout = 10 * time.Second

func main() {
	start := time.Now()
	concurrency := flag.Int("c", 10, "concurrency")
	number := flag.Int("n", 50, "number of requests")
	flag.Parse()

	url := flag.Arg(0)
	fmt.Printf("Site: %s\nConcurrency: %d\nNumber of requests: %d\n", url, *concurrency, *number)

	var client = http.Client{
		Timeout: timeout,
	}
	var results = make(chan time.Duration, *number)
	semaphore := make(chan struct{}, *concurrency)
	var wg sync.WaitGroup
	for i := 0; i < *number; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			semaphore <- struct{}{}
			start := time.Now()
			resp, err := client.Get(url)
			if err == nil && resp.StatusCode == 200 {
				results <- time.Since(start)
			} else {
				results <- 0
			}
			<-semaphore
		}()
	}
	wg.Wait()
	close(results)

	var errors int
	hist := hdrhistogram.New(1, int64(timeout)/1000, 3)

	// Run this separately to not introduce locks for histogram and errors counter.
	for duration := range results {
		if duration != 0 {
			hist.RecordValue(duration.Nanoseconds() / 1000000) // record milliseconds
		} else {
			errors++
		}
	}

	fmt.Printf("Latency mean: %dms\n", int(hist.Mean()))
	fmt.Printf("Latency 50%%ile: %dms\n", hist.ValueAtQuantile(50))
	fmt.Printf("Latency 90%%ile: %dms\n", hist.ValueAtQuantile(90))
	fmt.Printf("Latency 99%%ile: %dms\n", hist.ValueAtQuantile(99))
	fmt.Printf("Latency 99.99%%ile: %dms\n", hist.ValueAtQuantile(99.99))
	fmt.Printf("Errors: %d\n", errors)
	fmt.Printf("Total time: %dms\n", int(time.Since(start)/time.Millisecond))
}
