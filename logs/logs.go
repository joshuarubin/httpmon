package logs

import (
	"strings"
	"sync"

	"github.com/gizak/termui"
	"jrubin.io/httpmon/clf"
)

type Logs struct {
	sync.RWMutex
	*termui.Par
	entries []string
}

func New() *Logs {
	ret := Logs{Par: termui.NewPar("")}
	ret.BorderLabel = " Logs "

	return &ret
}

func (l *Logs) Buffer() termui.Buffer {
	l.RLock()
	ret := l.Par.Buffer()
	l.RUnlock()
	return ret
}

func (l *Logs) Handle(e clf.Entry) {
	l.Lock()

	l.entries = append(l.entries, termui.TrimStrIfAppropriate(e.String(), l.Width-2))
	if len(l.entries) > l.Height-2 {
		l.entries = l.entries[1:]
	}

	l.Text = strings.Join(l.entries, "\n")

	l.Unlock()
}

func (l *Logs) Resize(width, height, x, y int) {
	l.Lock()

	l.Width = width
	l.Height = height
	l.X = x
	l.Y = y

	l.Unlock()
}
