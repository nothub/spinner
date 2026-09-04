package main

import (
	"fmt"
	"time"

	"github.com/nothub/spinner"
)

func main() {
	spinner.Spin("Working", func() { work() },
		spinner.WithStartDelay(500*time.Millisecond),
	)

	c := -1
	sp := spinner.Start("Spinning 0/20",
		spinner.WithFrames([]string{"🌑", "🌘", "🌗", "🌖", "🌕", "🌔", "🌓", "🌒"}),
		spinner.WithStartDelay(0),               // 0 animates instantly, default 3s
		spinner.WithDelay(200*time.Millisecond), // frame interval, default 250ms
		spinner.WithLabelFunc(func() string {
			c = c + 1
			// regenerates the label every frame
			return fmt.Sprintf("Spinning %v/20", c)
		}))
	defer sp.Stop()
	work()
}

func work() {
	time.Sleep(4 * time.Second)
}
