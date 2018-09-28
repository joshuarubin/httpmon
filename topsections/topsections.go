package topsections

import (
	"container/heap"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/gizak/termui"
	"jrubin.io/httpmon/clf"
)

const ttl = 10 * time.Second

type TopSections struct {
	sync.RWMutex
	*termui.List
	sections map[string][]time.Time
}

func New() *TopSections {
	ret := TopSections{
		List:     termui.NewList(),
		sections: map[string][]time.Time{},
	}

	ret.BorderLabel = fmt.Sprintf(" Top Sections (last %s) ", ttl)

	return &ret
}

func (t *TopSections) Buffer() termui.Buffer {
	t.RLock()
	ret := t.List.Buffer()
	t.RUnlock()
	return ret
}

type item struct {
	section string
	count   int
}

type counter []item

func (c counter) Len() int { return len(c) }

func (c counter) Less(i, j int) bool {
	if c[i].count != c[j].count {
		return c[i].count > c[j].count
	}

	return c[i].section < c[j].section
}

func (c counter) Swap(i, j int) {
	c[i], c[j] = c[j], c[i]
}

func (c *counter) Push(x interface{}) {
	*c = append(*c, x.(item))
}

func (c *counter) Pop() interface{} {
	old := *c
	n := len(old)
	i := old[n-1]
	*c = old[0 : n-1]
	return i
}

func (t *TopSections) Handle(e clf.Entry) {
	// A section is defined as being what's before the second '/' in the path.
	// For example, the section for "http://my.site.com/pages/create” is
	// "http://my.site.com/pages"

	section := strings.Join(strings.Split(e.Request.Resource, "/")[:2], "/")

	// add the new entry
	t.Lock()
	t.sections[section] = append(t.sections[section], time.Now().Add(ttl))
	t.Unlock()

	t.Render()
}

func (t *TopSections) Render() {
	var c counter
	heap.Init(&c)

	t.Lock()

	// delete expired entries and build the heap
	for section, times := range t.sections {
		var deleted int
		for i, t := range times {
			j := i - deleted
			if t.Before(time.Now()) {
				times = append(times[:j], times[j+1:]...)
				deleted++
			}
		}
		t.sections[section] = times
		if len(times) == 0 {
			delete(t.sections, section)
			continue
		}

		heap.Push(&c, item{
			section: section,
			count:   len(times),
		})
	}

	// build the list
	t.Items = nil
	for i := 0; c.Len() > 0 && i < t.Height; i++ {
		i := heap.Pop(&c).(item)
		t.Items = append(t.Items, fmt.Sprintf("[%d] %s", i.count, i.section))
	}

	t.Unlock()
}

func (t *TopSections) Resize(width, height, x, y int) {
	t.Lock()
	t.Width = width
	t.Height = height
	t.X = x
	t.Y = y
	t.Unlock()
}
