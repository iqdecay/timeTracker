# A time tracker

The project is to develop a time tracker so I can track time 
spent on a development project.

I will use [andlabs/ui](https://github.com/andlabs/ui) to have
a nice GUI that is compatible with all platforms, and not terminal
based.


## Usage :

- First, install dependencies :

  `go get github.com/andlabs/ui`
- Run the program : 
 
  `go run tracker.go`



## Todos :
- [x] GUI for adding project
- [x] GUI for selecting project
- [x] Generate tracking project table
- [x] GUI when tracking project
- [x] Update session on the fly
- [x] Add comment at end of session
- [x] Fix GUI exit when a new project was created
- [x] Add timer display
- [x] Add return button
- [ ] Add keyboard support 
- [ ] Add return button when creating a project
- [ ] Allow deleting project

## Specifications :
- Add a project
- Select a project among a list of added projects
- Start/stop tracking the time on the click of a button
- Display a stopwatch
- Emit a sound every 30 mins for instance
- Log the time spent on each project
- Display a summary of all projects
- Delete or modify a project

