package rate

import (
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/gizak/termui"
	"jrubin.io/httpmon/clf"
)

type Rate struct {
	sync.RWMutex
	*termui.LineChart
	counter int
}

func New() *Rate {
	ret := Rate{LineChart: termui.NewLineChart()}
	ret.LineColor = termui.ColorGreen | termui.AttrBold
	ret.BorderLabel = ret.label(0)

	go ret.renderer()

	return &ret
}

func (r *Rate) label(rate int) string {
	return fmt.Sprintf(" Requests/s (%d/s, 1s avg) ", rate)
}

func (r *Rate) Buffer() termui.Buffer {
	r.RLock()
	ret := r.LineChart.Buffer()
	r.RUnlock()
	return ret
}

func (r *Rate) renderer() {
	ticker := time.NewTicker(1 * time.Second)
	for range ticker.C {
		r.Lock()
		r.Data = r.Data[1:]
		r.Data = append(r.Data, float64(r.counter))
		r.BorderLabel = r.label(r.counter)
		r.counter = 0
		r.Unlock()
	}
}

func (r *Rate) Handle(e clf.Entry) {
	r.Lock()
	r.counter++
	r.Unlock()
}

func (r *Rate) Resize(width, height, x, y int) {
	r.Lock()

	r.Width = width
	r.Height = height
	r.X = x
	r.Y = y

	width *= 2
	width -= 20
	if width < 0 {
		width = 0
	}

	data := make([]float64, width)
	j := width - 1
	for i := len(r.Data) - 1; i >= 0 && j >= 0; i-- {
		data[j] = r.Data[i]
		j--
	}
	r.Data = data

	r.DataLabels = make([]string, width)

	for i := 0; i < width; i++ {
		r.DataLabels[i] = strconv.FormatInt(int64(width-i), 10) + "s-ago"
	}

	r.Unlock()
}
