package main

import (
	"flag"
	"fmt"
	"net/http"
	"time"
)

var (
	c = make(chan time.Duration, 10)
)

func ping(url string, number int) {
	for i := 0; i < number; i++ {
		t := time.Now()
		resp, err := http.Get(url)
		if err == nil && resp.StatusCode == 200 {
			c <- time.Since(t)
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
	fmt.Println("Site:", url, "\nConcurrency:", *concurrency, "\nNumber of requests:", *number)

	for i := 0; i < *concurrency; i++ {
		go ping(url, *number / *concurrency)
	}
	go ping(url, *number%*concurrency)

	var errors int
	durations := make([]time.Duration, 0, *number)

	for {
		select {
		case duration := <-c:
			if duration != 0 {
				durations = append(durations, duration)
			} else {
				errors++
			}

			if len(durations)+errors == *number {
				sum := 0.0
				for _, val := range durations {
					sum += float64(val / time.Millisecond)
				}

				var avg float64
				if len(durations) != 0 {
					avg = sum / float64(len(durations))
				} else {
					avg = 0
				}

				fmt.Printf("Avg: %dms Errors: %d\n", int(avg), errors)
				fmt.Printf("Total time: %dms\n", int(time.Since(start)/time.Millisecond))

				return
			}
		}
	}
}
