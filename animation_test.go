package spinner

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/mattn/go-runewidth"
)

// eraseSeq prefixes every write the animation makes: return to column 0, then
// clear the row.
const eraseSeq = "\r\x1b[2K"

// lockedBuffer lets the test read while the spinner goroutine is still writing.
type lockedBuffer struct {
	mu  sync.Mutex
	buf bytes.Buffer
}

func (b *lockedBuffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	return b.buf.Write(p)
}

func (b *lockedBuffer) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()

	return b.buf.String()
}

// capture redirects output and forces the animated path. Start decides both from
// the real stderr, which a test cannot be.
func capture(w io.Writer) Option {
	return func(s *Spinner) {
		s.w = w
		s.animate = true
	}
}

// segments splits captured output into the payload of each write.
func segments(out string) []string {
	parts := strings.Split(out, eraseSeq)
	if len(parts) > 0 && parts[0] == "" {
		parts = parts[1:]
	}

	return parts
}

// waitFor polls until cond holds, so the tests do not race a fixed sleep against
// the ticker.
func waitFor(t *testing.T, buf *lockedBuffer, cond func(string) bool, want string) {
	t.Helper()

	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if cond(buf.String()) {
			return
		}
		time.Sleep(time.Millisecond)
	}

	t.Fatalf("timed out waiting for %s, got %q", want, buf.String())
}

func TestAnimationShowsLabelBeforeStartDelay(t *testing.T) {
	var buf lockedBuffer

	sp := Start("Loading", capture(&buf),
		WithStartDelay(time.Hour),
		WithFrames([]string{"A", "B"}),
	)
	waitFor(t, &buf, func(s string) bool { return strings.Contains(s, "Loading") }, "the label")
	sp.Stop()

	got := buf.String()
	if strings.ContainsAny(got, "AB") {
		t.Errorf("a frame rendered before startDelay elapsed: %q", got)
	}

	want := eraseSeq + "Loading" + eraseSeq + "Loading\n"
	if got != want {
		t.Errorf("output = %q, want %q", got, want)
	}
}

func TestAnimationCyclesFramesInOrder(t *testing.T) {
	var buf lockedBuffer

	sp := Start("x", capture(&buf),
		WithStartDelay(0),
		WithDelay(time.Millisecond),
		WithFrames([]string{"A", "B", "C"}),
	)
	waitFor(t, &buf, func(s string) bool {
		return strings.Contains(s, "C x")
	}, "a full cycle through the frames")
	sp.Stop()

	segs := segments(buf.String())
	if len(segs) < 4 {
		t.Fatalf("only %d writes: %q", len(segs), buf.String())
	}

	// segs[0] is the pre-delay label, then frames in order from segs[1].
	for i, want := range []string{"A x", "B x", "C x"} {
		if segs[1+i] != want {
			t.Errorf("write %d = %q, want %q", 1+i, segs[1+i], want)
		}
	}
}

func TestAnimationFinalWriteRestoresLabel(t *testing.T) {
	var buf lockedBuffer

	sp := Start("Done soon", capture(&buf),
		WithStartDelay(0),
		WithDelay(time.Millisecond),
		WithFrames([]string{"A"}),
	)
	waitFor(t, &buf, func(s string) bool { return strings.Contains(s, "A Done soon") }, "a frame")
	sp.Stop()

	got := buf.String()
	if !strings.HasSuffix(got, eraseSeq+"Done soon\n") {
		t.Errorf("output does not end with an erase, the bare label and a newline: %q", got)
	}

	segs := segments(got)
	if last := segs[len(segs)-1]; strings.Contains(last, "A ") {
		t.Errorf("final write still carries a frame: %q", last)
	}
}

func TestAnimationEveryWriteErasesFirst(t *testing.T) {
	var buf lockedBuffer

	sp := Start("x", capture(&buf),
		WithStartDelay(0),
		WithDelay(time.Millisecond),
		WithFrames([]string{"A", "B"}),
	)
	waitFor(t, &buf, func(s string) bool { return strings.Contains(s, "B x") }, "a frame")
	sp.Stop()

	got := buf.String()
	if !strings.HasPrefix(got, eraseSeq) {
		t.Fatalf("output does not open with an erase: %q", got)
	}

	// Every payload must be free of the sequence, i.e. writes never batch up
	// without clearing the row between them.
	for i, seg := range segments(got) {
		if strings.Contains(strings.TrimSuffix(seg, "\n"), "\r") {
			t.Errorf("write %d contains a stray carriage return: %q", i, seg)
		}
	}
}

func TestAnimationUsesLabelFuncEachTick(t *testing.T) {
	var (
		buf lockedBuffer
		n   atomic.Int64
	)

	sp := Start("initial", capture(&buf),
		WithStartDelay(0),
		WithDelay(time.Millisecond),
		WithFrames([]string{"A"}),
		WithLabelFunc(func() string { return fmt.Sprintf("tick-%d", n.Add(1)) }),
	)
	waitFor(t, &buf, func(s string) bool {
		return strings.Contains(s, "tick-1") && strings.Contains(s, "tick-3")
	}, "three regenerated labels")
	sp.Stop()

	got := buf.String()
	if !strings.HasPrefix(got, eraseSeq+"initial") {
		t.Errorf("pre-delay write should use the Start label, got %q", got)
	}
	if strings.HasSuffix(got, "initial\n") {
		t.Error("final write fell back to the Start label instead of the generated one")
	}
}

func TestAnimationCapsRenderedWidth(t *testing.T) {
	var buf lockedBuffer

	sp := Start(strings.Repeat("a", 200), capture(&buf),
		WithStartDelay(0),
		WithDelay(time.Millisecond),
		WithFrames([]string{"AAAA"}),
		WithMaxWidth(20),
	)
	waitFor(t, &buf, func(s string) bool { return strings.Contains(s, "AAAA a") }, "a frame")
	sp.Stop()

	for i, seg := range segments(buf.String()) {
		if w := runewidth.StringWidth(strings.TrimSuffix(seg, "\n")); w > 20 {
			t.Errorf("write %d is %d cells, want at most 20: %q", i, w, seg)
		}
	}
}

// The line Stop leaves behind outlives the animation, so it has to show the state
// at Stop, not the state at the last tick. A long delay guarantees no tick lands
// between the state change and Stop.
func TestStopRefreshesTheLabel(t *testing.T) {
	var (
		buf lockedBuffer
		n   atomic.Int64
	)

	sp := Start("initial", capture(&buf),
		WithStartDelay(0),
		WithDelay(time.Hour),
		WithFrames([]string{"A"}),
		WithLabelFunc(func() string { return fmt.Sprintf("n=%d", n.Load()) }),
	)
	waitFor(t, &buf, func(s string) bool { return strings.Contains(s, "n=0") }, "the first frame")

	n.Store(7)
	sp.Stop()

	if got := buf.String(); !strings.HasSuffix(got, "n=7\n") {
		t.Errorf("final line did not pick up the label change, got %q", got)
	}
}

// Same guarantee when Stop arrives before the animation ever starts.
func TestStopRefreshesTheLabelBeforeStartDelay(t *testing.T) {
	var (
		buf lockedBuffer
		n   atomic.Int64
	)

	sp := Start("initial", capture(&buf),
		WithStartDelay(time.Hour),
		WithLabelFunc(func() string { return fmt.Sprintf("n=%d", n.Load()) }),
	)
	waitFor(t, &buf, func(s string) bool { return strings.Contains(s, "initial") }, "the start label")

	n.Store(3)
	sp.Stop()

	got := buf.String()
	if !strings.HasPrefix(got, eraseSeq+"initial") {
		t.Errorf("the line shown during startDelay should be the Start label, got %q", got)
	}
	if !strings.HasSuffix(got, "n=3\n") {
		t.Errorf("final line did not pick up the label change, got %q", got)
	}
}

// The stopped-channel handshake has to mean the goroutine is done writing, not
// merely that it was told to stop.
func TestAnimationWritesNothingAfterStop(t *testing.T) {
	var buf lockedBuffer

	sp := Start("x", capture(&buf),
		WithStartDelay(0),
		WithDelay(time.Millisecond),
		WithFrames([]string{"A", "B"}),
	)
	waitFor(t, &buf, func(s string) bool { return strings.Contains(s, "B x") }, "a frame")
	sp.Stop()

	settled := buf.String()
	time.Sleep(50 * time.Millisecond)

	if got := buf.String(); got != settled {
		t.Errorf("spinner wrote %q after Stop returned", strings.TrimPrefix(got, settled))
	}
}

func TestPlainWriteWhenNotAnimating(t *testing.T) {
	var buf lockedBuffer

	sp := Start("Loading", func(s *Spinner) { s.w = &buf; s.animate = false })
	sp.Stop()

	if got := buf.String(); got != "Loading\n" {
		t.Errorf("output = %q, want %q", got, "Loading\n")
	}
}

func TestPlainWriteCleansLabel(t *testing.T) {
	var buf lockedBuffer

	sp := Start("two\nlines\there\x07", func(s *Spinner) { s.w = &buf; s.animate = false })
	sp.Stop()

	if got := buf.String(); got != "two lines here\n" {
		t.Errorf("output = %q, want %q", got, "two lines here\n")
	}
}
