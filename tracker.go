package main

import (
	"encoding/json"
	"fmt"
	"github.com/andlabs/ui"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strconv"
	"time"
)

const filename = "projects.json"
const dateFormat = "Mon 01/02/06 15:04"
const gitDateFormat = "01/02/06 15:04:05"
const width = 600
const height = 300

var projects = loadProjects()

type Session struct {
	Begin     time.Time
	End       time.Time
	Duration  time.Duration
	ProjectId int
	Comment   string
	Commits   int
}

func (s *Session) getCommits() {
	// Get the number of commits made between Begin and End time

	id := s.ProjectId
	beginDate := s.Begin.Format(gitDateFormat)
	endDate := s.End.Format(gitDateFormat)

	gitCommand := exec.Command("git", "log", "--since", beginDate, "--until", endDate, "--pretty=oneline")
	gitCommand.Dir = projects.List[id].Dir // Run the command in the project directory
	output, err := gitCommand.Output()
	if err != nil {
		panic(err)
	}
	var commits int
	// Each line of commit ends with a newline
	for _, b := range output {
		if b == byte('\n') {
			commits ++
		}
	}
	fmt.Printf("%d commits made during this session \n", commits)
	s.Commits = commits
}

type History []Session

type Project struct {
	Name        string        `json:"name"`
	Description string        `json:"description"`
	Created     time.Time     `json:"created"`
	Duration    time.Duration `json:"duration"`
	History     History       `json:"history-list"`
	Id          int           `json:"unique-id"`
	LastComment string        `json:"last-comment"`
	Commits     int           `json:"commits"`
	Dir         string        `json:"working-directory"`
}

type ProjectList struct {
	MaxId int             `json:"max-id"`
	List  map[int]Project `json:"project-list"`
}

func (p *ProjectList) save() error {
	// Save the project list to memory using JSON marshaling
	data, err := json.MarshalIndent(p, "", "	")
	if err != nil {
		log.Fatalf("error during json marshaling : %s", err)
	}
	return ioutil.WriteFile(filename, []byte(data), 0600)
}

func loadProjects() ProjectList {
	// Load the project list from memory using JSON unmarshaling
	var projects ProjectList
	projects.List = make(map[int]Project)
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

func (p *Project) Add(s Session) {
	// Add the finished session to the project
	p.Duration += s.Duration
	p.History = append(p.History, s)
	p.Commits += s.Commits
	p.LastComment = s.Comment
}

type tabHandler struct {
	history History
	rows    int
}

func newTabHandler(h History) *tabHandler {
	// Create a table to display recent sessions first
	m := new(tabHandler)
	// We reverse the slice so recent sessions appear on top
	var opp int
	for i := len(h)/2 - 1; i >= 0; i-- {
		opp = len(h) - 1 - i
		h[i], h[opp] = h[opp], h[i]
	}
	m.history = h
	m.rows = len(h)
	return m
}

func (t tabHandler) ColumnTypes(m *ui.TableModel) []ui.TableValue {
	// Return the type of the data in each column
	l := 4
	types := make([]ui.TableValue, l)
	types[0] = ui.TableString("")
	types[1] = ui.TableString("")
	types[2] = ui.TableString("")
	types[3] = ui.TableString("")
	return types
}

func (t tabHandler) NumRows(m *ui.TableModel) int {
	return t.rows
}

func (t tabHandler) CellValue(m *ui.TableModel, row, column int) ui.TableValue {
	// Return value of cell in [row][column]
	switch column {
	case 0:
		// The date of the session
		return ui.TableString(t.history[row].Begin.Format(dateFormat))
	case 1:
		// Duration of the session
		return ui.TableString(t.history[row].Duration.String())
	case 2:
		// Number of commits in the session
		return ui.TableString(strconv.Itoa(t.history[row].Commits))
	case 3:
		// Comment on the session
		if t.history[row].Comment == "" {
			return ui.TableString("None")
		} else {
			return ui.TableString(t.history[row].Comment)
		}
	}
	return ui.TableString("error")
}

func (t *tabHandler) SetCellValue(m *ui.TableModel, row, column int, value ui.TableValue) {
	var err error
	switch column {
	case 0:
		// Convert string into a begin time
		t.history[row].Begin, err = time.Parse(dateFormat, string(value.(ui.TableString)))
		if err != nil {
			panic(err)
		}
	case 1:
		// Convert string into a duration
		t.history[row].Duration, err = time.ParseDuration(string(value.(ui.TableString)))
		if err != nil {
			panic(err)
		}
	case 2:
		t.history[row].Commits, _ = strconv.Atoi(string(value.(ui.TableString)))
	case 3:
		t.history[row].Comment = string(value.(ui.TableString))
	}
}

func generateTable(project Project) (*ui.Table, *tabHandler, *ui.TableModel) {
	// Generate the history table with the underlying tabHandler model
	handler := newTabHandler(project.History)
	tabModel := ui.NewTableModel(handler)
	params := ui.TableParams{Model: tabModel, RowBackgroundColorModelColumn: -1}
	table := ui.NewTable(&params)
	table.AppendTextColumn("Date", 0, ui.TableModelColumnNeverEditable, nil)
	table.AppendTextColumn("Duration", 1, ui.TableModelColumnNeverEditable, nil)
	table.AppendTextColumn("Commits", 2, ui.TableModelColumnNeverEditable, nil)
	table.AppendTextColumn("Comment", 3, ui.TableModelColumnNeverEditable, nil)
	return table, handler, tabModel
}

func initSelectGUI() {
	// GUI where the user either chooses an existing project or creates a new one

	// SELECTION
	// Add a selection combobox
	var ids []int
	combobox := ui.NewCombobox()
	for _, v := range projects.List {
		combobox.Append(v.Name)
		ids = append(ids, v.Id)
	}
	combobox.SetSelected(0)
	// Fit it nicely into a box
	box := ui.NewVerticalBox()
	box.SetPadded(true)
	box.Append(combobox, true)

	// Setup the window
	window := ui.NewWindow("Select a project or create a new one", width, height, false)
	window.SetChild(box)
	// Quit the app when the window is closed
	window.OnClosing(func(*ui.Window) bool {
		ui.Quit()
		return true
	})
	window.Show()

	// Add a select button
	selectButton := ui.NewButton("Work on this project")
	selectButton.OnClicked(func(button *ui.Button) {
		// Get the selected id from combobox
		selectedIndex := combobox.Selected()
		if selectedIndex == -1 {
			ui.MsgBox(window, "Error", "Please choose a project or create a new one")
			return
		}
		selectedId := ids[selectedIndex]
		window.Destroy()
		// Display the GUI for working on a project
		workonProject(selectedId)
	})
	// Fit it nicely into a box
	selectBox := ui.NewHorizontalBox()
	selectBox.Append(ui.NewHorizontalBox(), true)
	selectBox.Append(selectButton, true)
	selectBox.Append(ui.NewHorizontalBox(), true)
	box.Append(selectBox, true)

	// Add a create button
	createButton := ui.NewButton(" \n\n\n Or create a new project \n\n\n")
	createButton.OnClicked(func(button *ui.Button) {
		window.Destroy()
		ui.Main(initCreateGUI)
	})
	// Fit it nicely into a box
	createBox := ui.NewHorizontalBox()
	createBox.Append(ui.NewHorizontalBox(), true)
	createBox.Append(createButton, true)
	createBox.Append(ui.NewHorizontalBox(), true)
	box.Append(createBox, true)
}

func initCreateGUI() {
	// GUI to create a new project
	window := ui.NewWindow("Create a project", width, height, true)

	// Setup the creation form
	form := ui.NewForm()
	form.SetPadded(true)
	// Add a return button
	returnButton := ui.NewButton("Return to project list")
	returnButton.OnClicked(func(b *ui.Button) {
		window.Destroy()
		ui.Main(initSelectGUI)
	})
	// Fit it nicely into a box
	topBox := ui.NewHorizontalBox()
	topBox.Append(returnButton, false)
	topBox.Append(ui.NewHorizontalBox(), true)
	form.Append("", topBox, false)
	// Create the input in the form
	titleEntry := ui.NewEntry()
	form.Append("Enter project name", titleEntry, false)
	descriptionEntry := ui.NewMultilineEntry()
	form.Append("Enter the description of the project", descriptionEntry, false)
	dirEntry := ui.NewEntry()
	form.Append("Enter the project directory, in absolute path ", dirEntry, false)
	button := ui.NewButton("\n\n\n\n Save this project \n\n\n\n")
	form.Append("", button, false)

	// Setup the window
	window.SetChild(form)
	// Quit the app when the window is closed
	window.OnClosing(func(*ui.Window) bool {
		ui.Quit()
		return true
	})
	window.Show()

	// Setup the creation via the confirmation button
	id := projects.MaxId + 1
	button.OnClicked(func(b *ui.Button) {
		// Get data and register it as new project
		var history History
		var duration time.Duration
		title := titleEntry.Text()
		description := descriptionEntry.Text()
		dir := dirEntry.Text()
		// Title must not be empty
		if title == "" {
			ui.MsgBox(window, "Error", "Please provide a non-empty title !")
			return
		}
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			ui.MsgBox(window, "Error", "The provided path is incorrect")
			return
		}
		gitStatus := exec.Command("git", "status")
		gitStatus.Dir = dir
		_, err := gitStatus.Output()
		if err != nil {
			ui.MsgBox(window, "Error", "The specified directory is not a Git directory."+
				"\nPlease run 'git init .' ")
			return
		}

		project := Project{title, description, time.Now(),
			duration, history, id, "Project created", 0, dir}
		projects.List[id] = project
		projects.MaxId = id
		projects.save()
		window.Destroy()
		fmt.Printf("Working on project : %d \n", id)
		workonProject(id)
		return
	})
}

func workonProject(id int) {
	// GUI when working on a project

	// Send and receive times for tracking
	beginTimes := make(chan time.Time)
	endTimes := make(chan time.Time)
	comments := make(chan string)

	project := projects.List[id]

	// Initialize window

	box := ui.NewHorizontalBox()
	windowTitle := fmt.Sprintf("Project : %s", project.Name)
	window := ui.NewWindow(windowTitle, width, height, false)
	window.SetMargined(true)
	window.SetChild(box)

	// Add play/pause button
	button := ui.NewButton("Start")
	box.Append(button, true)

	// Alternate between start and stop button
	button.OnClicked(func(b *ui.Button) {
		if b.Text() == "Start" {
			b.SetText("Stop")
			beginTimes <- time.Now()
		} else {
			b.SetText("Start")
			endTimes <- time.Now()
		}
	})

	// Add a return button
	returnButton := ui.NewButton("Return to project list")
	returnButton.OnClicked(func(b *ui.Button) {

		window.Destroy()
		ui.Main(initSelectGUI)
	})
	// Right part of the ui
	rightbox := ui.NewVerticalBox()
	rightbox.Append(returnButton, false)

	// Add history tabular display

	table, handler, model := generateTable(project)
	rightbox.Append(table, true)
	box.Append(rightbox, true)

	// Add a delete button
	deleteButton := ui.NewButton("Delete this project")
	deleteButton.OnClicked(func(b *ui.Button) {
		delete(projects.List, id)
		ui.MsgBox(window, "Confirmation", "Project was correctly deleted")
		window.Destroy()
		ui.Main(initSelectGUI)
	})
	rightbox.Append(deleteButton, false)

	// Quit the app when the window is closed
	window.OnClosing(func(*ui.Window) bool {
		ui.Quit()
		return true
	})
	window.Show()
	go func() {

		// Generate a form for commenting the session
		form := ui.NewForm()
		form.Append("\n", ui.NewHorizontalBox(), false)
		commentEntry := ui.NewEntry()
		form.Append("Enter comment for the session ", commentEntry, true)
		submitButton := ui.NewButton("\n\n Save this session \n\n")
		// The buttons submits the comment
		submitButton.OnClicked(func(button *ui.Button) {
			comments <- commentEntry.Text()
		})
		form.Append("", submitButton, false)

		// Get session duration and update project with new session
		for {
			beginTime := <-beginTimes
			ended := false
			go func() {
				// Update the timer display every second until the button is pressed again
				for !ended {
					// Queue the event because we are in a goroutine
					ui.QueueMain(func() {
						button.SetText(time.Since(beginTime).Truncate(time.Second).String())
					})
					time.Sleep(1 * time.Second)
				}
			}()
			endTime := <-endTimes
			ended = true

			// Display the form
			ui.QueueMain(func() {
				window.SetTitle("Enter comment about the session")
				window.SetChild(form)
			})

			// Hold until the button is pressed
			comment := <-comments

			// Clear the form for later use
			commentEntry.SetText("")

			// Go back to tracking
			ui.QueueMain(func() {
				window.SetTitle(windowTitle)
				window.SetChild(box)
			})

			// Add the new session to project
			duration := endTime.Sub(beginTime)
			fmt.Println(comment)
			session := Session{beginTime, endTime, duration, id, comment, 0}
			session.getCommits()
			project.Add(session)
			fmt.Printf("Project n° %d was updated with a session of %s \n", id, duration)
			projects.List[id] = project
			projects.save()

			// Update the history display with the new session
			handler.rows += 1
			previousHistory := handler.history
			previousHistory = append(previousHistory, Session{})
			copy(previousHistory[1:], previousHistory[0:])
			previousHistory[0] = session
			handler.history = previousHistory
			model.RowInserted(0)
		}
	}()
}

func main() {
	ui.OnShouldQuit(func() bool {
		return true
	})
	ui.Main(initSelectGUI)
}
