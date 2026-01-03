package ui

import (
	"fmt"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
)

type timer struct {
	showMillis bool
	startTime  time.Time
	mtx        *sync.Mutex
	text       *canvas.Text
	stop       chan struct{}
}

func newTimer(showMillis bool) *timer {
	initText := "00:00"
	if showMillis {
		initText = "00:00.000"
	}
	return &timer{
		showMillis: showMillis,
		startTime:  time.Time{},
		mtx:        &sync.Mutex{},
		text:       canvas.NewText(initText, nil),
		stop:       make(chan struct{}),
	}
}

func (t *timer) Set(start time.Time) {
	t.mtx.Lock()
	t.startTime = start
	t.mtx.Unlock()
}

func (t *timer) Stop() {
	close(t.stop)
}

func (t *timer) Go(waitForStart chan struct{}) {
	d := time.Second
	if t.showMillis {
		d = 64 * time.Millisecond
	}

	go func() {
		<-waitForStart
		for range time.Tick(d) {
			select {
			case <-t.stop:
				return
			default:
			}
			fyne.Do(func() {
				t.mtx.Lock()
				elapsed := time.Since(t.startTime)
				minutes := int(elapsed.Minutes())
				seconds := int(elapsed.Seconds()) % 60
				if t.showMillis {
					millis := int(elapsed.Milliseconds()) % 1000
					t.text.Text = fmt.Sprintf("%02d:%02d.%03d", minutes, seconds, millis)
				} else {
					t.text.Text = fmt.Sprintf("%02d:%02d", minutes, seconds)
				}
				t.text.Refresh()
				t.mtx.Unlock()
			})
		}
	}()
}
