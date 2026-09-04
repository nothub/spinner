package spinner

import (
	"testing"
	"time"
)

func TestWithDelayIgnoresNonPositive(t *testing.T) {
	for _, d := range []time.Duration{0, -time.Second} {
		s := &Spinner{delay: 250 * time.Millisecond}
		WithDelay(d)(s)

		if s.delay != 250*time.Millisecond {
			t.Errorf("WithDelay(%v) set delay to %v, want the default kept", d, s.delay)
		}
	}
}

func TestWithFramesIgnoresEmpty(t *testing.T) {
	for _, frames := range [][]string{nil, {}} {
		s := &Spinner{frames: []string{"a"}}
		WithFrames(frames)(s)

		if len(s.frames) != 1 {
			t.Errorf("WithFrames(%q) set frames to %q, want the default kept", frames, s.frames)
		}
	}
}
