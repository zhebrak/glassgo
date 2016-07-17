package main

import (
	"flag"
	"fmt"
	"net/http"
	"time"
)

var c = make(chan time.Duration, 10)

func ping(url string, number int) {
	for i := 0; i < number; i++ {
		start := time.Now()
		resp, err := http.Get(url)
		if err == nil && resp.StatusCode == 200 {
			c <- time.Since(start)
		} else {
			c <- 0
		}
	}
}

func main() {
	start := time.Now()
	concurrency := flag.Int("c", 10, "concurrency")
	number := flag.Int("n", 50, "number of requests")
	flag.Parse()

	url := flag.Arg(0)
	fmt.Printf("Site: %s\nConcurrency: %d\nNumber of requests: %d\n", url, *concurrency, *number)

	for i := 0; i < *concurrency; i++ {
		go ping(url, *number / *concurrency)
	}
	go ping(url, *number%*concurrency)

	var errors int
	durations := make([]time.Duration, 0, *number)

	for {
		duration := <-c
		if duration != 0 {
			durations = append(durations, duration)
		} else {
			errors++
		}

		if len(durations)+errors == *number {
			var sum, avg float64
			for _, val := range durations {
				sum += float64(val / time.Millisecond)
			}

			if len(durations) != 0 {
				avg = sum / float64(len(durations))
			}

			fmt.Printf("Average response time: %dms Errors: %d\n", int(avg), errors)
			fmt.Printf("Total time: %dms\n", int(time.Since(start)/time.Millisecond))

			return
		}
	}
}
