package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/TwiN/go-color"
	"github.com/thedevsaddam/gojsonq"
)

type JsonData struct {
	Id       int       `json:"id"`
	Activity string    `json:"activity"`
	Short    string    `json:"short"`
	Hours    int       `json:"hours"`
	Minutes  int       `json:"minutes"`
	Projects []Project `json:"projects"`
	Function string    `json:"function"`
}

type Project struct {
	Name  string   `json:"name"`
	Tasks []string `json:"tasks"`
}

var ProgramVersion = "1.3" // Update version
var filename = "data/data.json"

//go:generate goversioninfo -icon=resource/timem.ico -manifest=resource/goversioninfo.exe.manifest

func main() {

	// Create data.json file to save data if not exist
	MakeDirAndJson()

	// Start commandline
	Commandline()
}

/*<=================================================== Main functions ===================================================>*/

// Command line
func Commandline() {

	// Print commands to console
	Commandline_commands()

	// Add default commands to data
	data := AddDefaultCommands()

	// Get reader
	reader := bufio.NewReader(os.Stdin)

	for true {

		command := Get_input(reader)
		MainSwitch(command, data, reader)

	}
}

// Add default commands to data temporarly
func AddDefaultCommands() []JsonData {

	// Get data from json
	data := OpenAndGetDataFromJson()

	// If db contains data then get a new id
	GetId := CheckDBForData(data)

	// Make map of maps
	DefaultCommands := make(map[int]map[string]string)

	// Add to map
	DefaultCommands[1] = map[string]string{}
	DefaultCommands[1]["name"] = "top"
	DefaultCommands[1]["short"] = "t"

	DefaultCommands[2] = map[string]string{}
	DefaultCommands[2]["name"] = "add"
	DefaultCommands[2]["short"] = "a"

	DefaultCommands[3] = map[string]string{}
	DefaultCommands[3]["name"] = "delete"
	DefaultCommands[3]["short"] = "del"

	DefaultCommands[4] = map[string]string{}
	DefaultCommands[4]["name"] = "q"
	DefaultCommands[4]["short"] = "00"

	// Append to data
	for _, value := range DefaultCommands {
		NewValues := ConvertAnswersToJsonData(value["name"], value["short"], GetId, value["name"])
		data = append(data, NewValues)
	}

	return data
}

// MainSwitch
func MainSwitch(command string, data []JsonData, reader *bufio.Reader) {

	switch command {
	case "add", "a":
		AddActivity()
	case "delete", "del":
		DeleteActivity()
	case "quit", "q", "00":
		quit()
	default:

		start := time.Now()

		for _, value := range data {
			if value.Activity == command || value.Short == command || fmt.Sprint(value.Id) == command {
				ClearScreen()
				StartActivity(reader, start, value.Activity, value.Id)
			}
		}

		Feedback("<< ", "No such command or activity!", " >>", true)
		fmt.Printf(ColorGreen("\n=> "))

	}
}

/*<=================================================== Activity functions ===================================================>*/

// The loop
func StartActivity(reader *bufio.Reader, start time.Time, Activity string, id int) {

	// Get data from json
	data := OpenAndGetDataFromJson()

	// Prevent ERROR when some activity is deleted and max ID has been changed!
	id = FindRealId(id)

	// Get hours and minutes from json
	hours := data[id].Hours
	minutes := data[id].Minutes

	// Tell user about started activity
	PrintActivityInfo(id, data, Activity, start, hours, minutes)

	// Print Projects
	PrintProjects(id)

	// Loop for input
	loop := true

	// Pause print
	PausePrintingCommands := false

	// Define Pause time
	PauseTime := 0

	for loop {

		// Print main commands
		if !PausePrintingCommands {
			PrintCommands("Projects")

		}

		// Get input from user
		command := Get_input(reader)

		// Elapsed time since activity start
		elapsed := time.Since(start)

		switch command {
		case "add", "a":
			// Add new project
			AddProject(id)
		case "del", "d", "delete":
			// Delete project
			DeleteProject(id)
			PrintElapsedTime(Activity, elapsed, start)
		case "projects", "p":
			// Print all projects
			ClearScreen()
			PrintProjects(id)
		case "select", "s":
			// Select project
			SelectProject(id, start, Activity)
		case "quit", "00", "q":

			// Tell user elapsed time
			Feedback("\n<< You have spent ", elapsed, " >>\n", false)

			// Ask for save time
			Save_time(reader, elapsed, id, PauseTime)

			// End loop
			loop = false

		case "pause", "+":

			// Tell user that this activity is paused
			Feedback("<< [", Activity, "] paused! Press any key to continue! >>", true)

			// Print pause commands
			PausePrintingCommands = true

			// Time now
			startPause := time.Now()

			// Wait for pressing any key or enter
			PressEnter()

			// Elapsed pause time
			elapsedPause := time.Since(startPause)

			// Add minutes to pause time
			PauseTime += int(math.Round(elapsedPause.Minutes()))

			// Tell user about Unpause
			Feedback("<< Unpaused [Pause time: ", elapsedPause, "] >>\n", false)

			// Print default commands
			PausePrintingCommands = false

		default:
			ClearScreen()
			PrintElapsedTime(Activity, elapsed, start)

		}
	}
}

// Add new activity to json file
func AddActivity() {

	// Questions array
	questions := []string{"\n<< Activity name? >>", "\n << Short name? >>"}

	// Store answers
	Answers := []string{}

	// Get reader
	reader := bufio.NewReader(os.Stdin)

	// Get data from json
	data := OpenAndGetDataFromJson()

	for _, value := range questions {

		// Bookmark
	loop:

		Feedback(value, "", "\n=> ", false)

		// Get answer
		readerAnswer := Get_input(reader)

		// Check if name already exist in db
		for _, value := range data {

			switch readerAnswer {
			case value.Activity, value.Short, "delete", "del", "quit", "q", "add", "a", "t", "top", "back", "b":

				// Tell user
				Feedback("[ERROR] : '", readerAnswer, "' already exist in db\n", true)

				// Restart the for loop, go to back to loop label
				goto loop
			}

		}

		// Add answer to answer array
		Answers = append(Answers, readerAnswer)

	}

	// If db contains data then get a new id
	GetId := CheckDBForData(data)

	// Convert to JsonData
	NewValues := ConvertAnswersToJsonData(Answers[0], Answers[1], GetId, "Default")

	// Add new Values to the end of file
	data = append(data, NewValues)

	// Convert it back to byte
	dataBytes := MarshalIndentToByte(data, "AddItem")

	// Override json file with updated data
	WriteToFile(dataBytes)

	// Clear the screen
	ClearScreen()

	// Go back to CommandLine
	Commandline()
}

// Get Top activities
func topActivities(data []JsonData) {

	// Select file
	jq := gojsonq.New().File(filename)

	// Sort by activity, hours and minutes and select top 5
	top := jq.Limit(5).SortBy("hours", "desc").Select("activity", "hours", "minutes").Get()

	// Pretty print
	b, err := json.MarshalIndent(top, "", "  ")

	// Print error if any
	ErrorHandling(err, "topActivities func")

	ClearScreen()

	// Print activities
	fmt.Print(string(b))

	// Press enter to go back to commandline
	Feedback("\n<< PRESS", " ENTER ", "TO GO BACK TO COMMANDLINE >>", false)

	// Check if enter is pressed
	PressEnter()

	ClearScreen()

	// Start commandline
	Commandline()
}

// Delete activity
func DeleteActivity() {
	// Ask for id
	id := AskForId()

	// Clear the screen
	ClearScreen()

	// Get data from json
	data := OpenAndGetDataFromJson()

	// Find Index
	index := FindIndexOf(id, data)

	// Check if index exist
	if index == -1 {

		// Tell user that index does not exist
		Feedback("<< ID: '", id, "' not found! >>", true)

		// Wait for enter
		PressEnter()

		// Return to commandline
		Commandline()
	}

	// Delete
	data = append(data[:index], data[index+1:]...)

	// Convert it back to byte
	dataBytes := MarshalIndentToByte(data, "DeleteItem")

	// Override json file with updated data
	WriteToFile(dataBytes)

	// Tell about successful operation
	Feedback("<< ID: '", id, "' Removed! >>", true)

	// Press enter to continue
	PressEnter()

	// Return to commandline
	Commandline()
}

// Save time
func Save_time(reader *bufio.Reader, elapsed time.Duration, id int, PauseTime int) {

	// Print save message
	Feedback("\n<< Do you want to save the time? (", "type no if not", ")\n=> ", false)

	// Ask before delete
	check := DeleteCheckQuestion("")

	if check {

	

		// If 'no' is entered tell the user
		Feedback("<< ", "LAST TIME NOT SAVED", " >>\n", true)

		// Press enter to continue
		PressEnter()

		ClearScreen()

		// Return to commandline
		Commandline()

	} else {

		ClearScreen()
		// Tell the user about saving the time
		Feedback("<< ", "LAST TIME HAS BEEN SAVED", " >>\n", false)

		// Save time to db
		UpdateJsonFile(elapsed, id, PauseTime)

		// Return to commandline
		Commandline()
	}
}

// Save time function
func UpdateJsonFile(elapsed time.Duration, id int, PauseTime int) {
	// Get data from json
	data := OpenAndGetDataFromJson()

	//Old + new Minutes
	NewMinutes := int(float64(data[id].Minutes) + math.Round(elapsed.Minutes()))

	// Remove PauseTime minutes
	if PauseTime > 0 {
		NewMinutes -= PauseTime
	}

	// Get hours out of all minutes
	GetHours := NewMinutes / 60

	// Remove hours and get minutes left
	GetMinutes := NewMinutes - (GetHours * 60)

	// Add new hours
	HoursToAdd := data[id].Hours + GetHours

	// Add new minutes
	MinutesToAdd := GetMinutes

	// Add new minutes to db
	data[id].Minutes = MinutesToAdd

	// Add new hours to db
	data[id].Hours = HoursToAdd

	// Convert it back to byte
	dataBytes := MarshalIndentToByte(data, "UpdateItem")

	// Override json file with updated data
	WriteToFile(dataBytes)
}

/*<=================================================== Project functions ===================================================>*/

// Add new Project
func AddProject(id int) {

	// Bookmark
loop:

	// Get reader
	reader := bufio.NewReader(os.Stdin)

	// Ask for project name
	Feedback("\n<< Project name? >>", "", "\n=> ", false)

	// Save answer
	pName := Get_input(reader)

	// Get data from json
	data := OpenAndGetDataFromJson()

	// Check if project name already exist in db
	for _, value := range data {
		for _, v := range value.Projects {
			switch pName {
			case v.Name:

				// Tell user that project already exist in database
				Feedback("\n<< [Error]: Project '", pName, "' already exist in db >>\n", true)

				// Restart the for loop, go to back to loop label
				goto loop
			}
		}

	}

	// Init new project
	NewProject := Project{pName, []string{}}

	// Append new project to db
	data[id].Projects = append(data[id].Projects, NewProject)

	// Convert it back to byte
	dataBytes := MarshalIndentToByte(data, "UpdateItem")

	// Override json file with updated data
	WriteToFile(dataBytes)

	// Tell the user about successful operation
	Feedback("\n<< Project '", pName, "' added to db! >>\n", false)
}

// Delete Project
func DeleteProject(id int) {
loop:
	// Ask and save id
	projectID := AskForId()

	// Get data from json
	data := OpenAndGetDataFromJson()

	// Save maximum id
	MaxID := len(data[id].Projects) - 1

	if projectID > MaxID {
		// ERROR message
		Feedback("<< [ERROR] Max ID: [", MaxID, "] >>\n\n", true)
		goto loop
	} else if projectID < 0 {
		// ERROR message
		Feedback("<< [ERROR] id cant be negative!", "", " >>\n", true)
		goto loop
	}

	// Project details
	project := data[id].Projects[projectID]

	// Ask before delete
	check := DeleteCheckQuestion(project.Name)

	// If check is false delete item
	if !check {
		// Delete
		data[id].Projects = append(data[id].Projects[:projectID], data[id].Projects[projectID+1:]...)

		// Convert it back to byte
		dataBytes := MarshalIndentToByte(data, "DeleteItem")

		// Override json file with updated data
		WriteToFile(dataBytes)

		// Tell user about successful operation
		Feedback("\nProject '", project.Name, "' has been deleted!\n", true)
	} else {
		ClearScreen()
	}
}

/*<=================================================== Tasks functions ===================================================>*/
// Print tasks
func ShowTasks(id int, projectid int) {

	// Get data from json
	data := OpenAndGetDataFromJson()

	// Save Project name and tasks
	project := data[id].Projects[projectid]

	// Print all tasks with id's
	for key, value := range project.Tasks {
		Feedback("\nTask(", key, ") : '", false)
		Feedback("", value, "'", false)
	}
	fmt.Println()

}

// Delete Task
func DeleteTask(id int, projectid int) {
loop:
	// Ask and save id
	taskID := AskForId()

	// Get data from json
	data := OpenAndGetDataFromJson()

	// Max task id
	MaxTaskID := len(data[id].Projects[projectid].Tasks) - 1

	if taskID > MaxTaskID {
		Feedback("<< [ERROR] Max ID: [", MaxTaskID, "] >>\n\n", true)
		goto loop
	} else if taskID < 0 {

		// ERROR message
		Feedback("<< [ERROR] id cant be negative!", "", " >>\n\n", true)
		goto loop
	}

	// Project details
	project := data[id].Projects[projectid]

	// Ask before delete
	check := DeleteCheckQuestion(project.Tasks[taskID])

	if !check {
		// Delete
		data[id].Projects[projectid].Tasks = append(data[id].Projects[projectid].Tasks[:taskID], data[id].Projects[projectid].Tasks[taskID+1:]...)

		// Convert it back to byte
		dataBytes := MarshalIndentToByte(data, "DeleteItem")

		// Override json file with updated data
		WriteToFile(dataBytes)

		// Tell user about successful operation
		Feedback("\nTask '", project.Tasks[taskID], "' has been deleted!\n", true)

		// Print commands
		PrintCommands("Tasks")

	} else {
		// Print commands
		PrintCommands("Tasks")
	}
}

// Add task to project
func SelectProject(id int, start time.Time, Activity string) {

	// Bookmark
loop:

	// Get reader
	reader := bufio.NewReader(os.Stdin)

	// Ask for id
	ProjectId := AskForId()

	// Get data from json
	data := OpenAndGetDataFromJson()

	// Save maximum id
	MaxID := len(data[id].Projects) - 1

	// Error if id is bigger than MaxID or negative
	if ProjectId > MaxID {

		// ERROR message
		Feedback("<< [ERROR] Max ID: [", MaxID, "] >>\n\n", true)

		// Go to bookmark
		goto loop

	} else if ProjectId < 0 {

		// ERROR message
		Feedback("<< [ERROR] id cant be negative!", "", " >>\n\n", true)

		// Go to bookmark
		goto loop
	}

	ClearScreen()

	// Find project name by projectid
	CurrentProject := data[id].Projects[ProjectId]

	// Save project name
	pName := CurrentProject.Name

	// Print project name
	Feedback("\n<< Project: ", pName, " >>\n", false)

	// Show tasks
	ShowTasks(id, ProjectId)

	// Print add task commands
	PrintCommands("Tasks")

	loop2 := true

	for loop2 {

		// Get input from user
		command := Get_input(reader)

		// Elapsed time since activity start
		elapsed := time.Since(start)

		switch command {
		case "back", "b":

			loop2 = false
			ClearScreen()
			// Print elapsed time since start
			PrintElapsedTime(Activity, elapsed, start)
			// Print Projects
			PrintProjects(id)

		case "add", "a":
			AddTask(pName, ProjectId, id)
		case "delete", "del", "d":
			DeleteTask(id, ProjectId)
		case "show", "s":
			ShowTasks(id, ProjectId)
			PrintCommands("Tasks")
		case "quit", "q", "00":

			// Tell user about elapsed time
			Feedback("<< You have spent ", elapsed, " >>\n", false)

			// Ask for save time
			Save_time(reader, elapsed, id, 0)

			// End loop
			loop2 = false

		default:
			ClearScreen()

			// Print elapsed time since start
			PrintElapsedTime(Activity, elapsed, start)

			// Print commands for tasks
			PrintCommands("Tasks")

		}
	}
}

// Add task
func AddTask(pName string, SelectedIdint int, id int) {

	reader := bufio.NewReader(os.Stdin)

	// Ask task name
	Feedback("\n<< Task name? >>", "", "\n=> ", false)

	// Save answer
	tName := Get_input(reader)

	// Get data from json
	data := OpenAndGetDataFromJson()

	// Empty slice for tasks
	var tasks []string

	// Take old tasks and add to tasks slice
	for key, value := range data[id].Projects {
		if key == SelectedIdint {
			tasks = value.Tasks
		}
	}

	// Append new task to tasks slice
	tasks = append(tasks, tName)

	// Append new tasks slice to data.json
	data[id].Projects[SelectedIdint] = Project{Name: pName, Tasks: tasks}

	// Convert it back to byte
	dataBytes := MarshalIndentToByte(data, "UpdateItem")

	// Override json file with updated data
	WriteToFile(dataBytes)

	// Print about successful operation
	PrintTaskAddedToProject(tName, pName)

	// Print Commands
	PrintCommands("Tasks")
}

/*<=================================================== Print functions ===================================================>*/

// Print commands
func Commandline_commands() {

	// Print how much time is left till 22:00
	PrintTimeleft()

	// Get data from json
	data := OpenAndGetDataFromJson()

	if len(data) == 0 {
		Feedback("", "<< WARNING: No data in database >>", "", true)
	} else {
		PrintAllActivities(data)
	}

	// Add main commands
	PrintCommands("Activity")
}

// Calculate time left
func PrintTimeleft() {

	// Time now
	start := time.Now()

	// Calculate minutes
	MinutesNow := (start.Hour() * 60) + start.Minute()

	// Target time 22:00
	EndTime := 1320

	// Calculate how many minutes till target time
	MinutesTillEndTime := EndTime - MinutesNow

	// If Minutes till end are negative
	if MinutesTillEndTime < 0 {
		MinutesTillEndTime *= -1
		MinutesTillEndTime += EndTime
	}

	// Get Hours left
	HoursLeft := MinutesTillEndTime / 60

	// Get Minutes left
	MinutesLeft := MinutesTillEndTime - (HoursLeft * 60)

	// Print main info about programm
	PrintProgramName(HoursLeft, MinutesLeft)
}

func Feedback(first interface{}, middle interface{}, last interface{}, red bool) {

	ToPrint := []string{ColorGreen(first), ColorWhite(middle), ColorGreen(last)}

	if red {
		ToPrint = []string{ColorRed(first), ColorWhite(middle), ColorRed(last)}
	}

	for _, v := range ToPrint {
		fmt.Printf(v)
	}
}

func ColorRed(item interface{}) string {
	colorized := fmt.Sprintf(color.Colorize(color.Red, "%v"), item)
	return colorized
}

func ColorPurple(item interface{}) string {
	colorized := fmt.Sprintf(color.Colorize(color.Purple, "%v"), item)
	return colorized
}

func ColorGreen(item interface{}) string {
	colorized := fmt.Sprintf(color.Colorize(color.Green, "%v"), item)
	return colorized
}

func ColorWhite(item interface{}) string {
	colorized := fmt.Sprintf(color.Colorize(color.White, "%v"), item)
	return colorized
}

func PrintElapsedTime(Activity string, elapsed time.Duration, start time.Time) {
	Feedback("\n<< [", Activity, "]", false)
	Feedback(" Elapsed Time: ", elapsed, "", false)
	Feedback(" since start ", start.Format("15:04:05"), " >>\n", false)
}

// Print Program name and version
func PrintProgramName(HoursLeft int, MinutesLeft int) {

	Feedback("\n<< VK TimeManager v", ProgramVersion, " >>\n", false)
	Feedback("\n<< You have ", HoursLeft, " hours ", false)
	Feedback("and ", MinutesLeft, " minutes left", false)
	Feedback(" till ", "22:00", " >>\n\n", false)
}

func PrintTaskAddedToProject(tName string, pName string) {
	Feedback("\n<< Task '", tName, "' added to project '", false)
	Feedback("", pName, "'! >>\n", false)
}

func PrintCommands(command_type string) {

	// Print command type
	fmt.Printf(ColorGreen("\n<< ") + ColorWhite(command_type) + ColorGreen(" Commands >> \n"))

	// Print commands by type
	switch command_type {
	case "Activity":
		PrintActivityCommands()
	case "Projects":
		PrintProjectsCommands()
	case "Tasks":
		PrintTasksCommands()
	}

	fmt.Printf(ColorGreen("\n=> "))
}

func PrintActivityCommands() {

	Feedback("<< | <", "top", "> or ", false)
	Feedback("<", "t", ">", false)

	Feedback(" | <", "add", "> or ", false)
	Feedback("<", "a", ">", false)

	Feedback(" | <", "delete", "> or ", false)
	Feedback("<", "del", ">", false)
	Feedback(" | <", "quit", "> or ", false)
	Feedback("<", "q", "> or ", false)
	Feedback("<", "00", ">  | >>", false)
}

func PrintProjectsCommands() {
	Feedback("<< | <", "add", "> or ", false)
	Feedback("<", "a", ">", false)
	Feedback(" | <", "delete", "> or ", false)
	Feedback("<", "del", ">", false)
	Feedback(" | <", "projects", "> or ", false)
	Feedback("<", "p", ">", false)
	Feedback(" | <", "select", "> or ", false)
	Feedback("<", "s", "> | >>", false)
	Feedback("\n<< | <", "pause", "> or ", false)
	Feedback("<", "+", ">", false)
	Feedback(" | <", "quit", "> or ", false)
	Feedback("<", "q", "> or ", false)
	Feedback("<", "00", "> | >>", false)
}

func PrintTasksCommands() {
	Feedback("<< | <", "add", "> or ", false)
	Feedback("<", "a", ">", false)
	Feedback(" | <", "delete", "> or ", false)
	Feedback("<", "del", ">", false)
	Feedback(" | <", "show", "> or ", false)
	Feedback("<", "s", "> | >>", false)
	Feedback("\n<< | <", "back", "> or ", false)
	Feedback("<", "b", ">", false)
	Feedback(" | <", "quit", "> or ", false)
	Feedback("<", "q", "> or ", false)
	Feedback("<", "00", "> | >>", false)
}

func MakeCommandsArray(name string, short string) string {
	array := ColorGreen("<") + ColorPurple(name) + ColorGreen("> or <") + ColorPurple(short) + ColorGreen(">")
	return array
}

func PrintAllActivities(data []JsonData) {

	Feedback("<< ", " What do you want to do now? ", ">>\n", false)

	// Print all activities
	for _, component := range data {
		Feedback("<< [", component.Hours, "h:", false)
		Feedback("", component.Minutes, "m] ", false)
		Feedback("", component.Activity, " || ", false)
		Feedback("", component.Short, "(", false)
		Feedback("", component.Id, ") >>\n", false)
	}
}

// Tell user about started activity
func PrintActivityInfo(id int, data []JsonData, Activity string, start time.Time, hours int, minutes int) {

	Feedback("<< Starting ", Activity, "", false)
	Feedback(" at ", start.Format("02.01.2006 15:04:05"), " >>\n", false)
	Feedback("\n<< Total time spent on this activity: ", hours, " hours ", false)
	Feedback("", minutes, " minutes >>\n", false)
	//Feedback("\n<< Nr of Projects: ", len(data[id].Projects), " >>\n", false)
}

// Print projects
func PrintProjects(id int) {

	// Get data from json
	data := OpenAndGetDataFromJson()

	Feedback("\n<< My Projects (", len(data[id].Projects), ") >>\n", false)

	// Print all projects id --> name --> tasks
	for key, value := range data[id].Projects {

		Feedback("<< (", key, ")'", false)
		Feedback("", value.Name, "' | (", false)
		Feedback("", len(value.Tasks), " Tasks) >>\n", false)

	}
}

/*<=================================================== Small Help functions ===================================================>*/

// Ask before delete
func DeleteCheckQuestion(name string) bool {

	if name != "" {
		Feedback("<< Do you really want to delete '", name, "' ??? >>", true)
	}

	// Get reader
	reader := bufio.NewReader(os.Stdin)

	// Get input
	input := Get_input(reader)

	// Check if 'no' is entered
	check := "no" == input

	return check
}

func PressEnter() {
	var command string
	fmt.Scanln(&command)
}

// Check db for data and return bool
func CheckDBForData(data []JsonData) bool {
	// Dont get id if db is empty
	GetId := false

	// Get id if db is not empty
	if len(data) != 0 {
		GetId = true
	}

	return GetId
}

// Open and get data
func OpenAndGetDataFromJson() []JsonData {
	// Open json file
	file := ReadFile()

	// Get data and store it in data variable
	data := ConvertToJsonDataStruct(file)

	return data
}

// Make dir and data.json if not exist
func MakeDirAndJson() {
	if _, err := os.Stat(filename); os.IsNotExist(err) {

		// Make new directory data
		_ = os.Mkdir("data", 0700)

		f, err := os.Create(filename)
		ErrorHandling(err, "MakeDirAndJson")

		defer f.Close()
		fmt.Fprintln(f, "[]")
	}
}

// Ask for id
func AskForId() int {

	// Bookmark
loop:

	// Ask for id
	Feedback("<< ", "", "ID: ", true)

	// Get reader
	reader := bufio.NewReader(os.Stdin)

	// Get input as string
	GetIdString := Get_input(reader)

	// Convert string to int
	GetId, err := strconv.Atoi(GetIdString)

	// ERROR if a string or a negative number is entered
	if err != nil {

		// Error message
		Feedback("[ERROR] : ", "ID", " must be a number!\n", true)

		// Go back and ask again
		goto loop
	}

	return GetId
}

// Encrypt new data and construct a WebsiteData struct for adding it to json file
func ConvertAnswersToJsonData(Activity_Name string, Activity_Name_short string, GetLastid bool, function string) JsonData {

	// ID is 0 if database is empty
	Id := 0

	// If db is not empty get last id + 1
	if GetLastid {
		Id = GetLastId(filename)
	}

	ValuesToAdd := JsonData{
		Id:       Id,
		Activity: Activity_Name,
		Short:    Activity_Name_short,
		Hours:    0,
		Minutes:  0,
		Projects: []Project{},
		Function: function,
	}

	return ValuesToAdd
}

// Get max id and add +1
func GetLastId(filename string) int {
	BiggestId := gojsonq.New().File(filename).Max("id")
	id := 0
	if BiggestId == 0 {
		id = 1
	} else {
		id = int(BiggestId + 1)
	}

	return id
}

// Find Real id to prevent error when some activity is deleted
func FindRealId(id int) int {
	FindAllitems := gojsonq.New().File(filename).Count()

	Maxid := int(FindAllitems) - 1

	// if entered id is bigger then maximum possible id change it
	if id > Maxid {
		id = Maxid
	}

	return id
}

// Find index
func FindIndexOf(element int, data []JsonData) int {
	for index, Struct := range data {
		if element == Struct.Id {
			return index
		}
	}
	return -1 //not found.
}

// MarshalIndent []WebsiteData array to []byte (Makes json pretty)
func MarshalIndentToByte(data []JsonData, location string) []byte {
	dataBytes, err := json.MarshalIndent(data, "", "  ")
	ErrorHandling(err, location)

	return dataBytes
}

// Open file
func ReadFile() []byte {
	file, err := ioutil.ReadFile("data/data.json")
	ErrorHandling(err, "ReadFile")
	return file
}

// Write to file (override)
func WriteToFile(dataBytes []byte) {
	var err error
	err = ioutil.WriteFile("data/data.json", dataBytes, 0644)
	ErrorHandling(err, "WriteToFile")
}

// Convert []byte to WebsiteData struct (array)
func ConvertToJsonDataStruct(body []byte) []JsonData {

	JsonDataStruct := []JsonData{}

	err := json.Unmarshal(body, &JsonDataStruct)
	ErrorHandling(err, "ConvertToJsonDataStruct")

	return JsonDataStruct
}

// Check answer
func Get_input(reader *bufio.Reader) string {
	// Read the answer
	input, _ := reader.ReadString('\n')
	// convert CRLF to LF
	input = strings.Replace(input, "\r\n", "", -1)

	return input
}

// Clear screen
func ClearScreen() {
	clearScreen := exec.Command("cmd", "/c", "cls")
	clearScreen.Stdout = os.Stdout
	err := clearScreen.Run()
	ErrorHandling(err, "ClearScreen")
}

// Handle Errors
func ErrorHandling(err error, location string) {
	if err != nil {
		Feedback(location, ":", err.Error(), true)
	}
}

// quit
func quit() {
	ClearScreen()
	os.Exit(0)
}
