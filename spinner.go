package spinner

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"time"
	"unicode"

	"golang.org/x/term"
)

var elements = []string{
	"૮(｡◕‿◕｡)っ",
	"૮(｡◕‿◕｡)つ",
	"⊂(｡◕‿◕｡)っ",
	"⊂(｡◕‿◕｡)つ",
}

// Spinner animates a progress indicator on stderr.
type Spinner struct {
	done    chan struct{}
	stopped chan struct{}
	once    sync.Once
}

// Start prints label and animates in place on stderr when stderr is a TTY.
// On non-TTY output, such as pipes or log files, it prints the label as a line.
// Call Stop when the operation finishes.
func Start(label string) *Spinner {
	label = cleanLabel(label)

	s := &Spinner{
		done:    make(chan struct{}),
		stopped: make(chan struct{}),
	}

	if !term.IsTerminal(int(os.Stderr.Fd())) {
		fmt.Fprintln(os.Stderr, label)
		close(s.stopped)
		return s
	}

	go func() {
		defer close(s.stopped)

		t := time.NewTicker(250 * time.Millisecond)
		defer t.Stop()

		i := 0
		printFrame := func() {
			fmt.Fprintf(os.Stderr, "\r\033[2K%s %s", elements[i%len(elements)], label)
			i++
		}

		printFrame()

		for {
			select {
			case <-s.done:
				fmt.Fprintf(os.Stderr, "\r\033[2K%s\n", label)
				return
			case <-t.C:
				printFrame()
			}
		}
	}()

	return s
}

// Stop halts the spinner. On TTY output, the spinner line is erased.
func (s *Spinner) Stop() {
	s.once.Do(func() {
		close(s.done)
	})
	<-s.stopped
}

func cleanLabel(s string) string {
	return strings.Map(func(r rune) rune {
		switch r {
		case '\n', '\r', '\t':
			return ' '
		}

		if unicode.IsControl(r) {
			return -1
		}

		return r
	}, s)
}
