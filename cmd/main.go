package main

import (
	"time"

	"codeberg.org/fhuebner/spinner"
)

func main() {
	sp := spinner.Start("Spinning...")
	defer sp.Stop()
	work()
}

func work() {
	time.Sleep(3 * time.Second)
}
