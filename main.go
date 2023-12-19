package main

import (
	"errors"
	"fmt"
	"go_todo/database"
	taskAction "go_todo/database"
	"os"
	"strconv"
	"strings"
)

func main() {
    args := os.Args

    if len(args) == 1 {
        fmt.Printf("Welcome to To-do, Go!\n\n")

        printHelp()
        return
    }

    db, err := taskAction.OpenDatabase("./")

    if err != nil {
        fmt.Printf("error while opening the database: %s\n", err)
    }

    switch args[1]  {
        case "a":
            addTask(db, args[1:])
        case "l":
            listTasks(db, args[1:])
        case "d":
            deleteTasks(db, args[1:])
        case "u":
            updateTask(db, args[1:])
        case "help":
            help(args[1:])
        default:
            fmt.Printf("%s\n\n", fmt.Sprintf("Option %s doesn't exist", args[1]))
            printHelp()
    }
}

const UsageStrAddTask = "Usage: gotodo a -name <name> [-completed <true|false>]"
func addTask(db taskAction.DB, args []string) {
    if len(args) == 1 {
        fmt.Println(UsageStrAddTask)
        return
    }

    props := taskAction.AddTaskProp{}

    optionValueMap, err := GetOptionValue(args, []string{"-name", "-completed"})

    if err != nil {
        fmt.Println(err)
        return
    }

    if taskName, ok := optionValueMap["-name"]; !ok {
        fmt.Println(UsageStrAddTask)
        return
    } else {
        props.Name = taskName
    }

    if taskStatus, ok := optionValueMap["-completed"]; ok {
        if taskStatus == "true" {
            props.Completed = true
        } else if taskStatus == "false" {
            props.Completed = false
        } else {
            fmt.Println(UsageStrAddTask)
            return
        }
    }
    
    task, err := taskAction.AddTaskAction(db, props)

    if err != nil {
        fmt.Printf("error while creating task: %s\n", err)
        return
    }

    fmt.Printf("Task with ID: %d created!\n", task.ID)
}

const DefaultListUsageStr = "Usage: gotodo l [-sort <id|name>,<asc,desc>] [-completed <true|false>]"
func listTasks(db database.DB, args []string) {
    props := database.ListTaskProps{}
    optionValueMap, err := GetOptionValue(args, []string{"-sort", "-completed"})
    if err != nil {
        fmt.Println(err)
        return
    }
    if sortVal, ok := optionValueMap["-sort"]; ok {
        colOrd := strings.Split(sortVal, ",") 

        if !Include([]string{"id", "name"}, colOrd[0]){
            fmt.Println(DefaultListUsageStr)
            return
        }

        if !Include([]string{"asc", "desc"}, colOrd[1]) {
            fmt.Println(DefaultListUsageStr)
            return
        }
        sortingParameters := [2]string{colOrd[0], colOrd[1]}
        props.SortBy = &sortingParameters
    }

    if filterVal, ok := optionValueMap["-completed"]; ok {
        if !Include([]string{"true", "false"}, filterVal){
            fmt.Println(DefaultListUsageStr)
            return
        }

        if filterVal == "true" {
            val := true
            props.WhereCompleted = &val
        } else {
            val := false
            props.WhereCompleted = &val
        }
    }
    printTasksList(db, props)
}

const DefaultDeleteUsageStr = "Usage: gotodo d <...ids>"
func deleteTasks(db database.DB, args []string) {
    if len(args) == 1 {
        fmt.Println(DefaultDeleteUsageStr)
        return
    }

    ids := make([]int, len(args))

    for idx, arg := range args[1:] {
        id, err := strconv.Atoi(arg)
        if err != nil {
            fmt.Println(fmt.Sprintf("Error: '%s' isn't a numeric character", arg))
            return
        }
        ids[idx] = id
    }

    deleteCount, err := database.DeleteTaskBulkAction(db, ids)

    if err != nil {
        fmt.Println("Error:", err)
        return
    }

    fmt.Println(fmt.Sprintf("Deleted %d tasks.", deleteCount))
}

const DefaultUpdateUsageStr = "Usage: gotodo u <id> <-name string>|<-completed true|false>"
func updateTask(db database.DB, args []string) {
    if len(args) < 2 {
        fmt.Println(DefaultUpdateUsageStr)
        return
    }

    idToUpdate, err := strconv.Atoi(args[1])

    if err != nil {
        fmt.Println(DefaultUpdateUsageStr)
        return
    }

    optionValueMap, err := GetOptionValue(args, []string{"-name", "-completed"})

    if err != nil {
        fmt.Println(err)
        return
    }

    nameVal, isNamePresent := optionValueMap["-name"]
    completedVal, isCompletedPresent := optionValueMap["-completed"]

    if !isNamePresent && !isCompletedPresent {
        fmt.Println(DefaultUpdateUsageStr)
        return
    }

    if isCompletedPresent && !Include([]string{"true", "false"}, completedVal){
        fmt.Println(DefaultUpdateUsageStr)
        return
    }

    props := database.UpdateTaskProp{}

    if isNamePresent {
       props.Name = &nameVal 
    }

    if isCompletedPresent {
        if completedVal == "true" {
            val := true
            props.Completed = &val
        } else {
            val := false
            props.Completed = &val
        }
    }

    updatedTask, err := database.UpdateTaskAction(db, idToUpdate, props)

    if err != nil {
        if fmt.Sprintf("%s", err) == "Task doesn't exist" {
            fmt.Println(fmt.Sprintf("Error: %s", err))
            return
        }
        fmt.Println(err)
    }

    fmt.Println(fmt.Sprintf("Task %d updated", updatedTask.ID))
}

const DefaultHelpUsageStr = "Usage: gotodo help <a|l|d|u>"
func help(args []string) {
    if len(args) == 1 {
        fmt.Println(DefaultHelpUsageStr)
        return
    }

    switch args[1] {
        case "a":
            fmt.Println(UsageStrAddTask)
        case "l":
            fmt.Println(DefaultListUsageStr)
        case "d":
            fmt.Println(DefaultDeleteUsageStr)
        case "u":
            fmt.Println(DefaultUpdateUsageStr)
        default: 
            fmt.Println(fmt.Sprintf("Option %s not recognized", args[1]))
    }
}

func printTasksList(db database.DB, props database.ListTaskProps) {
    tasks, err := database.ListTasksAction(db, props)
    if err != nil {
        fmt.Println(err)
        return
    }
    for _, task := range tasks {
        if task.Completed {
            fmt.Printf("%d.[x] - %s\n", task.ID, task.Name)
            continue
        }
        fmt.Printf("%d.[ ] - %s\n", task.ID, task.Name)
    }
}

func Include(set []string, val string) bool {
    isPresent := false
    for _, setVal := range set {
        if setVal == val {
            isPresent = true
            break
        }
    }

    return isPresent
}

func GetOptionValue(args []string, allowedParameters []string) (map[string]string, error) {
    maps := make(map[string]string)

    lastOption := ""

    for _, arg := range args {
        if arg[0] == '-' {
            if Include(allowedParameters, arg) {
                lastOption = arg
                continue
            }
            return nil, errors.New(fmt.Sprintf("Parameter %s not recognized", arg))
        }

        if lastOption != "" {
            maps[lastOption] = arg
            lastOption = ""
            continue
        }
    }

    return maps, nil
}

func printHelp() {
    fmt.Printf("Usage: gotodo <option>\n\n")

    fmt.Println("a - Add tasks")
    fmt.Println("l - List tasks")
    fmt.Println("d - Delete tasks")
    fmt.Println("u - Update task")

    fmt.Printf("\n")

    fmt.Println("For more information about a option: gotodo help <a|l|d|u>")
}
