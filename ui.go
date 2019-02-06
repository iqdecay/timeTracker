package main

import (
	"encoding/json"
	"fmt"
	"github.com/andlabs/ui"
	"io/ioutil"
	"log"
	"os"
	"time"
)

const filename = "projects.json"

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

type ProjectList []Project

func (p *ProjectList) save() error {
	data, err := json.MarshalIndent(p, "", "	")
	if err != nil {
		log.Fatalf("error during json marshaling : %s", err)
	}
	return ioutil.WriteFile(filename, []byte(data), 0600)
}

func loadProjects() ProjectList {
	var projects ProjectList
	// if the file doesn't exist, the project list is empty
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return projects
	} else {
		// process data
		data, err := ioutil.ReadFile(filename)
		if err != nil {
			log.Fatalf("error reading projects list : #{err}")
		}
		if err := json.Unmarshal(data, &projects); err != nil {
			log.Fatalf("error during json unmarshaling: %s", err)
		}
		return projects
	}
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
			duration := <-durations
			fmt.Println(duration)
		}
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
