package main

import (
	"flag"
	"fmt"
	"github.com/yosssi/gb/context"
	"github.com/yosssi/gb/options"
	"github.com/yosssi/gb/result"
	"net/http"
	"os"
	"runtime"
	"sync"
	"time"
)

const (
	expectedArgsLen = 1
)

var (
	ctx context.Context
)

// init executes initializing processes.
func init() {
	parseArgs()
	maxProcs()
}

// parseArgs parses args and generates a context.
func parseArgs() {
	requests := flag.Int("n", 1, "Number of requests to perform")
	concurrency := flag.Int("c", 1, "Number of multiple requests to make")
	help := flag.Bool("h", false, "Display usage information (this message)")
	debug := flag.Bool("d", false, "Run a command in debug mode")
	flag.Usage = usage
	flag.Parse()
	if *help || len(flag.Args()) != expectedArgsLen {
		flag.Usage()
		os.Exit(0)
	}
	if *requests < *concurrency {
		fmt.Println("Cannot use concurrency level greater than total number of requests")
		flag.Usage()
		os.Exit(0)
	}
	ctx = context.Context{Options: options.Options{Requests: *requests, Concurrency: *concurrency}, Url: flag.Args()[0], Debug: *debug}
	ctx.Dprintf("parse done. ctx: %+v", ctx)
}

// maxProcs sets the maximum number of CPUs to the number of logical CPUs on the local machine.
func maxProcs() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	ctx.Dprintf("maxProcs done. The maximum number of CPUs: %d", runtime.GOMAXPROCS(0))
}

func main() {
	requestC := generateClients()
	requests(requestC)
	printResult()
}

func usage() {
	fmt.Println("Usage: gb [OPTIONS] URL")
	fmt.Println("Options:")
	flag.PrintDefaults()
}

func generateClients() chan<- *sync.WaitGroup {
	var wg sync.WaitGroup
	wg.Add(ctx.Options.Concurrency)

	requestC := make(chan *sync.WaitGroup)

	for i := 0; i < ctx.Options.Concurrency; i++ {
		go client(i+1, requestC, &wg)
	}

	wg.Wait()

	return requestC
}

func client(id int, requestC <-chan *sync.WaitGroup, wg *sync.WaitGroup) {
	ctx.Dprintf("[Client %d] Starts.", id)
	wg.Done()
	for {
		reqWg := <-requestC
		ctx.Dprintf("[Client %d] Request starts.", id)
		r := result.Result{StartT: time.Now()}
		res, err := http.Get(ctx.Url)
		if err != nil {
			r.Error = err
		} else {
			r.HTTPStatusCode = res.StatusCode
		}
		r.EndT = time.Now()
		ctx.AppendResult(r)
		ctx.Dprintf("[Client %d] Request ends. [Result: %+v][Time: %d]", id, r, r.Millisecond())
		reqWg.Done()
	}
}

func requests(requestC chan<- *sync.WaitGroup) {
	var reqWg sync.WaitGroup
	reqWg.Add(ctx.Options.Requests)

	for i := 0; i < ctx.Options.Requests; i++ {
		go request(i+1, requestC, &reqWg)
	}

	reqWg.Wait()
}

func request(id int, requestC chan<- *sync.WaitGroup, reqWg *sync.WaitGroup) {
	ctx.Dprintf("[Request %d] Starts.", id)
	requestC <- reqWg
}

func printResult() {
	successful := 0
	failed := 0
	totalTimeMillisecond := 0
	var maxTime, minTime int
	maxMinTimeSet := false
	for _, r := range ctx.Results {
		if r.HTTPStatusCode == http.StatusOK {
			successful++
			t := r.Millisecond()
			totalTimeMillisecond += t
			if !maxMinTimeSet {
				maxTime = t
				minTime = t
				maxMinTimeSet = true
			} else {
				if maxTime < t {
					maxTime = t
				}
				if minTime > t {
					minTime = t
				}
			}
		} else {
			failed++
		}
	}
	fmt.Printf("Successful: %d\nFailed: %d\nTotal Time: %.3f sec\nAverage Time: %d ms\nMax Time: %d ms\nMin Time: %d ms\n", successful, failed, float64(totalTimeMillisecond)/1000, totalTimeMillisecond/ctx.Options.Requests, maxTime, minTime)
}
