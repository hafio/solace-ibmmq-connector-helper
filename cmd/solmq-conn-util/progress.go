package main

// Progress reporting for status, which is the one verb that can spend a long
// time collecting before it has anything to print. A status run makes two
// container execs per instance and the script inside each one makes seven
// sequential actuator calls, so on a real cluster the report can be tens of
// seconds behind the command -- with nothing on screen to say the run is
// advancing rather than hung.
//
// Two renderings of the same steps, chosen once per run by newProgress:
//
//   - spinner (the default): one line on stderr, rewritten in place, naming the
//     step in flight. Only when stderr is a terminal, since the rewrite is
//     meaningless in a file.
//   - steps (--verbose): one durable line per step, printed when the step ends,
//     always carrying its elapsed time -- which is what answers "why is this
//     environment slow", and what survives a `2>` into a log.
//
// Everything here writes to stderr and only to stderr: the report is the
// artifact and it owns stdout, so `status --output json > doc.json` is
// byte-identical with and without progress. The in-place rewrite is a bare
// `\r` and the erase blanks the line with spaces -- no ANSI at all, so
// unlike clearScreen this needs no VT processing enabled and works in a plain
// console host.

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"time"
)

// progressMode is how a run reports the step it is on. progressOff is the zero
// value, so a progress that was never configured reports nothing rather than
// writing to a stream nobody is watching.
type progressMode int

const (
	progressOff progressMode = iota
	progressSpinner
	progressSteps
)

// progressFrames is the spinner. Plain ASCII, like every other byte this tool
// writes: a braille or box-drawing spinner is prettier and unreadable in the
// consoles and log captures this output has to survive.
var progressFrames = []string{"-", "\\", "|", "/"}

// progressLineWidth caps the live line at one row of a classic 80-column
// console. Nothing here can ask the terminal how wide it is -- that needs an
// ioctl on unix and a console call on Windows, and this file deliberately
// depends on neither -- so the narrowest width worth designing for is the cap.
// It has to hold, because a line that wraps is a line the erase cannot undo: a
// carriage return goes to column 0 of the row the cursor is on and leaves
// whatever wrapped above it on screen for the report to land under. Labels do
// reach that far -- a pod name alone can be 63 characters -- so the label is
// trimmed rather than the terminal trusted.
const progressLineWidth = 79

// progressFrameRate and progressNow are the two test seams, vars for the same
// reason stdinIsTerminal is one: a hard-coded 120ms tick and a real clock would
// make the rendered output untestable and the elapsed times unpinnable.
var (
	progressFrameRate = 120 * time.Millisecond
	progressNow       = time.Now
)

// progress reports the step a status run is on. Every method is a no-op on a
// nil receiver and in progressOff, so call sites are one unconditional line
// with no branching -- the same rule enableVirtualTerminal follows.
type progress struct {
	w    *os.File
	mode progressMode

	// mu guards everything below it: the frame ticker runs in its own goroutine
	// and the collectors call step from theirs, so the label and the live line's
	// width are shared state.
	mu    sync.Mutex
	label string    // the step in flight; "" when none is
	start time.Time // when that step began
	frame int
	width int // widest live line written since the last erase, so the erase covers it

	// quit and joined are non-nil exactly while the frame goroutine is running.
	// joined is what stop waits on: a goroutine still writing after stop
	// returned would race captureStderr's swap of os.Stderr in the tests, and
	// would draw a frame over the report in production.
	quit   chan struct{}
	joined chan struct{}
}

// newProgress picks the one rendering this run uses. The order is the order the
// flags were settled in: --watch wins over everything (its redraw owns the
// screen, and step lines would scroll under a report being cleared every tick),
// then --verbose, which is an explicit request and so prints whether or not
// stderr is a terminal, and only then the spinner, which needs one.
func newProgress(o statusOpts, w *os.File) *progress {
	switch {
	case o.watch.on:
		return &progress{w: w, mode: progressOff}
	case o.verbose:
		return &progress{w: w, mode: progressSteps}
	case stderrIsTerminal():
		return &progress{w: w, mode: progressSpinner}
	default:
		return &progress{w: w, mode: progressOff}
	}
}

// step ends whatever step was in flight and begins this one. Steps are named
// after what is being waited on, with the instance and i/N where there is one
// call per instance, since that is what tells an operator the run is advancing
// rather than stuck on the same exec.
func (p *progress) step(format string, args ...any) {
	if p == nil || p.mode == progressOff {
		return
	}
	label := fmt.Sprintf(format, args...)
	p.quiesce()

	p.mu.Lock()
	defer p.mu.Unlock()
	p.finishLocked()
	p.label, p.start, p.frame = label, progressNow(), 0
	if p.mode == progressSpinner {
		p.drawLocked()
		p.startLocked()
	}
}

// stop ends the step in flight and leaves the stream as it found it: the live
// line erased, no goroutine running. Idempotent, and a later step starts a new
// one -- which is what lets both actStatus and collect defer it, so no path out
// of either can leave a half-drawn spinner on screen.
func (p *progress) stop() {
	if p == nil || p.mode == progressOff {
		return
	}
	p.quiesce()
	p.mu.Lock()
	defer p.mu.Unlock()
	p.finishLocked()
}

// pause hands stderr back for the duration of fn. The step in flight ends
// first, so no frame is drawn over the question the install confirmation asks
// on stderr (see readStdinLine) and, under --verbose, that step reports the
// time the call took rather than the time the operator took to answer. Nothing
// is restored afterwards because nothing needs to be: every site that can
// prompt opens a new step before its next slow call.
func (p *progress) pause(fn func()) {
	p.stop()
	fn()
}

// quiesce stops the frame goroutine and returns only once it has exited, so
// nothing can draw between here and the caller's next write. Callers must not
// hold mu: the goroutine takes it on every frame.
func (p *progress) quiesce() {
	p.mu.Lock()
	quit, joined := p.quit, p.joined
	p.quit, p.joined = nil, nil
	p.mu.Unlock()
	if quit == nil {
		return
	}
	close(quit)
	<-joined
}

// startLocked starts the frame goroutine if it is not already running. The
// channels are passed to it rather than read off p, so a stop-then-step pair
// cannot leave the old goroutine watching the new channel.
func (p *progress) startLocked() {
	if p.quit != nil {
		return
	}
	quit, joined := make(chan struct{}), make(chan struct{})
	p.quit, p.joined = quit, joined
	go p.spin(quit, joined)
}

// spin advances the live line until quit closes. It is the only goroutine this
// tool runs besides logs's signal watcher.
func (p *progress) spin(quit, joined chan struct{}) {
	defer close(joined)
	tick := time.NewTicker(progressFrameRate)
	defer tick.Stop()
	for {
		select {
		case <-quit:
			return
		case <-tick.C:
			p.mu.Lock()
			if p.label != "" {
				p.frame++
				p.drawLocked()
			}
			p.mu.Unlock()
		}
	}
}

// finishLocked closes out the step in flight: one durable line with its elapsed
// time in --verbose, or the erase of the live line for the spinner.
func (p *progress) finishLocked() {
	if p.label == "" {
		return
	}
	if p.mode == progressSteps {
		fmt.Fprintf(p.w, "step: %s %s\n", p.label, progressElapsed(progressNow().Sub(p.start)))
	} else {
		p.eraseLocked()
	}
	p.label = ""
}

// drawLocked rewrites the live line in place: return to column 0, then write
// the line. Nothing is padded, because within one step the line only ever grows
// -- the label is fixed and the seconds counter only counts up -- and every new
// label is preceded by an erase.
func (p *progress) drawLocked() {
	head := progressFrames[p.frame%len(progressFrames)] + " "
	// Whole seconds here, not the tenths finishLocked prints: this is a counter
	// an operator watches, and a tenths digit changing eight times a second
	// reads as noise. Under a second there is nothing worth counting yet.
	tail := ""
	if d := progressNow().Sub(p.start); d >= time.Second {
		tail = fmt.Sprintf("  %ds", int(d.Seconds()))
	}
	line := head + progressTrim(p.label, progressLineWidth-len(head)-len(tail)) + tail
	fmt.Fprintf(p.w, "\r%s", line)
	if len(line) > p.width {
		p.width = len(line)
	}
}

// progressTrim shortens a label to fit one row, eliding the middle rather than
// the tail: a label opens with the call being waited on and closes with the
// instance and its i/N, and the i/N is the half that says the run is advancing
// rather than stuck on the same exec -- cutting it would remove the reason the
// counter is there at all. What goes instead is the middle of a long generated
// name, which is its least distinguishing part.
func progressTrim(label string, width int) string {
	const gap = "..."
	switch {
	case width <= 0:
		return ""
	case len(label) <= width:
		return label
	case width <= len(gap):
		// Narrower than the elision itself, so there is nothing left to signal
		// with: the label is cut hard rather than reduced to dots.
		return label[:width]
	}
	keep := width - len(gap)
	head := (keep + 1) / 2
	return label[:head] + gap + label[len(label)-(keep-head):]
}

// eraseLocked blanks the live line and leaves the cursor at column 0, so
// whatever is written next starts on a clean line -- including the report on
// stdout, which shares the terminal with this. There is no zero-width case to
// guard: finishLocked calls this only with a label in flight, and a label is
// never set in spinner mode without the draw that follows it in the same
// critical section, so the width is always the width of something on screen.
// The cap in drawLocked is what keeps that width to one row, so the blanking
// cannot wrap either.
func (p *progress) eraseLocked() {
	fmt.Fprintf(p.w, "\r%s\r", strings.Repeat(" ", p.width))
	p.width = 0
}

// progressElapsed renders a step's duration the way the report renders ages:
// short, fixed-shape, and never more precise than is useful. Tenths below a
// minute, because the whole point is telling a 0.3s call from a 4.1s one.
func progressElapsed(d time.Duration) string {
	if d < 0 {
		d = 0
	}
	if d < time.Minute {
		return fmt.Sprintf("%.1fs", d.Seconds())
	}
	return fmt.Sprintf("%dm%02ds", int(d.Minutes()), int(d.Seconds())%60)
}
