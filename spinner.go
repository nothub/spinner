// Package spinner spins a small progress spinner on stderr while a long-running operation is running.
//
//	sp := spinner.Start("Spinning...")
//	defer sp.Stop()
//	work()
package spinner

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"time"
	"unicode"

	"github.com/mattn/go-runewidth"
	"golang.org/x/term"
)

// Option configures a Spinner.
type Option func(*Spinner)

// WithStartDelay sets how long to wait before the spinner appears.
func WithStartDelay(d time.Duration) Option {
	return func(s *Spinner) { s.startDelay = d }
}

// WithDelay sets the animation frame interval.
func WithDelay(d time.Duration) Option {
	return func(s *Spinner) {
		if d > 0 {
			s.delay = d
		}
	}
}

// WithFrames sets the animation frames.
func WithFrames(frames []string) Option {
	return func(s *Spinner) {
		if len(frames) > 0 {
			s.frames = frames
		}
	}
}

// WithLabelFunc sets a function called on every animation tick to regenerate the label.
// The label passed to Start is still used for non-TTY output and the frame shown during startDelay.
// The func is called from the spinner's background goroutine, so it must be safe for concurrent use
// alongside any state it reads that the caller also mutates.
func WithLabelFunc(f func() string) Option {
	return func(s *Spinner) { s.labelFunc = f }
}

// WithMaxWidth sets the cap on the rendered line, in terminal cells.
func WithMaxWidth(n int) Option {
	return func(s *Spinner) {
		if n > 0 {
			s.maxWidth = n
		}
	}
}

// Spinner animates a progress indicator on stderr.
type Spinner struct {
	done       chan struct{}
	stopped    chan struct{}
	once       sync.Once
	startDelay time.Duration
	delay      time.Duration
	frames     []string
	labelFunc  func() string
	maxWidth   int
}

// Start prints label and animates in place on stderr when stderr is a TTY.
// The animated line is cut to 80 columns; the non-TTY line is printed in full.
// On non-TTY output, such as pipes or log files, it prints the label as a line.
// Call Stop when the operation finishes.
func Start(label string, opts ...Option) *Spinner {
	label = cleanLabel(label)

	s := &Spinner{
		done:       make(chan struct{}),
		stopped:    make(chan struct{}),
		startDelay: 3 * time.Second,
		delay:      250 * time.Millisecond,
		maxWidth:   80,
		frames: []string{
			"૮(｡◕‿◕｡)っ",
			"૮(｡◕‿◕｡)つ",
			"⊂(｡◕‿◕｡)っ",
			"⊂(｡◕‿◕｡)つ",
		},
	}

	for _, o := range opts {
		o(s)
	}

	if !term.IsTerminal(int(os.Stderr.Fd())) || os.Getenv("NO_COLOR") != "" {
		fmt.Fprintln(os.Stderr, label)
		close(s.stopped)
		return s
	}

	go func() {
		defer close(s.stopped)

		current := label
		fmt.Fprintf(os.Stderr, "\r\033[2K%s", s.truncate(current))

		select {
		case <-s.done:
			fmt.Fprintf(os.Stderr, "\r\033[2K%s\n", s.truncate(current))
			return
		case <-time.After(s.startDelay):
		}

		t := time.NewTicker(s.delay)
		defer t.Stop()

		i := 0
		printFrame := func() {
			if s.labelFunc != nil {
				current = cleanLabel(s.labelFunc())
			}
			fmt.Fprintf(os.Stderr, "\r\033[2K%s", s.truncate(s.frames[i%len(s.frames)]+" "+current))
			i++
		}

		printFrame()

		for {
			select {
			case <-s.done:
				fmt.Fprintf(os.Stderr, "\r\033[2K%s\n", s.truncate(current))
				return
			case <-t.C:
				printFrame()
			}
		}
	}()

	return s
}

// Stop halts the spinner. On TTY output, the animation is erased and the label
// is left behind on its own line. Stop blocks until the spinner has stopped
// writing, and is safe to call more than once. Calling it on a Spinner that
// Start did not return does nothing.
func (s *Spinner) Stop() {
	// A Spinner that Start did not build has nil channels; closing one panics,
	// and Stop is usually deferred, where a panic would mask the caller's own.
	if s.done == nil {
		return
	}

	s.once.Do(func() {
		close(s.done)
	})
	<-s.stopped
}

// truncate cuts line to at most s.maxWidth cells. runewidth measures whole
// grapheme clusters, so a cut never splits a combining sequence.
//
// Ambiguous-width runes follow the locale runewidth detects at init, but its
// ambiguous table is a subset of UAX #11: U+25D5 in the default frames counts as
// one cell either way, so a CJK terminal rendering it wide can still wrap.
func (s *Spinner) truncate(line string) string {
	return runewidth.Truncate(line, s.maxWidth, "")
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

// Spin starts a spinner with a label, runs f, then stops the spinner.
func Spin(label string, f func(), opts ...Option) {
	sp := Start(label, opts...)
	defer sp.Stop()
	f()
}
