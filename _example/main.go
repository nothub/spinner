package main

import (
	"fmt"
	"sync/atomic"
	"time"

	"github.com/nothub/spinner"
)

func main() {
	spinner.Spin("Working", func() { time.Sleep(4 * time.Second) },
		spinner.WithStartDelay(500*time.Millisecond),
	)

	var done atomic.Int64
	sp := spinner.Start(fmt.Sprintf("Spinning 0/20"),
		spinner.WithFrames([]string{"🌑", "🌘", "🌗", "🌖", "🌕", "🌔", "🌓", "🌒"}),
		spinner.WithStartDelay(0),               // 0 animates instantly, default 3s
		spinner.WithDelay(200*time.Millisecond), // frame interval, default 250ms
		spinner.WithLabelFunc(func() string {
			// must be safe for concurrent use
			return fmt.Sprintf("Spinning %d/20", done.Load())
		}))
	defer sp.Stop()
	for range 20 {
		done.Add(1)
		time.Sleep(200 * time.Millisecond)
	}
}
