package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strconv"
	"time"
)

const (
	EnvVarDB = "TASKD_DB"
)

func main() {
	args := os.Args
	if len(args) < 2 || args[1] == "help" {
		fmt.Println("taskd new <label> <description>")
		fmt.Println("taskd done <id>")
		fmt.Println("taskd close <id>")
		fmt.Println("taskd open")
		fmt.Println("taskd list")
		fmt.Println("taskd old")
		fmt.Println("taskd details <id>")
		return
	}

	mgr := newTaskMgr(os.Getenv(EnvVarDB))

	switch args[1] {
	case "new":
		lbl, desc := args[2], args[3]
		mgr.newTask(lbl, desc)
	case "done":
		id := args[2]
		idInt, err := strconv.Atoi(id)
		pie(err)
		mgr.mark(idInt, "DONE")
	case "close":
		id := args[2]
		idInt, err := strconv.Atoi(id)
		pie(err)
		mgr.mark(idInt, "CLOSED")
	case "list":
		mgr.printTasks(func(t task) bool { return true })
	case "old":
		mgr.printTasks(func(t task) bool { return t.Done })
	case "open":
		mgr.printTasks(func(t task) bool { return !t.Done })
	case "details":
		id := args[2]
		idInt, err := strconv.Atoi(id)
		pie(err)
		mgr.editDetails(idInt)
	default:
		pie(fmt.Errorf("unknown command"))
	}
}

type task struct {
	ID          int       // auto increment
	Date        time.Time // auto set
	Label       string
	Description string
	HasDetails  bool
	Done        bool
	DoneAt      time.Time
	Status      string
}

func (t task) String() string {
	s := ""

	details := ""
	if t.HasDetails {
		details = "(details available)"
	}

	s += fmt.Sprintf("%s - [%d] %s - %s ago %s\n", t.Status, t.ID, t.Label, time.Since(t.Date).Truncate(time.Minute), details)
	s += fmt.Sprintf(t.Description + "\n\n")
	return s
}

type taskMgr struct {
	dbFilepath string
	tasks      []task
}

func newTaskMgr(dbFilepath string) *taskMgr {
	mgr := &taskMgr{
		dbFilepath: dbFilepath,
		tasks:      make([]task, 0),
	}
	mgr.sync()
	return mgr
}

func (t *taskMgr) newTask(label, description string) {
	t.tasks = append(t.tasks, task{
		ID:          len(t.tasks) + 1,
		Date:        time.Now(),
		Label:       label,
		Description: description,
		Done:        false,
		Status:      "OPEN",
	})
	t.flush()
}

func (t *taskMgr) mark(id int, status string) {
	for i := range t.tasks {
		if t.tasks[i].ID == id {
			t.tasks[i].Status = status
			t.tasks[i].Done = true
			t.tasks[i].DoneAt = time.Now()
			break
		}
	}
	t.flush()
}

func (t *taskMgr) editDetails(id int) {
	found := false
	for i := range t.tasks {
		if t.tasks[i].ID == id {
			t.tasks[i].HasDetails = true
			found = true
			break
		}
	}
	if !found {
		panic("task not found")
	}

	filename := fmt.Sprintf("%s_%d.txt", t.dbFilepath, id)
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		f, err := os.Create(filename)
		pie(err)
		defer f.Close()
		_, err = f.Write([]byte(""))
		pie(err)
	}

	cmd := exec.Command("open", filename)
	pie(cmd.Start())

	t.flush()
}

func (t *taskMgr) printTasks(shouldPrint func(t task) bool) {
	ord := t.tasks
	sort.Slice(ord, func(i, j int) bool {
		return ord[i].Date.After(ord[j].Date)
	})
	for _, v := range ord {
		if !shouldPrint(v) {
			continue
		}
		fmt.Println(v)
	}
}

func (t *taskMgr) flush() {
	js, err := json.Marshal(t.tasks)
	pie(err)

	err = os.WriteFile(t.dbFilepath, js, 0644)
	pie(err)
}

func (t *taskMgr) sync() {
	if _, err := os.Stat(t.dbFilepath); os.IsNotExist(err) {
		f, err := os.Create(t.dbFilepath)
		pie(err)

		defer f.Close()
		_, err = f.Write([]byte("[]"))
		pie(err)
	}

	b, err := os.ReadFile(t.dbFilepath)
	if err != nil {
		panic(err)
	}
	pie(json.Unmarshal(b, &t.tasks))
}

func pie(err error) {
	if err != nil {
		panic(err)
	}
}
