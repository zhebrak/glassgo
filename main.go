package main

import (
	"flag"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"gopkg.in/cheggaaa/pb.v1"
)

type requestResult struct {
	duration int // ms
	err      error
}

func startPB(number int) *pb.ProgressBar {
	pbar := pb.New(number)

	pbar.ShowBar = false
	pbar.ShowTimeLeft = false
	pbar.ShowFinalTime = false
	pbar.Start()

	return pbar
}

func finishPB(pbar *pb.ProgressBar) {
	pbar.ShowPercent = false
	pbar.ShowCounters = false
	pbar.Finish()
}

func percentile(s []int, p int) int {
	return s[len(s)*p/100-1]
}

func ping(url string, timeout int) *requestResult {
	start := time.Now()

	client := http.Client{
		Timeout: time.Duration(timeout) * time.Second,
	}
	_, err := client.Get(url)

	return &requestResult{int(time.Since(start) / time.Millisecond), err}
}

func main() {
	start := time.Now()

	concurrency := flag.Int("c", 10, "concurrency")
	number := flag.Int("n", 50, "number of requests")
	timeout := flag.Int("t", 10, "timeout per request in seconds")
	percentiles := flag.String("p", "90,99", "comma separated percentiles")

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
	durations := make([]int, 0, *number)

	pbar := startPB(*number)
	for res := range c {
		pbar.Increment()

		if res.err != nil {
			errors++
		} else {
			durations = append(durations, res.duration)
		}

		if len(durations)+errors == *number {
			var sum, avg int
			for _, val := range durations {
				sum += val
			}

			if len(durations) != 0 {
				avg = sum / len(durations)
			}

			finishPB(pbar)

			sort.Ints(durations)

			fmt.Printf("Site: %s\n", url)
			fmt.Printf("Number: %d\n", *number)
			fmt.Printf("Concurrency: %d\n\n", *concurrency)

			for _, p := range strings.Split(*percentiles, ",") {
				i, _ := strconv.Atoi(p)
				fmt.Printf("%s%%: <%dms\n", p, percentile(durations, i))
			}

			fmt.Printf("Avg: %dms\n", avg)
			fmt.Printf("Max: %dms\n\n", durations[len(durations)-1])

			fmt.Printf("Total: %dms\n", int(time.Since(start)/time.Millisecond))
			fmt.Printf("Errors: %d\n\n", errors)

			return
		}
	}
}
