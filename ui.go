package main

import (
	"fmt"
	"github.com/andlabs/ui"
	"time"
)

type Project struct {
	name     string
	duration time.Duration
	id       int
}

func startTracking(b *ui.Button) {
	//beginTime := time.Now()

}

func main() {
	// Send and receive begin times for tracking
	beginTimes := make(chan time.Time)
	endTimes := make(chan time.Time)
	// Send and receive projects ids
	selectedProject := make(chan int)
	go func() {
		project := <-selectedProject
		fmt.Println(project)
	}()
	go func() {
		for {
			beginTime := <-beginTimes
			endTime := <-endTimes
			duration := endTime.Sub(beginTime)
			fmt.Println(duration)
		}
	}()
	err := ui.Main(func() {
		box := ui.NewVerticalBox()
		button := ui.NewButton("Play")
		empty := ui.NewVerticalBox()
		box.Append(button, false)
		box.Append(empty, true)
		window := ui.NewWindow("Hello", 400, 200, false)
		window.SetMargined(true)
		window.SetChild(box)
		button.OnClicked(func(b *ui.Button) {
			if b.Text() == "Play" {
				b.SetText("Pause")
				beginTimes <- time.Now()

			} else {
				b.SetText("Play")
				endTimes <- time.Now()
			}

		})
		window.OnClosing(func(*ui.Window) bool {
			ui.Quit()
			return true
		})
		window.Show()
	})
	if err != nil {
		panic(err)
	}

}
