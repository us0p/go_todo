package database

import (
	"errors"
	"fmt"
)

type Task struct {
    ID int
    Name string
    Completed bool
}

type AddTaskProp struct {
    Name string
    Completed bool
}

const ADD_TASK_SQL = "INSERT INTO tasks (name,completed) VALUES ($1,$2) RETURNING *;"

func AddTaskAction(db DB, props AddTaskProp) (Task, error) {
    row := db.QueryRow(ADD_TASK_SQL, props.Name, props.Completed)
    task := Task{}

    err := row.Scan(
        &task.ID,
        &task.Name,
        &task.Completed,
    )

    if err != nil {
        return Task{}, err
    }

    return task, nil
}
type UpdateTaskProp struct {
    Name *string
    Completed *bool
}

const UPDATE_TASK_SQL = "UPDATE tasks SET %s WHERE id = $1 RETURNING *;"

const GET_TASK_SQL = "SELECT * FROM tasks WHERE id = $1;"
func UpdateTaskAction(db DB, taskID int, payload UpdateTaskProp) (Task, error) {
    existingRow := db.QueryRow(GET_TASK_SQL, taskID)

    task := Task{}

    existingRowErr := existingRow.Scan(
        &task.ID,
        &task.Name,
        &task.Completed,
    )

    if existingRowErr != nil {
        return Task{}, errors.New("Task doesn't exist")
    }

    query := ""

    if payload.Name != nil && payload.Completed != nil {
        query = fmt.Sprintf("name = '%s', completed = %t", *payload.Name, *payload.Completed)
    } else if payload.Name != nil && payload.Completed == nil {
        query = fmt.Sprintf("name = '%s'", *payload.Name)
    } else {
        query = fmt.Sprintf("completed = %t", *payload.Completed)
    }

    updatedQuery := fmt.Sprintf(UPDATE_TASK_SQL, query)

    row := db.QueryRow(updatedQuery, taskID)

    scanErr := row.Scan(
        &task.ID,
        &task.Name,
        &task.Completed,
    )

    if scanErr != nil {
        return Task{}, scanErr
    }

    return task, nil
}
const DELETE_TASK_SQL = "DELETE FROM tasks WHERE ID IN (%s);"
func DeleteTaskBulkAction(db DB, IDs []int) (int, error) {
    ids := ""

    for idx, id := range IDs {
        if idx != len(IDs) - 1 {
            ids += fmt.Sprintf("%d,", id)
            continue
        }
        ids += fmt.Sprintf("%d", id)
    }

    result, err := db.Exec(fmt.Sprintf(DELETE_TASK_SQL, ids))

    if err != nil {
        return 0, err
    }

    delCount, err := result.RowsAffected()

    if err != nil {
        return 0, err
    }

    return int(delCount), nil
}

type ListTaskProps struct {
    WhereCompleted *bool
    SortBy *[2]string
}

const LIST_TASKS_SQL = "SELECT * FROM tasks"

func ListTasksAction(db DB, props ListTaskProps) ([]Task, error) {
    var filters string
    if props.WhereCompleted != nil {
        filters = fmt.Sprintf("WHERE completed = %t", *props.WhereCompleted)
    }

    if props.SortBy != nil {
        filters = fmt.Sprintf("%s ORDER BY %s %s", filters, props.SortBy[0], props.SortBy[1])
    }

    query := fmt.Sprintf("%s %s;", LIST_TASKS_SQL, filters)

    rows, err := db.Query(query)

    if err != nil {
        return []Task{}, err
    }

    tasks := make([]Task, 0)

    for rows.Next() {
        task := Task{}
        if scanErr := rows.Scan(&task.ID, &task.Name, &task.Completed); scanErr != nil {
            return []Task{}, scanErr
        }
        tasks = append(tasks, task)
    }

    return tasks, nil
}

const LIST_TASK_ID_SQL = "SELECT * FROM tasks WHERE id = $1;"

func ListTaskActionByID(db DB, ID uint) (Task, error) {
    row := db.QueryRow(LIST_TASK_ID_SQL, ID)

    task := Task{}

    err := row.Scan(
        &task.ID,
        &task.Name,
        &task.Completed,
    )

    if err != nil {
        return Task{}, err
    }

    return task, nil
}
