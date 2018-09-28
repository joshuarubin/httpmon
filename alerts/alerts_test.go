package alerts

import (
	"testing"
	"time"

	"jrubin.io/httpmon/clf"
)

func TestNew(t *testing.T) {
	a := New(10, 2*time.Minute)
	if a == nil {
		t.Error("expected New not to return nil")
	}

	defer func() {
		if r := recover(); r == nil {
			t.Error("expected New to panic")
		}
	}()

	New(10, 0)
}

func TestAlerting(t *testing.T) {
	const rate = 2 * time.Second
	const lim = 10

	t.Run("NoAlert", func(t *testing.T) {
		t.Parallel()

		a := New(lim, rate)
		a.alertCh = make(chan msg)

		// should not produce an error at this rate

		go func() {
			var i int
			for range time.NewTicker(rate / (lim + 1)).C {
				if i > 10 {
					break
				}
				a.Handle(clf.Entry{})
				i++
			}
		}()

		timer := time.NewTimer(5 * time.Second)

		select {
		case <-a.alertCh:
			t.Error("unexpected message")
		case <-timer.C:
		}
	})

	t.Run("Alert", func(t *testing.T) {
		t.Parallel()

		a := New(lim, rate)
		a.alertCh = make(chan msg)

		// SHOULD produce an error and recovery at this rate

		go func() {
			for i := 0; i < 50; i++ {
				a.Handle(clf.Entry{})
			}
		}()

		timer := time.NewTimer(5 * time.Second)

		var n int
	LOOP:
		for {
			select {
			case m := <-a.alertCh:
				if n == 0 && m != alert || n == 1 && m != recovery {
					t.Errorf("unexpected msg(%d): %v", n, m)
				}
				if n == 1 {
					break LOOP
				}
				n++
			case <-timer.C:
				t.Error("timed out waiting for messages")
				break LOOP
			}
		}
	})
}
