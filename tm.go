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

var ProgramVersion = "1.26" // Update version [16 updates]
var filename = "data/data.json"

//go:generate goversioninfo -icon=resource/timem.ico -manifest=resource/goversioninfo.exe.manifest

func main() {

	// Create data.json file to save data if not exist
	MakeDirAndJson()

	// Start commandline
	Commandline()
}

/*<=================================================== Main functions ===================================================>*/

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
	fmt.Printf(color.Colorize(color.Green, "\n<================== VK TimeManager v%v ==================>\n"), ProgramVersion)
	fmt.Printf(color.Colorize(color.Green, "\n<--- You have %v hours and %v minutes left till 22:00 --->\n\n"), HoursLeft, MinutesLeft)
	
}

// Command line
func Commandline() {

	// Print commands to console
	Print_commands()

	// Add default commands to data
	data := AddDefaultCommands()

	// Get reader
	reader := bufio.NewReader(os.Stdin)

	for true {

		var command string
		fmt.Scanln(&command)

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

	for _, value := range data {
		switch command {
		case value.Activity, value.Short, fmt.Sprint(value.Id):

			if value.Function == "add" {
				AddActivity()
			} else if value.Function == "delete" {
				DeleteActivity()
			} else if value.Function == "q" {
				quit()
			} else if value.Function == "top" {
				topActivities(data)
			} else {
				ClearScreen()
				start := time.Now()
				StartActivity(reader, start, value.Activity, value.Id)
			}
		}
	}
}

// The loop
func StartActivity(reader *bufio.Reader, start time.Time, Activity string, id int) {

	// Get data from json
	data := OpenAndGetDataFromJson()

	// Get hours and minutes from json
	hours := data[id].Hours
	minutes := data[id].Minutes

	// Tell user about started activity
	fmt.Println()
	fmt.Printf(color.Colorize(color.Green, "<--- Starting %v at %v --->\n"), Activity, start.Format("02.01.2006 15:04:05"))
	fmt.Printf(color.Colorize(color.Green, "\n<--- Total time spent on this activity: %v hours %v minutes --->\n"), hours, minutes)
	fmt.Println(color.Colorize(color.Green, "\n=>> Nr of Projects:"), len(data[id].Projects))

	// Loop for input
	loop := true

	// Pause print
	pausePrintCommands := false

	// Define Pause time
	PauseTime := 0

	for loop {

		// Print main commands
		if !pausePrintCommands {
			fmt.Println(color.Colorize(color.Green, "\n--> (Press enter to see elapsed time!)"))
			fmt.Println(color.Colorize(color.Blue, "--> (Type 'add' or 'a' to add a project)"))
			fmt.Println(color.Colorize(color.Blue, "--> (Type 'delete', 'del' or 'd' to delete a project)"))
			fmt.Println(color.Colorize(color.Blue, "--> (Type 'projects' or 'p' to see projects)"))
			fmt.Println(color.Colorize(color.Blue, "--> (Type 'select' or 's' to select a project)"))
			fmt.Println(color.Colorize(color.Blue, "--> (Type 'done' or '0' or 'q' to end)"))
			fmt.Println(color.Colorize(color.Blue, "--> (Type 'pause' or '+' to pause)"))
		}

		// Commandline
		fmt.Print(color.Colorize(color.Green, "\n=>  "))

		// Get input from user
		command := Get_input(reader)

		// Elapsed time since activity start
		elapsed := time.Since(start)

		switch command {
		case "add", "a":
			// Add new project
			AddNewProject(id)

		case "del", "d", "delete":
			// Delete project
			DeleteProject(id)

		case "projects", "p":
			// Print all projects
			PrintProjects(id)

		case "select", "s":
			// Dont print commands
			pausePrintCommands = true

			// Select project
			SelectIdAndAddTask(id, start, Activity)

			// Resume printing commands
			pausePrintCommands = false

		case "done", "0", "q":

			// Tell user about elapsed time
			fmt.Printf(color.Colorize(color.Green, "You have spent %v"), elapsed)
			fmt.Println()

			// Ask for save time
			Save_time(reader, elapsed, id, PauseTime)

			// End loop
			loop = false

		case "pause", "+":

			// Tell user that this activity is paused
			fmt.Printf(color.Colorize(color.Red, "\n--> %v Paused! Press any key to continue! <--"), Activity)

			// Print pause commands
			pausePrintCommands = true

			// Time now
			startPause := time.Now()

			// Wait for pressing any key or enter
			var command string
			fmt.Scanln(&command)

			// Elapsed pause time
			elapsedPause := time.Since(startPause)

			// Add minutes to pause time
			PauseTime += int(math.Round(elapsedPause.Minutes()))

			// Tell user about Unpause
			fmt.Printf(color.Colorize(color.Green, "--> Unpaused [Pausetime: %v] <--\n"), elapsedPause)

			// Print default commands
			pausePrintCommands = false

		default:
			ClearScreen()
			fmt.Printf(color.Colorize(color.Green, "---> (%v) Elapsed Time: %v since start [%v] <---\n"), Activity, elapsed, start.Format("15:04:05"))
		}

	}
}

// Print tasks
func ShowTasks(id int, projectid int) {
	// Get data from json
	data := OpenAndGetDataFromJson()

	// Save Project name and tasks
	project := data[id].Projects[projectid]

	// Print all tasks with id's
	for key, value := range project.Tasks {
		fmt.Printf(color.Colorize(color.Purple, "\nid: (%v) task: '%v'"), key, value)
	}

	// Print commands
	fmt.Println()
	PrintAddTaskCommands()
}

// Delete Task
func DeleteTask(id int, projectid int) {
	// Ask and save id
	index := AskForId()

	// Get data from json
	data := OpenAndGetDataFromJson()

	// Project details
	project := data[id].Projects[projectid]

	// Ask before delete
	check := DeleteCheckQuestion(project.Tasks[index])

	if !check {
		// Delete
		data[id].Projects[projectid].Tasks = append(data[id].Projects[projectid].Tasks[:index], data[id].Projects[projectid].Tasks[index+1:]...)

		// Convert it back to byte
		dataBytes := MarshalIndentToByte(data, "DeleteItem")

		// Override json file with updated data
		WriteToFile(dataBytes)

		fmt.Printf(color.Colorize(color.Red, "\nTask %v has been deleted!\n"), project.Name)

		// Print commands
		fmt.Println()
		PrintAddTaskCommands()

	} else {
		// Print commands
		fmt.Println()
		PrintAddTaskCommands()
	}
}

// Delete Project
func DeleteProject(id int) {

	// Ask and save id
	index := AskForId()

	// Get data from json
	data := OpenAndGetDataFromJson()

	// Project details
	project := data[id].Projects[index]

	// Ask before delete
	check := DeleteCheckQuestion(project.Name)

	// If check is false delete item
	if !check {
		// Delete
		data[id].Projects = append(data[id].Projects[:index], data[id].Projects[index+1:]...)

		// Convert it back to byte
		dataBytes := MarshalIndentToByte(data, "DeleteItem")

		// Override json file with updated data
		WriteToFile(dataBytes)

		fmt.Printf(color.Colorize(color.Red, "\nProject %v has been deleted!\n"), project.Name)

	}
}

// Add task to project
func SelectIdAndAddTask(id int, start time.Time, Activity string) {

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
		fmt.Printf(color.Colorize(color.Red, "ERROR: Max id is (%v)\n"), MaxID)

		// Go to bookmark
		goto loop

	} else if ProjectId < 0 {

		// ERROR message
		fmt.Println(color.Colorize(color.Red, "ERROR: id cant be negative!"))

		// Go to bookmark
		goto loop
	}

	// Find project name by projectid
	CurrentProject := data[id].Projects[ProjectId]

	// Save project name
	pName := CurrentProject.Name

	// Print project name
	fmt.Println(color.Colorize(color.Green, "\n=> Project:"), pName)

	// Print add task commands
	PrintAddTaskCommands()

	loop2 := true

	for loop2 {

		// Get input from user
		command := Get_input(reader)

		// Elapsed time since activity start
		elapsed := time.Since(start)

		//fmt.Printf("1=> ")
		switch command {
		case "back", "b":
			loop2 = false
		case "add", "a":
			AddTask(pName, ProjectId, id)
		case "delete", "del", "d":
			DeleteTask(id, ProjectId)
		case "show", "s":
			ShowTasks(id, ProjectId)
		default:
			ClearScreen()
			fmt.Printf(color.Colorize(color.Green, "---> (%v) Elapsed Time: %v since start [%v] <---\n"), Activity, elapsed, start.Format("15:04:05"))
			PrintAddTaskCommands()

		}
	}
}

// Ask before delete
func DeleteCheckQuestion(name string) bool{

	if name != "" {
		fmt.Printf(color.Colorize(color.Red, "\nDo you really want to delete '%v' ???\n"), name)
	}
	

	// Get reader
	reader := bufio.NewReader(os.Stdin)

	// Get input
	input := Get_input(reader)

	// Check if 'no' is entered
	check := "no" == input

	return check
}

// Add task
func AddTask(pName string, SelectedIdint int, id int) {

	reader := bufio.NewReader(os.Stdin)
	fmt.Println(color.Colorize(color.Green, "\nUpdate or Task name?"))
	fmt.Printf(color.Colorize(color.Green, "=> "))

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
	fmt.Printf(color.Colorize(color.Red, "\n--->> Task %v added to project %v! <<---\n"), tName, pName)
	PrintAddTaskCommands()
}

// Add new Project
func AddNewProject(id int) {

	// Bookmark
loop:

	// Get reader
	reader := bufio.NewReader(os.Stdin)

	// Ask for project name
	fmt.Println(color.Colorize(color.Green, "Project name?"))
	fmt.Printf(color.Colorize(color.Green, "=> "))

	// Save answer
	pName := Get_input(reader)

	// Get data from json
	data := OpenAndGetDataFromJson()

	// Check if project name already exist in db
	for _, value := range data {
		for _, v := range value.Projects {
			switch pName {
			case v.Name:
				fmt.Printf(color.Colorize(color.Red, "\nError: Project %v already exist in db\n\n"), pName)

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

	fmt.Printf("\n--->> Project %v added to db! <<---\n", pName)
}

// Get Top activities
func topActivities(data []JsonData) {

	// Select file
	jq := gojsonq.New().File(filename)

	// Sort by activity, hours and minutes
	top := jq.SortBy("hours", "desc").Only("activity", "hours", "minutes")

	// Pretty print
	b, err := json.MarshalIndent(top, "", "  ")

	// Print error if any
	ErrorHandling(err, "topActivities func")

	// Print activities
	fmt.Print(string(b))

	// Press enter to go back to commandline
	fmt.Println(color.Colorize(color.Green, "\n====> PRESS ENTER TO GO BACK TO COMMANDLINE <===="))

	// Check if enter is pressed
	var command string
	fmt.Scanln(&command)

	// Start commandline
	Commandline()
}

// Save time
func Save_time(reader *bufio.Reader, elapsed time.Duration, id int, PauseTime int) {

	// Print save message
	fmt.Println(color.Colorize(color.Green, "Do you want to save the time? (Press enter or type no)"))
	fmt.Print(color.Colorize(color.Green, "=>  "))

	// Ask before delete
	check := DeleteCheckQuestion("")

	if check {

		// If 'no' is entered tell the user
		fmt.Println(color.Colorize(color.Red, "===>> Last Time NOT SAVED <<==="))

		// Return to commandline
		Commandline()
	} else {

		// Tell the user about saving the time
		fmt.Println(color.Colorize(color.Red, "===>> Last Time has been SAVED <<==="))

		// Save time to db
		UpdateJsonFile(elapsed, id, PauseTime)

		// Return to commandline
		Commandline()
	}
}

// Print commands
func Print_commands() {

	// Print how much time is left till 22:00
	PrintTimeleft()

	// Get data from json
	data := OpenAndGetDataFromJson()

	if len(data) == 0 {

		// Warn if db is empty
		fmt.Println(color.Colorize(color.Red, "<--- WARNING: No data in database --->"))
	} else {

		// Print question
		fmt.Println(color.Colorize(color.Green, "=> What do you want to do now?"))

		// Print all activities
		for _, component := range data {
			fmt.Printf(color.Colorize(color.Yellow, "-> [%vh:%vm] %v || %v (%v) \n"), component.Hours, component.Minutes, component.Activity, component.Short, component.Id)
		}
	}

	// Add main commands
	fmt.Println(color.Colorize(color.Green, "\n=> Commands:"))
	fmt.Println(color.Colorize(color.Yellow, "-> ('top' or 't')"))
	fmt.Println(color.Colorize(color.Yellow, "-> ('add' or 'a')"))
	fmt.Println(color.Colorize(color.Yellow, "-> ('delete' or 'del')"))
	fmt.Println(color.Colorize(color.Yellow, "-> (QUIT PROGRAM: 'q' or '00')"))
	fmt.Print(color.Colorize(color.Green, "\n=> "))
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

		fmt.Printf(color.Colorize(color.Red, "--> ID: %v not found!"), id)

		// Wait for enter
		var command string
		fmt.Scanln(&command)

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
	fmt.Printf(color.Colorize(color.Red, "--> ID: %v Removed!"), id)

	// Press enter to continue
	var command string
	fmt.Scanln(&command)

	// Return to commandline
	Commandline()
}

// Save time function
func UpdateJsonFile(elapsed time.Duration, id int, PauseTime int) {
	// Get data from json
	data := OpenAndGetDataFromJson()

	// Find Index
	index := FindIndexOf(id, data)

	// Check if index exist
	if index == -1 {

		// Tell the user that index does not exist
		fmt.Printf(color.Colorize(color.Red, "--> ID: %v not found!"), id)

		// Return to commandline
		Commandline()
	}

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

// Add new activity to json file
func AddActivity() {

	// Questions array
	questions := []string{"Activity name?", "Activity short name?"}

	// Store answers
	Answers := []string{}

	// Get reader
	reader := bufio.NewReader(os.Stdin)

	// Get data from json
	data := OpenAndGetDataFromJson()

	for _, value := range questions {

		// Bookmark
	loop:

		fmt.Println()
		fmt.Println(color.Colorize(color.Green, value))
		fmt.Print(color.Colorize(color.Green, "=> "))

		// Get answer
		readerAnswer := Get_input(reader)

		// Check if name already exist in db
		for _, value := range data {

			switch readerAnswer {
			case value.Activity, value.Short, "delete", "del", "quit", "q", "add", "a":
				fmt.Printf(color.Colorize(color.Red, "Error: %v already exist in db\n"), readerAnswer)

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

/*<=================================================== Small Help functions ===================================================>*/

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

// Print add tasks commands
func PrintAddTaskCommands() {
	fmt.Println(color.Colorize(color.Cyan, "\n---> (Type 'add' or 'a' to add task)"))
	fmt.Println(color.Colorize(color.Cyan, "---> (Type 'delete', 'del' or 'd' to delete task)"))
	fmt.Println(color.Colorize(color.Cyan, "---> (Type 'show' or 's' to print all tasks)"))
	fmt.Println(color.Colorize(color.Cyan, "---> (Type 'back' or 'b' to leave project)"))
	fmt.Printf(color.Colorize(color.Green, "=> "))
}

// Print projects
func PrintProjects(id int) {

	// Get data from json
	data := OpenAndGetDataFromJson()

	fmt.Println(color.Colorize(color.Green, "\n=> My Projects:"))

	// Print all projects id --> name --> tasks
	for key, value := range data[id].Projects {
		fmt.Printf(color.Colorize(color.Purple, "=> [id: %v] -> '%v' -> (%v)\n"), key, value.Name, len(value.Tasks))
	}
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
	fmt.Print(color.Colorize(color.Red, "--> ID: "))

	// Get reader
	reader := bufio.NewReader(os.Stdin)

	// Get input as string
	GetIdString := Get_input(reader)

	// Convert string to int
	GetId, err := strconv.Atoi(GetIdString)

	// ERROR if a string or a negative number is entered
	if err != nil {

		// Error message
		fmt.Println(color.Colorize(color.Red, "ERROR: ID must be a positive number!"))

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
		fmt.Println(color.Colorize(color.Red, (location+":")), err.Error())
	}
}

// quit
func quit() {
	ClearScreen()
	os.Exit(0)
}
