package main

import (
	"encoding/json"
	"fmt"
	"github.com/andlabs/ui"
	"io/ioutil"
	"log"
	"os"
	"time"
    "strconv"
)

const filename = "projects.json"
const durationFormat = "15:04:05"
const dateFormat = "Mon 01/02/00 03:04"

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
	sessions := make(chan Session)
	var selectedProject = make(chan int)
	// Update project with new session
	go func() {
		for {
			session := <-sessions
			id := session.ProjectId
			projects := loadProjects()
			duration := session.Duration.Round(time.Second)
			fmt.Printf("Project n° %d was updated with a session of %s \n", id, duration)
			project := projects.List[id]
			project.Add(session)
			projects.List[id] = project
			projects.save()
		}
	}()

	// Get session duration
	go func() {
		for {
			projectId := <-selectedProject
			beginTime := <-beginTimes
			endTime := <-endTimes
			duration := endTime.Sub(beginTime)
			session := Session{beginTime, endTime, duration, 0, projectId}
			sessions <- session
		}
	}()
	selectedProject <- id
	box := ui.NewVerticalBox()
	button := ui.NewButton("Play")
	box.Append(button, true)
	window := ui.NewWindow("Hello", 400, 200, false)
	window.SetMargined(true)
	window.SetChild(box)
	window.OnClosing(func(*ui.Window) bool {
		ui.Quit()
		return true
	})
	button.OnClicked(func(b *ui.Button) {
		if b.Text() == "Play" {
			b.SetText("Pause")
			beginTimes <- time.Now()
		} else {
			b.SetText("Play")
			endTimes <- time.Now()
			selectedProject <- id

		}
	})
	window.Show()

}

type tabHandler struct {
	history History
	rows int
}

func newTabHandler(h History) *tabHandler {
	m := new(tabHandler)
	m.history = h
	m.rows = len(h)
	return m
}

func (t tabHandler) ColumnTypes(m *ui.TableModel) []ui.TableValue {
	l := 3
	types := make([]ui.TableValue, l)
	for i := 0; i < l; i++ {
		types[i] = ui.TableString("")
	}
	return types
}

func (t tabHandler) NumRows(m *ui.TableModel) int {
	return t.rows
}

func (t tabHandler) CellValue(m *ui.TableModel, row, column int) ui.TableValue {
	switch column {
	case 0 :
		return ui.TableString(t.history[row].Begin.Format(dateFormat))
	case 1 :
		return ui.TableString(t.history[row].Duration)
	case 2 :
		return ui.TableInt(t.history[row].Commits)
	}
}

func (t *tabHandler) SetCellValue(m *ui.TableModel, row, column int, value ui.TableValue) {
	var err error
	switch column {
	case 0 :
		t.history[row].Begin, err = time.Parse(dateFormat, string(value.(ui.TableString)))
		if err != nil {
			panic(err)
		}
	case 1 :
		t.history[row].Duration, err = time.ParseDuration(string(value.(ui.TableString)))
		if err != nil {
			panic(err)
		}
	case 2 :
		t.history[row].Commits = int(value.(ui.TableInt))
	}
}


func (areaHandler) Draw(a *ui.Area, p *ui.AreaDrawParams) {
	tl := ui.DrawNewTextLayout(&ui.DrawTextLayoutParams{
		String:      attrstr,
		DefaultFont: fontButton.Font(),
		Width:       p.AreaWidth,
		Align:       ui.DrawTextAlign(alignment.Selected()),
	})
	defer tl.Free()
	p.Context.Text(tl, 0, 0)
}

func (areaHandler) MouseEvent(a *ui.Area, me *ui.AreaMouseEvent) {
	// do nothing
}

func (areaHandler) MouseCrossed(a *ui.AreaMouseEvent, left bool) {
	// do nothing
}

func (areaHandler) DragBroken(a *ui.Area) {
	// do nothing
}

func (areaHandler) KeyEvent(a *ui.Area, ke *ui.AreaKeyEvent) (handled bool) {
	// reject all keys
	return false
}
func initTable() {
	handler := newTabHandler()
	tabModel := ui.NewTableModel(handler)
	params := ui.TableParams{Model: tabModel, RowBackgroundColorModelColumn: -1}
	table := ui.NewTable(&params)
	table.AppendTextColumn("", 0, ui.TableModelColumnNeverEditable, nil)
	table.AppendTextColumn("", 1, ui.TableModelColumnNeverEditable, nil)
	box := ui.NewVerticalBox()
	box.Append(table, true)
	window := ui.NewWindow("test", 800, 400, false)
	window.SetChild(box)
	window.Show()
	window.OnClosing(func(*ui.Window) bool {
		ui.Quit()
		return true
	})
}

func main() {
	ui.Main(initTable)
	//ui.Main(initSelectGUI)
}
