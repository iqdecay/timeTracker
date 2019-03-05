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
  
  
 Or you can just use the provided binary by running :
 
 `./tracker`
 
## TODOS :
- [ ] Add confirmation for deletion
- [ ] Format time down to the minute
- [ ] Add the possibility for a todolist
- [ ] Don't save sessions < 1 min
- [ ] Display project summary
- [ ] Fix sessions ordering in display
- [ ] Change window size
- [ ] Add sound at regular intervals


## Specifications :
- Add a project
- Select a project among a list of added projects
- Start/stop tracking the time spent on a project
- Track the commits made during a session
- Display a stopwatch
- Log the time spent on each project
- Delete a project
- Track the commits on the project

