package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Kqzz/MCsniperGO/claimer"
	"github.com/Kqzz/MCsniperGO/log"
	"github.com/Kqzz/MCsniperGO/pkg/parser"
)

const help = `usage:
    mcsnipergo [options]
options:
    --username, -u <str>    username to snipe
    --drop-range, -d <str>  droptime range (start-end/infinite)
    --disable-bar           disables the status bar
`

var disableBar bool

func init() {
	flag.Usage = func() {
		fmt.Print(help)
	}
}

func isFlagPassed(names ...string) bool {
	found := false
	for _, name := range names {
		flag.Visit(func(f *flag.Flag) {
			if f.Name == name {
				found = true
			}
		})
	}
	return found
}

func statusBar(startTime time.Time) {
	fmt.Print("\x1B7")     // Save the cursor position
	fmt.Print("\x1B[2K")   // Erase the entire line - breaks smth else so idk
	fmt.Print("\x1B[0J")   // Erase from cursor to end of screen
	fmt.Print("\x1B[?47h") // Save screen
	// fmt.Print("\x1B[1J")   // Erase from cursor to beginning of screen
	fmt.Print("\x1B[?47l") // Restore screen

	fmt.Printf("\x1B[%d;%dH", 0, 0) // move cursor to row #, col #

	elapsed := time.Since(startTime).Seconds()

	requestsPerSecond := float64(claimer.Stats.Total) / elapsed

	fmt.Printf("[RPS: %.2f | DUPLICATE: %d | NOT_ALLOWED: %d | TOO_MANY_REQUESTS: %d]     ", requestsPerSecond, claimer.Stats.Duplicate, claimer.Stats.NotAllowed, claimer.Stats.TooManyRequests)
	fmt.Print("\x1B8") // Restore the cursor position util new size is calculated
}

func main() {

	var startUsername string
	var startDropRange string
	flag.StringVar(&startUsername, "username", "", "username to snipe")
	flag.StringVar(&startUsername, "u", "", "username to snipe")
	flag.StringVar(&startDropRange, "drop-range", "", "droptime range (start-end/infinite)")
	flag.StringVar(&startDropRange, "d", "", "droptime range (start-end/infinite)")
	flag.BoolVar(&disableBar, "disable-bar", false, "disables status bar")
	if isFlagPassed("disable-bar") {
		disableBar = true
	}

	flag.Parse()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-c
		fmt.Print("\r")
		log.Log("err", "ctrl-c pressed, exiting...      ")
		os.Exit(0)
	}()

	for {

		log.Log("", log.GetHeader())

		// Get proxies, program continues even if this fails
		proxies, err := parser.ReadLines("proxies.txt")

		if err != nil {
			log.Log("err", "failed to load proxies: %v", err)
		}

		err = nil

		// get accounts, fails if no accs present
		accounts, err := getAccounts("gc.txt", "gp.txt", "ms.txt")

		if err != nil {
			log.Log("err", "fatal: %v", err)
			log.Input("press enter to continue")
			continue
		}

		var username string

		if !isFlagPassed("u", "username") {
			username = log.Input("target username")
		} else {
			username = startUsername
		}
		
		var drop string

		if !isFlagPassed("d", "drop-range") {
			drop = log.Input("droptime range (start-end/infinite)")
		} else {
			drop = startDropRange
		}

		dropRange := log.GetDropRange(drop)

		go func() {

			if disableBar {
				return
			}

			if dropRange.Start.After(time.Now()) {
				time.Sleep(time.Until(dropRange.Start))
			}

			start := dropRange.Start
			if start.Before(time.Now()) {
				start = time.Now()
			}

			for {
				statusBar(start)
				time.Sleep(time.Second * 1)
			}
		}()

		err = claimer.ClaimWithinRange(username, dropRange, accounts, proxies)

		if err != nil {
			log.Log("err", "fatal: %v", err)
		}

		log.Input("snipe completed, press enter to continue")
	}

}
