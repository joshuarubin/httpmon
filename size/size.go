package size

import (
	"sync"
	"time"

	"github.com/gizak/termui"
	"jrubin.io/httpmon/clf"
)

type Size struct {
	sync.RWMutex
	*termui.Sparklines
	num, size int
}

func New() *Size {
	ret := Size{Sparklines: termui.NewSparklines(termui.NewSparkline())}
	ret.BorderLabel = " Response Sizes (1s avg)"

	go ret.renderer()

	return &ret
}

func (s *Size) Buffer() termui.Buffer {
	s.RLock()
	ret := s.Sparklines.Buffer()
	s.RUnlock()
	return ret
}

func (s *Size) renderer() {
	ticker := time.NewTicker(1 * time.Second)
	for range ticker.C {
		s.Lock()
		s.Lines[0].Data = s.Lines[0].Data[1:]
		if s.num > 0 {
			s.Lines[0].Data = append(s.Lines[0].Data, s.size/s.num)
		} else {
			s.Lines[0].Data = append(s.Lines[0].Data, 0)
		}
		s.size = 0
		s.num = 0
		s.Unlock()
	}
}

func (s *Size) Handle(e clf.Entry) {
	s.Lock()
	s.num++
	s.size += e.Size
	s.Unlock()
}

func (s *Size) Resize(width, height, x, y int) {
	s.Lock()

	s.Width = width
	s.Height = height
	s.X = x
	s.Y = y

	s.Lines[0].Height = height - 3

	data := make([]int, width)
	j := width - 1
	for i := len(s.Lines[0].Data) - 1; i >= 0 && j >= 0; i-- {
		data[j] = s.Lines[0].Data[i]
		j--
	}
	s.Lines[0].Data = data

	s.Unlock()
}
