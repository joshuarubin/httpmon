package status

import (
	"sync"

	"github.com/gizak/termui"
	"jrubin.io/httpmon/clf"
)

type Status struct {
	sync.RWMutex
	*termui.BarChart
}

func New() *Status {
	ret := Status{BarChart: termui.NewBarChart()}

	ret.BorderLabel = " Status Codes "
	ret.Data = make([]int, 5)
	ret.DataLabels = []string{"1xx", "2xx", "3xx", "4xx", "5xx"}

	return &ret
}

func (s *Status) Buffer() termui.Buffer {
	s.RLock()
	ret := s.BarChart.Buffer()
	s.RUnlock()
	return ret
}

const (
	code1xx = iota
	code2xx
	code3xx
	code4xx
	code5xx
)

func (s *Status) Handle(e clf.Entry) {
	s.Lock()
	switch {
	case e.StatusCode <= 199:
		s.Data[code1xx]++
	case e.StatusCode <= 299:
		s.Data[code2xx]++
	case e.StatusCode <= 399:
		s.Data[code3xx]++
	case e.StatusCode <= 499:
		s.Data[code4xx]++
	default:
		s.Data[code5xx]++
	}
	s.Unlock()
}

func (s *Status) Resize(width, height, x, y int) {
	s.Lock()
	s.Width = width
	s.Height = height
	s.BarWidth = int(float64(width-4) / 5)
	s.X = x
	s.Y = y
	s.Unlock()
}
