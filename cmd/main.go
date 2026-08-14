package main

import (
	"time"

	"github.com/nothub/spinner"
)

func main() {
	spinner.Spin("Spinning...", func() { work() },
		spinner.WithStartDelay(3*time.Second),
		spinner.WithDelay(300*time.Millisecond),
	)
}

func work() {
	time.Sleep(10 * time.Second)
}
