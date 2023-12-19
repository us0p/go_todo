package main

import (
	"bytes"
	"database/sql"
	"fmt"
	"go_todo/database"
	"io"
	"os"
	"strconv"
	"testing"
)

func TestInclude (t *testing.T) {
    set := []string{"asdf", "fdsa", "fads", "afsd"}

    t.Run("include should return false if the provided string isn't present in the set", func (t *testing.T) {
        isPresent := Include(set, "dsfa")

        if isPresent {
            t.Error("isPresent should be false")
        }
    })

    t.Run("include should return true as the proved string is present in the set", func (t *testing.T) {
        isPresent := Include(set, "asdf")

        if !isPresent {
            t.Error("isPresent should be true")
        }
    })
}

func TestGetOptionValue(t *testing.T) {
    allowedParameters := []string{"-name", "-completed"}

    t.Run("Should fail if there's parameters not allowed in the parameter list", func (t *testing.T) {
        args := []string{"-name", "test", "-asdf", "fdsa"}

        _, err := GetOptionValue(args, allowedParameters)

        if err == nil {
            t.Fatalf("Should have failed with string 'Parameter -asdf not recognized', got %s\n", err)
        }
    })

    t.Run("Should return a map mapping the provided options to their respective values", func(t *testing.T) {
        args := []string{"-name", "test"}

        maps, err := GetOptionValue(args, allowedParameters)

        if err != nil {
            t.Fatalf("failed with %s\n", err)
        }

        nameOp, ok := maps["-name"]

        if !ok {
            t.Error("maps should have the option -name")
        }

        if nameOp == "" {
            t.Errorf("nameOp should be %s, got empty string\n", "test")
        }
    })
}

func TestAddTask(t *testing.T) {
    db := getDBTransaction(t)
    defer db.Rollback()

    defaultUsageStr := fmt.Sprintf("%s\n", UsageStrAddTask)

    t.Run("Should print usage to stdout if there's not enough parameters", func (t *testing.T) {
        oldStdout, r, w := mockTearUpStdout(t)
        addTask(db, []string{"a"})
        got := mockTearDownStdout(t, oldStdout, r, w)

        if got !=  defaultUsageStr {
            t.Errorf("expected message to be %s, got %s", defaultUsageStr, got)
        }
    })

    t.Run("Should print not recognized parameter to stdout if any", func (t *testing.T) {
        oldStdout, r, w := mockTearUpStdout(t)
        addTask(db, []string{"a", "-asdf", "fdsa"})
        got := mockTearDownStdout(t, oldStdout, r, w)
        want := "Parameter -asdf not recognized\n"

        if got !=  want {
            t.Errorf("expected message to be %s, got %s", want, got)
        }
    })

    t.Run("Should print usage to stdout if there's no -name parameter", func (t *testing.T) {
        oldStdout, r, w := mockTearUpStdout(t)
        addTask(db, []string{"a", "-completed", "false"})
        got := mockTearDownStdout(t, oldStdout, r, w)

        if got !=  defaultUsageStr {
            t.Errorf("expected message to be %s, got %s", defaultUsageStr, got)
        }
    })

    t.Run("Should print usage to stdout if when passing -completed parameter with a not allowed string", func (t *testing.T) {
        oldStdout, r, w := mockTearUpStdout(t)
        addTask(db, []string{"a", "-name", "test", "-completed", "asdf"})
        got := mockTearDownStdout(t, oldStdout, r, w)

        if got !=  defaultUsageStr {
            t.Errorf("expected message to be %s, got %s", defaultUsageStr, got)
        }
    })

    t.Run("Should print the created ID to stdout if everything is ok", func (t *testing.T) {
        oldStdout, r, w := mockTearUpStdout(t)
        addTask(db, []string{"a", "-name", "test", "-completed", "true"})
        got := mockTearDownStdout(t, oldStdout, r, w)
        want := "Task with ID: 1 created!\n"

        if got !=  want {
            t.Errorf("expected message to be %s, got %s", want, got)
        }
        
        createdTask, err := database.ListTaskActionByID(db, uint(1))

        if err != nil {
            t.Errorf("error while fetching created task, %s\n", err)
        }

        if createdTask.ID != 1 {
            t.Errorf("expected created task to have the same ID of the received by the message, got %d\n", createdTask.ID)
        }
    })
}

func TestListTasks (t *testing.T) {
    db := getDBTransaction(t)
    defer db.Rollback()
    mockTask(t, db)
    _, err := database.AddTaskAction(db, database.AddTaskProp{
        Name: "Test 2",
        Completed: true,
    })

    if err != nil {
        t.Fatalf("error while mocking task, %s\n", err)
    }

    t.Run("Should list tasks in ascending order", func (t *testing.T) {
        oldStdout, r, w := mockTearUpStdout(t)
        listTasks(db, []string{})
        got := mockTearDownStdout(t, oldStdout, r, w)
        want := "1.[ ] - Test\n2.[x] - Test 2\n"

        if got != want {
            t.Errorf("expected:\n%s\ngot:\n%s\n", want, got)
        }
    })

    t.Run("Should print not recognized parameter to stdout if any", func (t *testing.T) {
        oldStdout, r, w := mockTearUpStdout(t)
        listTasks(db, []string{"-asdf"})
        got := mockTearDownStdout(t, oldStdout, r, w)
        want := "Parameter -asdf not recognized\n"
        if got != want {
            t.Error("expected:", want, "got:", got)
        }
    })

    t.Run("Should print usage to stdout when receiving a unexpected column parameter", func (t *testing.T) {
        oldStdout, r, w := mockTearUpStdout(t)
        listTasks(db, []string{"-sort", "asdf,desc"})
        got := mockTearDownStdout(t, oldStdout, r, w)
        if got != fmt.Sprintf("%s\n", DefaultListUsageStr) {
            t.Error("expected:", DefaultListUsageStr, "got:", got)
        }
    })

    t.Run("Should print usage to stdout when receiving a unexpected sorting parameter", func (t *testing.T) {
        oldStdout, r, w := mockTearUpStdout(t)
        listTasks(db, []string{"-sort", "name,asdf"})
        got := mockTearDownStdout(t, oldStdout, r, w)
        if got != fmt.Sprintf("%s\n", DefaultListUsageStr) {
            t.Error("expected:", DefaultListUsageStr, "got:", got)
        }
    })

    t.Run("Should list task in descending order", func (t *testing.T) {
        oldStdout, r, w := mockTearUpStdout(t)
        listTasks(db, []string{"-sort", "id,desc"})
        got := mockTearDownStdout(t, oldStdout, r, w)
        want := "2.[x] - Test 2\n1.[ ] - Test\n"
        if got != want {
            t.Error("expected:", want, "got:", got)
        }
    })

    t.Run("Should print usage to stdout when receiving a unexpected filtering parameter", func (t *testing.T) {
        oldStdout, r, w := mockTearUpStdout(t)
        listTasks(db, []string{"-completed", "asdf"})
        got := mockTearDownStdout(t, oldStdout, r, w)
        if got != fmt.Sprintf("%s\n", DefaultListUsageStr) {
            t.Error("expected:", DefaultListUsageStr, "got:", got)
        }
    })

    t.Run("Should list only the task with completed = false", func (t *testing.T) {
        oldStdout, r, w := mockTearUpStdout(t)
        listTasks(db, []string{"-completed", "false"})
        got := mockTearDownStdout(t, oldStdout, r, w)
        want := "1.[ ] - Test\n"
        if got != want {
            t.Error("expected:", want, "got:", got)
        }
    })
}

func TestDeleteTasks(t *testing.T) {
    db := getDBTransaction(t)
    defer db.Rollback()

    t.Run("Should print usage if it wasn't provided any task ID", func (t *testing.T) {
        oldStdout, r, w := mockTearUpStdout(t)
        deleteTasks(db, []string{"d"})
        got := mockTearDownStdout(t, oldStdout, r, w)

        if got != fmt.Sprintf("%s\n", DefaultDeleteUsageStr) {
            t.Error("should have printed usage, got:", got)
        }
    })

    t.Run("Should return an error if there's not a numeric character in the provided ID list", func (t *testing.T) {
        oldStdout, r, w := mockTearUpStdout(t)
        deleteTasks(db, []string{"d", "1", "2", ","})
        got := mockTearDownStdout(t, oldStdout, r, w)
        want := "Error: ',' isn't a numeric character\n"

        if got != want {
            t.Error("should have failed with:", want, "got:", got)
        }
    })


    t.Run("Should delete tasks from database", func (t *testing.T) {
        task1 := mockTask(t, db)
        task2 := mockTask(t, db)
        oldStdout, r, w := mockTearUpStdout(t)
        deleteTasks(db, []string{"d", "1", "2"})
        got := mockTearDownStdout(t, oldStdout, r, w)
        want := "Deleted 2 tasks.\n"

        if got != want {
            t.Error("should have printed:", want, "got:", got)
        }

        _, err := database.ListTaskActionByID(db, uint(task1.ID))

        if err == nil {
            t.Error("should have failed with ErrNoRow as the task should be deleted", err)
        }

        _, errTask2 := database.ListTaskActionByID(db, uint(task2.ID))

        if errTask2 == nil {
            t.Error("should have failed with ErrNoRow as the task should be deleted", errTask2)
        }
    })
}

func TestUpdateTask(t *testing.T) {
    db := getDBTransaction(t)
    defer db.Rollback()
    task := mockTask(t, db)

    t.Run("Should print usage if missing task ID to update", func (t *testing.T) {
        oldStdout, r, w := mockTearUpStdout(t)
        updateTask(db, []string{"u"})
        got := mockTearDownStdout(t, oldStdout, r, w)
        if got != fmt.Sprintf("%s\n", DefaultUpdateUsageStr) {
            t.Error("should have printed default usage string, got:", got)
        }
    })

    t.Run("Should print usage if provided ID isn't a numeric character", func (t *testing.T) {
        oldStdout, r, w := mockTearUpStdout(t)
        updateTask(db, []string{"u", "asdf"})
        got := mockTearDownStdout(t, oldStdout, r, w)
        if got != fmt.Sprintf("%s\n", DefaultUpdateUsageStr) {
            t.Error("should have printed default usage string, got:", got)
        }
    })

    t.Run("Should print unrecognized parameter if any", func (t *testing.T) {
        oldStdout, r, w := mockTearUpStdout(t)
        updateTask(db, []string{"u", "69", "-asdf"})
        got := mockTearDownStdout(t, oldStdout, r, w)
        want := "Parameter -asdf not recognized\n"
        if got != want {
            t.Error("should have printed:", want, "got:", got)
        }
    })

    t.Run("Should print usage if missing -name or -completed parameters", func (t *testing.T) {
        oldStdout, r, w := mockTearUpStdout(t)
        updateTask(db, []string{"u", strconv.Itoa(task.ID)})
        got := mockTearDownStdout(t, oldStdout, r, w)
        if got != fmt.Sprintf("%s\n", DefaultUpdateUsageStr) {
            t.Error("should have printed default usage string, got:", got)
        }
    })

    t.Run("Should print usage if value of parameter -completed is not as expected", func (t *testing.T) {
        oldStdout, r, w := mockTearUpStdout(t)
        updateTask(db, []string{"u", strconv.Itoa(task.ID), "-completed", "asdf"})
        got := mockTearDownStdout(t, oldStdout, r, w)
        if got != fmt.Sprintf("%s\n", DefaultUpdateUsageStr) {
            t.Error("should have printed default usage string, got:", got)
        }
    })

    t.Run("Should print error of unexisting task if the provided task ID wasn't present in the data set", func (t *testing.T) {
        oldStdout, r, w := mockTearUpStdout(t)
        updateTask(db, []string{"u", "69", "-name", "test"})
        got := mockTearDownStdout(t, oldStdout, r, w)
        want := "Error: Task doesn't exist\n"
        if got != want{
            t.Error("should have printed:", want, "got:", got)
        }
    })


    t.Run("Should update the provided task and print its ID if everything is ok", func (t *testing.T) {
        oldStdout, r, w := mockTearUpStdout(t)
        updateTask(db, []string{"u", strconv.Itoa(task.ID), "-completed", "true"})
        got := mockTearDownStdout(t, oldStdout, r, w)
        want := fmt.Sprintf("Task %d updated\n", task.ID)
        if got != want{
            t.Error("should have printed:", want, "got:", got)
        }

        updatedTask, err := database.ListTaskActionByID(db, uint(task.ID))

        if err != nil {
            t.Fatal("error while listing updated task", err)
        }

        if !updatedTask.Completed {
            t.Errorf("should have updated provided task completed status to true, got: %t\n", updatedTask.Completed)
        }
    })
}

func TestHelp (t *testing.T) {
    t.Run("Should print usage if there was no option provided", func (t *testing.T) {
        oldStdout, r, w := mockTearUpStdout(t)
        help([]string{"help"})
        got := mockTearDownStdout(t, oldStdout, r, w)
        want := fmt.Sprintf("%s\n", DefaultHelpUsageStr)

        if got != want {
            t.Errorf("expected: %s, got: %s", want, got)
        }
    })

    t.Run("Should print unexistent option if the provided option doesn't exist", func (t *testing.T) {
        oldStdout, r, w := mockTearUpStdout(t)
        help([]string{"help", "asdf"})
        got := mockTearDownStdout(t, oldStdout, r, w)
        want := "Option asdf not recognized\n"

        if got != want {
            t.Errorf("expected: %s, got: %s", want, got)
        }
    })

    t.Run("Should print the default usage of add option", func (t *testing.T) {
        oldStdout, r, w := mockTearUpStdout(t)
        help([]string{"help", "a"})
        got := mockTearDownStdout(t, oldStdout, r, w)
        want := fmt.Sprintf("%s\n", UsageStrAddTask)

        if got != want {
            t.Errorf("expected: %s, got: %s", want, got)
        }
    })

    t.Run("Should print the default usage of list option", func (t *testing.T) {
        oldStdout, r, w := mockTearUpStdout(t)
        help([]string{"help", "l"})
        got := mockTearDownStdout(t, oldStdout, r, w)
        want := fmt.Sprintf("%s\n", DefaultListUsageStr)

        if got != want {
            t.Errorf("expected: %s, got: %s", want, got)
        }
    })

    t.Run("Should print the default usage of delete option", func (t *testing.T) {
        oldStdout, r, w := mockTearUpStdout(t)
        help([]string{"help", "d"})
        got := mockTearDownStdout(t, oldStdout, r, w)
        want := fmt.Sprintf("%s\n", DefaultDeleteUsageStr)

        if got != want {
            t.Errorf("expected: %s, got: %s", want, got)
        }
    })

    t.Run("Should print the default usage of update option", func (t *testing.T) {
        oldStdout, r, w := mockTearUpStdout(t)
        help([]string{"help", "u"})
        got := mockTearDownStdout(t, oldStdout, r, w)
        want := fmt.Sprintf("%s\n", DefaultUpdateUsageStr)

        if got != want {
            t.Errorf("expected: %s, got: %s", want, got)
        }
    })
}

func mockTearUpStdout(t testing.TB) (oldStdout *os.File, r *os.File, w *os.File){
    t.Helper()
    // capturing the original stdout
    oldStdout = os.Stdout
    // r is a read file connected to the w write file, so bytes writen to w can be read from r
    r, w, err := os.Pipe()
    // changing os.Stdout to the new created write file
    os.Stdout = w

    if err != nil {
        t.Fatal("error while acquiring pair of files")
    }

    return oldStdout, r, w
}

func mockTearDownStdout(t testing.TB, oldStdout *os.File, r *os.File, w *os.File) string { 
    t.Helper()

    outputCopy := make(chan string)
    go func () {
        var buff bytes.Buffer
        // Will block until there are bytes available to be read from r
        io.Copy(&buff, r)
        // If anythin that writes to stdout is executed, we'll copy the data to our copy
        outputCopy <- buff.String()
    }()

    // If there's nothing to write to stdout, io.Copy will be able to recognize that the write end of the
    // file is closed and that no more data will be write.
    w.Close()
    os.Stdout = oldStdout

    return <-outputCopy
}

func getDBTransaction(t testing.TB) (*sql.Tx) {
    t.Helper()
    db, err := database.OpenDatabase("./")

    if err != nil {
        t.Fatalf("error while connecting to the database, %s\n", err)
    }
    defer db.Close()

    tx, err := db.Begin()

    if err != nil {
        t.Fatalf("error while acquiring transaction, %s\n", err)
    }

    return tx
}

func mockTask(t testing.TB, db *sql.Tx) database.Task {
    t.Helper()
    task, err := database.AddTaskAction(db, database.AddTaskProp{
        Name: "Test",
        Completed: false,
    })

    if err != nil {
        t.Fatalf("error while mocking task, %s\n", err)
    }

    return task
}
