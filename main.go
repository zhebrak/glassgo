package main

import (
	"flag"
	"fmt"
	"net/http"
	"time"
)

type requestResult struct {
	duration time.Duration
	err      error
}

func pinger(url string, queue chan int, c chan *requestResult, timeout int) {
	for _ = range queue {
		c <- ping(url, timeout)
	}

}

func ping(url string, timeout int) *requestResult {
	start := time.Now()

	client := http.Client{
		Timeout: time.Duration(timeout) * time.Second,
	}
	_, err := client.Get(url)

	return &requestResult{time.Since(start), err}
}

func main() {
	start := time.Now()

	concurrency := flag.Int("c", 10, "concurrency")
	number := flag.Int("n", 50, "number of requests")
	timeout := flag.Int("t", 10, "timeout per request in seconds")
	flag.Parse()
	url := flag.Arg(0)

	fmt.Printf("Site: %s\n", url)
	fmt.Printf("Number: %d\n", *number)
	fmt.Printf("Concurrency: %d\n", *concurrency)

	var c = make(chan *requestResult, *concurrency)
	var queue = make(chan int, *number)

	for i := 0; i < *concurrency; i++ {
		go pinger(url, queue, c, *timeout)
	}

	for i := 0; i < *number; i++ {
		queue <- i
	}

	var errors int
	durations := make([]time.Duration, 0, *number)

	for res := range c {
		if res.err != nil {
			errors++
		} else {
			durations = append(durations, res.duration)
		}

		if len(durations)+errors == *number {
			var sum, avg, max, tval float64
			for _, val := range durations {
				tval = float64(val / time.Millisecond)

				sum += tval
				if max < tval {
					max = tval
				}
			}

			if len(durations) != 0 {
				avg = sum / float64(len(durations))
			}

			fmt.Printf("Avg: %dms\n", int(avg))
			fmt.Printf("Max: %dms\n", int(max))
			fmt.Printf("Total: %dms\n", int(time.Since(start)/time.Millisecond))
			fmt.Printf("Errors: %d\n", errors)

			return
		}
	}
}
