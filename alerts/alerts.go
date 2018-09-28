package alerts

import (
	"fmt"
	"sync"
	"time"

	"github.com/gizak/termui"
	"jrubin.io/httpmon/clf"
)

type Alerts struct {
	sync.RWMutex
	Max float64
	Dur time.Duration
	*termui.Par
	counter int
	start   time.Time
	rates   []int
}

func New(max float64, dur time.Duration) *Alerts {
	if dur < time.Second {
		panic("invalid duration")
	}

	ret := Alerts{
		Max:   max,
		Dur:   dur,
		Par:   termui.NewPar(""),
		rates: make([]int, dur/time.Second),
	}

	ret.BorderLabel = ret.label(ret.avg())

	go ret.renderer()

	return &ret
}

func (a *Alerts) label(avg float64) string {
	return fmt.Sprintf(" Alerts (threshold: %.2f/s, current: %.2f/s, %s avg) ", a.Max, avg, a.Dur)
}

func (a *Alerts) avg() float64 {
	var avg float64
	for _, r := range a.rates {
		avg += float64(r)
	}

	avg /= float64(len(a.rates))
	return avg
}

func (a *Alerts) Buffer() termui.Buffer {
	a.RLock()
	ret := a.Par.Buffer()
	a.RUnlock()
	return ret
}

func (a *Alerts) renderer() {
	ticker := time.NewTicker(1 * time.Second)
	for range ticker.C {
		a.Lock()

		a.rates = a.rates[1:]
		a.rates = append(a.rates, a.counter)
		a.counter = 0

		avg := a.avg()
		a.BorderLabel = a.label(avg)

		if a.start == (time.Time{}) && avg >= a.Max {
			a.start = time.Now()
		}

		if a.start != (time.Time{}) && avg >= a.Max {
			a.Text = fmt.Sprintf("High traffic generated an alert\nTriggered at %s", a.start.Format(time.Stamp))
		}

		if a.start != (time.Time{}) && avg < a.Max {
			a.start = time.Time{}
			a.Text = fmt.Sprintf("Last alert recovered at %s", time.Now().Format(time.Stamp))
		}

		a.Unlock()
	}
}

func (a *Alerts) Handle(e clf.Entry) {
	a.Lock()
	a.counter++
	a.Unlock()
}

func (a *Alerts) Resize(width, height, x, y int) {
	a.Lock()

	a.Width = width
	a.Height = height
	a.X = x
	a.Y = y

	a.Unlock()
}
