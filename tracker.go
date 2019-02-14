package main

import (
	"encoding/json"
	"fmt"
	"github.com/andlabs/ui"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"time"
)

const filename = "projects.json"
const durationFormat = "15:04:05"
const dateFormat = "Mon 01/02/06 03:04"

var (
	attrstr    *ui.AttributedString
	fontButton *ui.FontButton
	alignment  *ui.Combobox
)

type History []Session

type Session struct {
	Begin     time.Time
	End       time.Time
	Duration  time.Duration
	Commits   int
	ProjectId int
}

type Project struct {
	Name        string        `json:"name"`
	Description string        `json:"description"`
	Created     time.Time     `json:"created"`
	Duration    time.Duration `json:"duration"`
	History     History       `json:"history-list"`
	Commits     int           `json:"commits"`
	Id          int           `json:"unique-id"`
}

type ProjectList struct {
	MaxId int             `json:"max-id"`
	List  map[int]Project `json:"project-list"`
}

func (p *ProjectList) save() error {
	data, err := json.MarshalIndent(p, "", "	")
	if err != nil {
		log.Fatalf("error during json marshaling : %s", err)
	}
	return ioutil.WriteFile(filename, []byte(data), 0600)
}

func loadProjects() ProjectList {
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
	p.Duration += s.Duration
	p.Commits += s.Commits
	p.History = append(p.History, s)
}

func initSelectGUI() {
	// Setup the project selection combobox
	projects := loadProjects()
	var ids []int
	box := ui.NewVerticalBox()
	box.SetPadded(true)
	combobox := ui.NewCombobox()
	for _, v := range projects.List {
		combobox.Append(v.Name)
		ids = append(ids, v.Id)
	}
	combobox.SetSelected(0)
	box.Append(combobox, true)

	// Setup the window
	window := ui.NewWindow("Select a project or create a new one", 800, 400, false)
	window.SetChild(box)
	window.OnClosing(func(*ui.Window) bool {
		ui.Quit()
		return true
	})
	window.Show()

	// Add a select button
	selectButton := ui.NewButton("Work on this project")
	box.Append(selectButton, true)
	box.Append(ui.NewHorizontalSeparator(), false)
	selectButton.OnClicked(func(button *ui.Button) {
		selectedIndex := combobox.Selected()
		if selectedIndex == -1 {
			ui.MsgBox(window, "Error", "Please choose a project or create a new one")
			return
		}
		selectedId := ids[selectedIndex]
		window.Destroy()
		workonProject(selectedId)
	})

	// Add a create button
	createButton := ui.NewButton(" \n\n\n Or create a new project \n\n\n")
	createButton.OnClicked(func(button *ui.Button) {
		window.Destroy()
		ui.Main(initCreateGUI)
	})
	box.Append(createButton, false)
}

func initCreateGUI() {
	// Setup the creation form
	form := ui.NewForm()
	box := ui.NewHorizontalBox()
	form.Append("\n", box, false)
	form.SetPadded(true)
	titleEntry := ui.NewEntry()
	form.Append("Enter project name", titleEntry, false)
	descriptionEntry := ui.NewMultilineEntry()
	form.Append("Enter the description of the project", descriptionEntry, true)
	button := ui.NewButton("\n\n\n\n Save this project \n\n\n\n")
	form.Append("", button, false)

	// Setup the window
	window := ui.NewWindow("Create a project", 800, 400, true)
	window.SetChild(form)
	window.OnClosing(func(*ui.Window) bool {
		ui.Quit()
		return true
	})
	window.Show()

	// Setup the creation via the confirmation button
	projects := loadProjects()
	id := projects.MaxId + 1
	button.OnClicked(func(b *ui.Button) {
		var history History
		var duration time.Duration
		title := titleEntry.Text()
		description := descriptionEntry.Text()
		// Title must not be empty
		if title == "" {
			ui.MsgBox(window, "Error", "Please provide a non-empty title !")
			return
		}
		project := Project{title, description, time.Now(), duration, history, 0, id}
		projects.List[id] = project
		projects.MaxId = id

		projects.save()
		window.Destroy()
		fmt.Printf("Working on project : %d \n", id)
		workonProject(id)
	})
}

func workonProject(id int) {
	// Send and receive times for tracking
	beginTimes := make(chan time.Time)
	endTimes := make(chan time.Time)

	projects := loadProjects()
	project := projects.List[id]
	box := ui.NewHorizontalBox()
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

	// Add history tabular display
	table, _ := generateTable(project)
	box.Append(table, true)
	window := ui.NewWindow("Hello", 800, 400, false)
	window.SetMargined(true)
	window.SetChild(box)

	// Quit the app when the window is closed
	window.OnClosing(func(*ui.Window) bool {
		ui.Quit()
		return true
	})


	// Get session duration and update project with new session
	go func() {
		for {
			beginTime := <-beginTimes
			endTime := <-endTimes
			duration := endTime.Sub(beginTime)
			session := Session{beginTime, endTime, duration, 0, id}
			project.Add(session)
			fmt.Printf("Project n° %d was updated with a session of %s \n", id, duration)
			projects.List[id] = project
			projects.save()
		}
	}()

	window.Show()


}

type tabHandler struct {
	history History
	rows    int
}

func newTabHandler(h History) *tabHandler {
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
	l := 3
	types := make([]ui.TableValue, l)
	types[0] = ui.TableString("")
	types[1] = ui.TableString("")
	types[2] = ui.TableString("")
	return types
}

func (t tabHandler) NumRows(m *ui.TableModel) int {
	return t.rows
}

func (t tabHandler) CellValue(m *ui.TableModel, row, column int) ui.TableValue {
	switch column {
	case 0:
		return ui.TableString(t.history[row].Begin.Format(dateFormat))
	case 1:
		return ui.TableString(t.history[row].Duration.String())
	case 2:
		if t.history[row].Commits == 0 {
			return ui.TableString("None")
		} else {
			return ui.TableString(strconv.Itoa(t.history[row].Commits))

		}
	}
	return ui.TableString("error")
}

func (t *tabHandler) SetCellValue(m *ui.TableModel, row, column int, value ui.TableValue) {
	var err error
	switch column {
	case 0:
		t.history[row].Begin, err = time.Parse(dateFormat, string(value.(ui.TableString)))
		if err != nil {
			panic(err)
		}
	case 1:
		t.history[row].Duration, err = time.ParseDuration(string(value.(ui.TableString)))
		if err != nil {
			panic(err)
		}
	case 2:
		t.history[row].Commits = int(value.(ui.TableInt))
	}
}

func generateTable(project Project) (*ui.Table, ui.TableModelHandler) {
	handler := newTabHandler(project.History)
	tabModel := ui.NewTableModel(handler)
	params := ui.TableParams{Model: tabModel, RowBackgroundColorModelColumn: -1}
	table := ui.NewTable(&params)
	table.AppendTextColumn("Date", 0, ui.TableModelColumnNeverEditable, nil)
	table.AppendTextColumn("Duration", 1, ui.TableModelColumnNeverEditable, nil)
	table.AppendTextColumn("Commits made", 2, ui.TableModelColumnNeverEditable, nil)
	return table, handler
}

func main() {
	ui.Main(initSelectGUI)
}
