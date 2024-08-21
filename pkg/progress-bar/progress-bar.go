package progress_bar

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"golang.org/x/term"
)

var ErrProgressBar = errors.New("writing progress bar")

type ProgressBar struct {
	name      string
	out       io.Writer
	startOnce sync.Once
	stop      chan struct{}
	ticker    *time.Ticker

	startTime   time.Time
	progress    atomic.Int64
	maxProgress int64
}

func NewProgressBar(name string, maxProgress int64, out io.Writer) *ProgressBar {
	return &ProgressBar{
		name:      name,
		out:       out,
		startOnce: sync.Once{},
		stop:      make(chan struct{}, 1),
		ticker:    time.NewTicker(500 * time.Millisecond),

		startTime:   time.Now(),
		progress:    atomic.Int64{},
		maxProgress: maxProgress,
	}
}

func (p *ProgressBar) Increment() {
	p.progress.Add(1)
}

func (p *ProgressBar) Set(progress int64) {
	p.progress.Swap(progress)
}

func (p *ProgressBar) Start() {
	go func() {
		p.startOnce.Do(func() {
			for {
				select {
				case <-p.stop:
					return
				case <-p.ticker.C:
					p.printProgress()
				}
			}
		})
	}()
}

func (p *ProgressBar) printProgress() {
	width, _, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		fmt.Println(fmt.Errorf("%w: %w", ErrProgressBar, err))
		return
	}

	maxProgress := strconv.FormatInt(p.maxProgress, 10)
	countIndicatorFormat := fmt.Sprintf("(%%%dd/%%s)", len(maxProgress))
	progress := p.progress.Load()
	countIndicator := fmt.Sprintf(countIndicatorFormat, progress, maxProgress)

	progressPercent := float64(p.progress.Load()) / float64(p.maxProgress)

	estimatedTimeLeftString := "--:--:--"
	if progressPercent > 0.0 {
		now := time.Now()
		timeSinceStart := now.Sub(p.startTime)
		estimatedDuration := time.Duration(float64(timeSinceStart) / progressPercent)
		estimatedEnd := p.startTime.Add(estimatedDuration)
		estimatedTimeLeft := estimatedEnd.Sub(now)
		hoursLeft := estimatedTimeLeft / time.Hour
		minutesLeft := (estimatedTimeLeft % time.Hour) / time.Minute
		secondsLeft := (estimatedTimeLeft % time.Minute) / time.Second
		estimatedTimeLeftString = fmt.Sprintf("%02d:%02d:%02d", hoursLeft, minutesLeft, secondsLeft)
	}

	usableProgressBarWidth := width - len(p.name) - 7 - len(countIndicator) - len(estimatedTimeLeftString)
	doneWidth := int(float64(usableProgressBarWidth) * progressPercent)
	doneBar := strings.Repeat("=", doneWidth)
	emptyBar := strings.Repeat(" ", usableProgressBarWidth-doneWidth)
	_, err = p.out.Write([]byte(
		fmt.Sprintf("\r%s: [%s>%s] %s %s", p.name, doneBar, emptyBar, countIndicator, estimatedTimeLeftString)))
	if err != nil {
		fmt.Println(fmt.Errorf("%w: %w", ErrProgressBar, err))
	}

	if progress == p.maxProgress {
		// Finished, no need to continue printing progress bar.
		_, err = p.out.Write([]byte("\n"))
		if err != nil {
			fmt.Println(fmt.Errorf("%w: %w", ErrProgressBar, err))
		}
		p.stop <- struct{}{}
	}
}

func (p *ProgressBar) Stop() {
	p.Set(p.maxProgress)
	p.printProgress()
	p.stop <- struct{}{}
	close(p.stop)
}
