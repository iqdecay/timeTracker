package main

import (
	"fmt"
	"github.com/andlabs/ui"
	"time"
)

type History []Session

type Session struct {
	Begin    time.Time
	End      time.Time
	Duration time.Duration
	Commits  int
}

type Project struct {
	Name        string
	Description string
	Duration    time.Duration
	History     History
	Commits     int
	Id          int
}

func main() {
	// Send and receive times for tracking
	beginTimes := make(chan time.Time)
	endTimes := make(chan time.Time)
	durations := make(chan time.Duration)
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
			durations <- duration
		}
	}()
	err := ui.Main(func() {
		box := ui.NewVerticalBox()
		button := ui.NewButton("Play")
		box.Append(button, true)
		window := ui.NewWindow("Hello", 400, 200, false)
		window.SetMargined(true)
		window.SetChild(box)
		button.OnClicked(func(b *ui.Button) {
			if b.Text() == "Play" {
				b.SetText("Pause")
				beginTimes <- time.Now()
				window.Destroy()
				window = ui.NewWindow("window2", 400, 200, true)
				newBox := ui.NewVerticalBox()
				newLabel := ui.NewLabel("welcome to new window")
				radioButtons := ui.NewRadioButtons()
				radioButtons.Append("choice 1")
				radioButtons.Append("choice 2")
				radioButtons.SetSelected(1)
				radioButtons.OnSelected(func(r *ui.RadioButtons) {
					fmt.Println("Pressed button : ", r.Selected())
				})
				separator := ui.NewHorizontalSeparator()
				newBox.Append(newLabel, true)
				newBox.Append(separator, true)
				newBox.Append(radioButtons, true)
				window.SetChild(newBox)
				window.OnClosing(func(*ui.Window) bool {
					ui.Quit()
					return true
				})
				window.Show()

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
