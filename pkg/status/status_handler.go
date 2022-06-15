package status

import (
	"context"
	"github.com/kluctl/kluctl/v2/pkg/utils"
	"github.com/kluctl/kluctl/v2/pkg/utils/term"
	"github.com/vbauerster/mpb/v7"
	"github.com/vbauerster/mpb/v7/decor"
	"io"
	"math"
	"strings"
	"syscall"
)

type MultiLineStatusHandler struct {
	ctx        context.Context
	out        io.Writer
	isTerminal bool
	progress   *mpb.Progress
	trace      bool
}

type statusLine struct {
	slh     *MultiLineStatusHandler
	total   int
	bar     *mpb.Bar
	filler  mpb.BarFiller
	message string

	barOverride string
}

func NewMultiLineStatusHandler(ctx context.Context, out io.Writer, isTerminal bool, trace bool) *MultiLineStatusHandler {
	sh := &MultiLineStatusHandler{
		ctx:        ctx,
		out:        out,
		isTerminal: isTerminal,
		trace:      trace,
	}

	sh.start()

	return sh
}

func (s *MultiLineStatusHandler) IsTerminal() bool {
	return s.isTerminal
}

func (s *MultiLineStatusHandler) Flush() {
	s.Stop()
	s.start()
}

func (s *MultiLineStatusHandler) SetTrace(trace bool) {
	s.trace = trace
}

func (s *MultiLineStatusHandler) start() {
	s.progress = mpb.NewWithContext(
		s.ctx,
		mpb.WithWidth(utils.GetTermWidth()),
		mpb.WithOutput(s.out),
		mpb.PopCompletedMode(),
	)
}

func (s *MultiLineStatusHandler) Stop() {
	s.progress.Wait()
}

func (s *MultiLineStatusHandler) StartStatus(total int, message string) StatusLine {
	return s.startStatus(total, message, 0, "")
}

func (s *MultiLineStatusHandler) startStatus(total int, message string, priority int, barOverride string) *statusLine {
	sl := &statusLine{
		slh:         s,
		total:       total,
		message:     message,
		barOverride: barOverride,
	}
	sl.filler = mpb.SpinnerStyle().PositionLeft().Build()

	opts := []mpb.BarOption{
		mpb.BarWidth(1),
		mpb.AppendDecorators(decor.Any(sl.DecorMessage, decor.WCSyncWidthR)),
	}
	if priority != 0 {
		opts = append(opts, mpb.BarPriority(priority))
	}

	sl.bar = s.progress.Add(int64(total), sl, opts...)

	return sl
}

func (sl *statusLine) DecorMessage(s decor.Statistics) string {
	return sl.message
}

func (sl *statusLine) Fill(w io.Writer, reqWidth int, stat decor.Statistics) {
	if sl.barOverride != "" {
		_, _ = io.WriteString(w, sl.barOverride)
		return
	}

	sl.filler.Fill(w, reqWidth, stat)
}

func (s *MultiLineStatusHandler) Info(message string) {
	o := withColor("green", "ⓘ")
	s.startStatus(1, message, math.MinInt, o).end(o)
}

func (s *MultiLineStatusHandler) InfoFallback(message string) {
	// no fallback needed
}

func (s *MultiLineStatusHandler) Warning(message string) {
	o := withColor("yellow", "⚠")
	s.startStatus(1, message, math.MinInt, o).end(o)
}

func (s *MultiLineStatusHandler) Error(message string) {
	o := withColor("red", "✗")
	s.startStatus(1, message, math.MinInt, o).end(o)
}

func (s *MultiLineStatusHandler) Trace(message string) {
	if s.trace {
		s.Info(message)
	}
}

func (s *MultiLineStatusHandler) PlainText(text string) {
	s.Info(text)
}

func (s *MultiLineStatusHandler) Prompt(password bool, message string) (string, error) {
	o := withColor("yellow", "?")
	sl := s.startStatus(1, message, math.MinInt, o)
	defer sl.end(o)

	doUpdate := func(ret []byte) {
		if password {
			sl.Update(message + strings.Repeat("*", len(ret)))
		} else {
			sl.Update(message + string(ret))
		}
	}

	ret, err := term.ReadLineNoEcho(int(syscall.Stdin), doUpdate)

	return string(ret), err
}

func (sl *statusLine) SetTotal(total int) {
	sl.total = total
	sl.bar.SetTotal(int64(total), false)
	sl.bar.EnableTriggerComplete()
}

func (sl *statusLine) Increment() {
	sl.bar.Increment()
}

func (sl *statusLine) Update(message string) {
	sl.message = message
}

func (sl *statusLine) end(barOverride string) {
	sl.barOverride = barOverride
	// make sure that the bar es rendered on top so that it can be properly popped
	sl.bar.SetPriority(math.MinInt)
	sl.bar.SetCurrent(int64(sl.total))
}

func (sl *statusLine) End(result EndResult) {
	switch result {
	case EndSuccess:
		sl.end(withColor("green", "✓"))
	case EndWarning:
		sl.end(withColor("yellow", "⚠"))
	case EndError:
		sl.end(withColor("red", "✗"))
	}
}
