package main

import (
	"bufio"
	"flag"
	"io"
	"log"
	"os"
	"time"

	"github.com/gizak/termui"
	"jrubin.io/httpmon/alerts"
	"jrubin.io/httpmon/clf"
	"jrubin.io/httpmon/internal"
	"jrubin.io/httpmon/logs"
	"jrubin.io/httpmon/rate"
	"jrubin.io/httpmon/size"
	"jrubin.io/httpmon/status"
	"jrubin.io/httpmon/tail"
	"jrubin.io/httpmon/topsections"
)

type config struct {
	File      string
	AlertRate float64
	AlertDur  time.Duration
	NoCatLog  bool
}

func main() {
	var c config

	flag.StringVar(&c.File, "file", "/var/log/access.log", "file to read, \"-\" for stdin")
	flag.Float64Var(&c.AlertRate, "alert-rate", 10, "number of requests/s for an alert to be triggered")
	flag.DurationVar(&c.AlertDur, "alert-duration", 2*time.Minute, "the request rate must exceed alert-rate for this much time, on average, before an alert is triggered")
	flag.BoolVar(&c.NoCatLog, "no-cat-log", false, "don't show the log lines")

	flag.Parse()

	if err := run(c); err != nil {
		log.Fatalf("%+v", err)
	}
}

func run(c config) error {
	rdr := io.Reader(os.Stdin)
	if c.File != "-" {
		f, err := tail.New(c.File)
		if err != nil {
			return err
		}
		defer f.Close()
		rdr = f
	}

	if err := termui.Init(); err != nil {
		return err
	}
	defer termui.Close()

	ts := topsections.New()
	r := rate.New()
	st := status.New()
	sz := size.New()
	a := alerts.New(c.AlertRate, c.AlertDur)

	bufferers := []termui.Bufferer{ts, r, st, sz, a}

	rows := 3
	var lg *logs.Logs
	if c.NoCatLog {
		rows--
	} else {
		lg = logs.New()
		bufferers = append(bufferers, lg)
	}

	draw := func() { termui.Render(bufferers...) }

	resize := func(width, height int) {
		ts.Resize(
			width/2,     // width
			height/rows, // height
			0,           // x
			0,           // y
		)

		a.Resize(
			width/2,         // width
			height/(rows*4), // height
			width/2,         // x
			0,               // y
		)

		st.Resize(
			width/2,         // width
			height/(rows*2), // height
			width/2,         // x
			height/(rows*4), // y
		)

		sz.Resize(
			width/2,             // width
			height/(rows*4),     // height
			width/2,             // x
			(height*3)/(rows*4), // y
		)

		r.Resize(
			width,       // width
			height/rows, // height
			0,           // x
			height/rows, // y
		)

		if !c.NoCatLog {
			lg.Resize(
				width,         // width
				height/rows,   // height
				0,             // x
				height*2/rows, // y
			)
		}
	}

	// get the terminal dimensions
	width := termui.TermWidth()
	if width == 0 {
		width = 80
	}

	height := termui.TermHeight()
	if height == 0 {
		height = 24
	}

	resize(width, height)
	draw()

	// exit with `q`, `Ctrl-c` and `Ctrl-d`
	for _, key := range []string{"/sys/kbd/q", "/sys/kbd/C-c", "/sys/kbd/C-d"} {
		termui.Handle(key, func(termui.Event) {
			termui.StopLoop()
		})
	}

	termui.Handle("/sys/wnd/resize", func(e termui.Event) {
		data := e.Data.(termui.EvtWnd)
		termui.Clear()
		resize(data.Width, data.Height)
		draw()
	})

	var handlers []internal.Handler
	for _, b := range bufferers {
		if h, ok := b.(internal.Handler); ok {
			handlers = append(handlers, h)
		}
	}

	go process(rdr, render(draw, handlers...))
	termui.Loop()

	return nil
}

func render(draw func(), handlers ...internal.Handler) chan<- clf.Entry {
	ch := make(chan clf.Entry)
	const ttl = 250 * time.Millisecond
	go func() {
		timer := time.NewTimer(ttl)
		for {
			select {
			case entry := <-ch:
				for _, h := range handlers {
					h.Handle(entry)
				}
			case <-timer.C:
				// rerender even if there is no new data so that stats continue
				// to be updated (e.g. ttls may expire)
				for _, h := range handlers {
					if r, ok := h.(internal.Renderer); ok {
						r.Render()
					}
				}
			}

			draw()
			timer.Reset(ttl)
		}
	}()
	return ch
}

func process(f io.Reader, ch chan<- clf.Entry) {
	scanner := bufio.NewScanner(f)

	for scanner.Scan() {
		entry, err := clf.ParseEntry(scanner.Bytes())
		if err != nil {
			continue
		}

		ch <- entry
	}
}
