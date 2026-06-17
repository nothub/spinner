package main

import (
	"time"

	"codeberg.org/fhuebner/spinner"
)

func main() {
	spinner.Spin("Spinning...", func() {
		work()
	})
}

func work() {
	time.Sleep(3 * time.Second)
}
