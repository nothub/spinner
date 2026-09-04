package spinner

import (
	"strings"
	"testing"
	"time"

	"github.com/mattn/go-runewidth"
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

// Guards the default wired up in Start, not just the field.
func TestStartDefaultsToEightyColumns(t *testing.T) {
	sp := Start("")
	defer sp.Stop()

	if sp.maxWidth != 80 {
		t.Errorf("Start() left maxWidth at %d, want 80", sp.maxWidth)
	}
}

// The cap has to reach truncate, not merely be stored on the struct.
func TestTruncateUsesMaxWidth(t *testing.T) {
	s := &Spinner{maxWidth: 10}

	if got, want := s.truncate(strings.Repeat("a", 20)), strings.Repeat("a", 10); got != want {
		t.Errorf("truncate() = %q, want %q", got, want)
	}
}

// Cases stick to runes whose width is locale-independent. Ambiguous-width runes
// such as U+2282 are deliberately absent: runewidth widens those under a CJK
// locale, so an expectation built on one would depend on the environment.
func TestTruncate(t *testing.T) {
	s := &Spinner{maxWidth: 80}
	maxWidth := s.maxWidth

	tests := []struct {
		name string
		in   string
		want string
	}{
		{"short", "spinning", "spinning"},
		{"exactly maxWidth", strings.Repeat("a", maxWidth), strings.Repeat("a", maxWidth)},
		{"over maxWidth", strings.Repeat("a", maxWidth+10), strings.Repeat("a", maxWidth)},
		{"wide runes cost two cells", strings.Repeat("っ", maxWidth), strings.Repeat("っ", maxWidth/2)},
		{"combining marks are free", strings.Repeat("e\u0301", maxWidth), strings.Repeat("e\u0301", maxWidth)},
		{"wide rune straddling the cap is dropped", strings.Repeat("a", maxWidth-1) + "っ", strings.Repeat("a", maxWidth-1)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := s.truncate(tt.in)
			if got != tt.want {
				t.Errorf("truncate() = %q, want %q", got, tt.want)
			}

			if c := runewidth.StringWidth(got); c > maxWidth {
				t.Errorf("truncate() = %d cells, want at most %d", c, maxWidth)
			}
		})
	}
}
