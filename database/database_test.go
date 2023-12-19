package database

import (
	"database/sql"
	"testing"
)

func TestOpenDatabase(t *testing.T) {
    db, err := OpenDatabase("../")

    if err != nil {
        t.Fatalf("error while opening database connection, %s\n", err)
    }

    pingErr := db.Ping()

    if pingErr != nil {
        t.Error("error with database connection, couldn't ping database")
    }
}

func TestAddTaskAction(t *testing.T) {
    tx := getDBTransaction(t)

    defer tx.Rollback()

    task, err := AddTaskAction(tx, AddTaskProp{"Practice Go", true})

    if err != nil {
        t.Fatalf("error while adding the task to the database, %s\n", err)
    }

    if task.Name != "Practice Go" {
        t.Errorf("expected task name to be 'Practice Go', got %s\n", task.Name)
    }

    if !task.Completed {
        t.Error("expected task status to be completed got to-do")
    }

    if task.ID != 1 {
        t.Errorf("Expected task id to be 1, got %d\n", task.ID)
    }
}

func TestUpdateTaskAction(t *testing.T) {
    tx := getDBTransaction(t)
    
    defer tx.Rollback()

    t.Run("Fails if task id doesn't exist", func(t *testing.T) {
        payload := UpdateTaskProp{}
        name := "Test"
        payload.Name = &name

        _, err := UpdateTaskAction(tx, 69, payload)

        if err == nil {
            t.Fatalf("should have failed with unexistent task")
        }

        if err.Error() != "Task doesn't exist" {
            t.Errorf("expected error: 'Task doesn't exist', got %s\n", err)
        }
    })

    t.Run("Testing updating task name", func(t *testing.T) {
        task := mockTask(t, tx)

        payload := UpdateTaskProp{}
        name := "Updated task"
        payload.Name = &name

        updatedTask, err := UpdateTaskAction(tx, task.ID, payload)

        if err != nil {
            t.Fatalf("error while updating task, %s\n", err)
        }

        if updatedTask.Name != *payload.Name {
            t.Errorf("expected task name to be %s, but got %s\n", *payload.Name, updatedTask.Name)
        }
    })

    t.Run("Testing updating task status", func(t *testing.T) {
        task := mockTask(t, tx)
        payload := UpdateTaskProp{}
        completed := true
        payload.Completed = &completed

        updatedTask, err := UpdateTaskAction(tx, task.ID, payload)

        if err != nil {
            t.Fatalf("error while updating task, %s\n", err)
        }

        if !updatedTask.Completed {
            t.Errorf("expected task status to be %t, but got %t\n", *payload.Completed, updatedTask.Completed)
        }
    })
}

func TestDeleteTaskBulkAction(t *testing.T) {
    tx := getDBTransaction(t)
    defer tx.Rollback()

    task := mockTask(t, tx)
    task2 := mockTask(t, tx)

    delCount, err := DeleteTaskBulkAction(tx, []int{task.ID, task2.ID})

    if err != nil {
        t.Fatalf("error while deleting task from the database, %s\n", err)
    }

    if delCount != 2 {
        t.Errorf("expecting delCount to be 2 got %d\n", delCount)
    }
}

func TestListTaskAction(t *testing.T) {
    tx := getDBTransaction(t)
    defer tx.Rollback()

    _, err := AddTaskAction(tx, AddTaskProp{
        "Test",
        true,
    })

    mockTask(t, tx)
    mockTask(t, tx)

    if err != nil {
        t.Fatalf("error while mocking tasks for test")
    }

    t.Run("Should apply the provided filters to the query", func(t *testing.T) {
        completed := false

        sort := [2]string{"id", "desc"}
        tasks, err := ListTasksAction(tx, ListTaskProps{
            WhereCompleted: &completed,
            SortBy: &sort,
        })

        if err != nil {
            t.Fatalf("error while listing tasks from the database, %s\n", err)
        }

        if len(tasks) != 2 {
            t.Errorf("expected 2 itens in the list but got %d\n", len(tasks))
        }

        if tasks[0].ID != 3 {
            t.Errorf("the list should be reversed but the id of the first id is %d, instead of 3.\n", tasks[0].ID)
        }
    })

    t.Run("Should return a list of all tasks", func(t *testing.T) {
        tasks, err := ListTasksAction(tx, ListTaskProps{})

        if err != nil {
            t.Fatalf("error while listing tasks, %s\n", err)
        }

        if len(tasks) != 3 {
            t.Errorf("expected a list with all the tasks in the database, but got %d\n", len(tasks))
        }
    })
}

func TestListTaskActionByID (t *testing.T) {
    db := getDBTransaction(t)
    defer db.Rollback()

    task := mockTask(t, db)

    t.Run("Should return nil if didn't find the task with the provided ID", func (t *testing.T) {
        _, err := ListTaskActionByID(db, 69)

        if err == nil {
            t.Errorf("expected to receive an ErrNoRow got: %v\n", err)
        }
    })

    t.Run("Should return the task with the provede ID", func (t *testing.T) {
        mockedTask, err := ListTaskActionByID(db, uint(task.ID))

        if err != nil {
            t.Fatalf("received an error while listing task %d\n", err)
        }

        if mockedTask.ID != task.ID {
            t.Errorf("expected returned task to have the same ID of the mocked one, got %d\n", mockedTask.ID)
        }
    })
}

func getDBTransaction(t testing.TB) (*sql.Tx) {
    t.Helper()
    db, err := OpenDatabase("../")

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

func mockTask(t testing.TB, db DB) Task {
    t.Helper()
    task, err := AddTaskAction(db, AddTaskProp{
        "Test",
        false,
    })

    if err != nil {
        t.Fatalf("error while mocking task, %s\n", err)
    }

    return task
}
