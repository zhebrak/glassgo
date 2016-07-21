package main

import (
	"flag"
	"fmt"
	"gopkg.in/cheggaaa/pb.v1"
	"net/http"
	"time"
)

type requestResult struct {
	duration time.Duration
	err      error
}

func startPB(number int) *pb.ProgressBar {
	progress := pb.New(number)

	progress.ShowBar = false
	progress.ShowTimeLeft = false
	progress.ShowFinalTime = false
	progress.Start()

	return progress
}

func finishPB(progress *pb.ProgressBar) {
	progress.ShowPercent = false
	progress.ShowCounters = false
	progress.Finish()
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

	var c = make(chan *requestResult, *concurrency)
	var queue = make(chan int, *number)

	for i := 0; i < *concurrency; i++ {
		go func() {
			for _ = range queue {
				c <- ping(url, *timeout)
			}
		}()
	}

	for i := 0; i < *number; i++ {
		queue <- i
	}

	var errors int
	durations := make([]time.Duration, 0, *number)

	progress := startPB(*number)
	for res := range c {
		progress.Increment()

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

			finishPB(progress)

			fmt.Printf("Site: %s\n", url)
			fmt.Printf("Number: %d\n", *number)
			fmt.Printf("Concurrency: %d\n", *concurrency)
			fmt.Printf("Avg: %dms\n", int(avg))
			fmt.Printf("Max: %dms\n", int(max))
			fmt.Printf("Total: %dms\n", int(time.Since(start)/time.Millisecond))
			fmt.Printf("Errors: %d\n\n", errors)

			return
		}
	}
}
