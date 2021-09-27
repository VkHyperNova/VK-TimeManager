package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"os"
	"os/exec"
	"strings"
	"time"
	"github.com/thedevsaddam/gojsonq"
)

type JsonData struct {
	Id          int    `json:"id"`
	Activity    string `json:"activity"`
	Short    	string `json:"short"`
	Hours   	int `json:"hours"`
	Minutes   	int `json:"minutes"`
	Function   	string `json:"function"`
	
}
var ProgramVersion = "1.16" // Update version [16 updates]
var filename = "data/data.json"
 
//go:generate goversioninfo -icon=resource/timem.ico -manifest=resource/goversioninfo.exe.manifest
 
func main() {

	
	MakeDirAndJson()
	Commandline()
}


/*<=================================================== Main functions ===================================================>*/

// Calculate time left
func PrintTimeleft(){

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
	MinutesLeft := MinutesTillEndTime - (HoursLeft*60)

	
	fmt.Printf("\n<================== VK TimeManager v%v ==================>\n", ProgramVersion)
	fmt.Printf("\n<--- You have %v hours and %v minutes left --->\n\n", HoursLeft, MinutesLeft)
}

// Command line
func Commandline(){



	// Print commands to console
	Print_commands()

	// Add default commands to data
	data := AddDefaultCommands()

	reader := bufio.NewReader(os.Stdin)

	for true {

		var command string
		fmt.Scanln(&command)

		MainSwitch(command, data, reader)

	}	
}


// Add default commands to data temporarly
func AddDefaultCommands() []JsonData{
	// Get data from json
	data := OpenAndGetDataFromJson()

	GetId := false

	if len(data) != 0 {
		GetId = true
	}

	// Make map of maps
	DefaultCommands := make(map[int]map[string]string)

	// Add to map
	DefaultCommands[1] = map[string]string{}
	DefaultCommands[1]["name"] = "add"
	DefaultCommands[1]["short"] = "a"

	DefaultCommands[2] = map[string]string{}
	DefaultCommands[2]["name"] = "delete"
	DefaultCommands[2]["short"] = "del"

	DefaultCommands[3] = map[string]string{}
	DefaultCommands[3]["name"] = "quit"
	DefaultCommands[3]["short"] = "q"

	// Append to data 
	for _, value := range DefaultCommands {
		NewValues := ConvertAnswersToJsonData(value["name"], value["short"], GetId, value["name"])
		data = append(data, NewValues)

	}

	return data
}

// MainSwitch
func MainSwitch(command string, data []JsonData, reader *bufio.Reader){

	for _, value := range data {
		switch command {
		case value.Activity, value.Short, fmt.Sprint(value.Id):

			if value.Function == "add" {
				AddActivity()
			} else if value.Function == "delete" {
				DeleteActivity()
			} else if value.Function == "quit" {
				quit()
			} else {
				ClearScreen()
				start := time.Now()
				Loop_for_answer(reader, start, value.Activity, value.Id)	
			}
		}
	}
}

// The loop
func Loop_for_answer(reader *bufio.Reader, start time.Time, Activity string, id int){

	// Get data from json
	data := OpenAndGetDataFromJson()


	hours := data[id].Hours
	minutes := data[id].Minutes

	fmt.Println()
	fmt.Printf("<--- Starting %v at %v [%v hours %v minutes] --->\n", Activity, start.Format("02.01.2006 15:04:05"), hours, minutes)
	


	
	
	loop := true

	for loop {
		
		fmt.Println("\nType 'done' or '0' or 'q'")
		fmt.Print("=>  ")

		input := Get_input(reader)

		chk := "done" == input || "0" == input || "q" == input

		elapsed := time.Since(start)

		if chk {
			//elapsed := time.Since(start)

			fmt.Printf("You have spent %v", elapsed)
			fmt.Println()
			

			// Ask for save time
			Save_time(reader, elapsed, id)

			loop = false
		} 
		ClearScreen()
		fmt.Printf("---> Time: %v <---", elapsed)
		fmt.Println()

		
	}
}

// Save time
func Save_time(reader *bufio.Reader, elapsed time.Duration, id int){
	
	// Save
	fmt.Println("Do you want to save the time? (Press enter or type no)")
	fmt.Print("=>  ")

	input := Get_input(reader)

	check := "no" == input

	if check {
		fmt.Println("===>> Last Time NOT SAVED <<===")
		Commandline()
	} else {
		fmt.Println("===>> Last Time has been SAVED <<===")
		UpdateJsonFile(elapsed, id)
		Commandline()
	}
}

// Print commands
func Print_commands(){


	

	PrintTimeleft()

	// Get data from json
	data := OpenAndGetDataFromJson()


	if len(data) == 0 {
		fmt.Println("<--- WARNING: No data in database --->")
	} else {
		// Print info
		for _, component := range data {
			fmt.Printf("-> [%vh:%vm] %v || %v (%v) \n", component.Hours, component.Minutes, component.Activity, component.Short, component.Id)		
		}		
	}

	
	fmt.Println()
	fmt.Println("-> add or a")
	fmt.Println("-> delete or del")
	fmt.Println("-> quit or q")

	fmt.Println()
	fmt.Print("=> ")
}

// Delete activity
func DeleteActivity(){
	// AskAndStoreQuestions for id
	id := AskForId()

	// Clear the screen
	ClearScreen()

	// Get data from json
	data := OpenAndGetDataFromJson()

	// Find Index
	index := FindIndexOf(id, data)

	// Check if index exist
	if index == -1 {
		fmt.Println("--> ID:", id, "not found!")
		Commandline()
	}

	// Delete
	data = append(data[:index], data[index+1:]...)

	// Convert it back to byte
	dataBytes := MarshalIndentToByte(data, "DeleteItem")

	// Override json file with updated data
	WriteToFile(dataBytes)

	Commandline()
}

// Save time function
func UpdateJsonFile(elapsed time.Duration, id int){
	// Get data from json
	data := OpenAndGetDataFromJson()

	// Find Index
	index := FindIndexOf(id, data)

	// Check if index exist
	if index == -1 {
		fmt.Println("--> ID:", id, "not found!")
		Commandline()
	}
	

	//Old + new Minutes
	NewMinutes := int(float64(data[id].Minutes) + math.Round(elapsed.Minutes()))

	// Get hours out of all minutes
	GetHours := NewMinutes/60

	// Remove hours and get minutes left
	GetMinutes := NewMinutes - (GetHours * 60)

	// Add new data
	HoursToAdd := data[id].Hours + GetHours
	MinutesToAdd := GetMinutes
	
	// Add new time to db
	data[id].Minutes = MinutesToAdd
	data[id].Hours = HoursToAdd

	// Convert it back to byte
	dataBytes := MarshalIndentToByte(data, "UpdateItem")

	// Override json file with updated data
	WriteToFile(dataBytes)		
}

// Add new activity to json file
func AddActivity(){

	questions := []string{"Activity name?","Activity short name?"}
	Answers := []string{}
	reader := bufio.NewReader(os.Stdin)
	// Get data from json
	data := OpenAndGetDataFromJson()

	for _, value := range questions{
		// Defined a label named "loop" 
		loop:
		fmt.Println()
		fmt.Println(value)
		fmt.Print("=> ")
		// Read the answer
		readerAnswer, _ := reader.ReadString('\n')
		// convert CRLF to LF
		readerAnswer = strings.Replace(readerAnswer, "\r\n", "", -1)

		for _, value := range data {

			switch readerAnswer {
			case value.Activity, value.Short, "delete", "del", "quit", "q", "add", "a":
				fmt.Printf("Error: %v already exist in db\n", readerAnswer)

				// Restart the for loop, go to back to loop label
				goto loop
			}
			
			 
		}

		Answers = append(Answers, readerAnswer)

		
	}	
	

	GetId := false

	if len(data) != 0 {
		GetId = true
	}

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


// Open and get data
func OpenAndGetDataFromJson() []JsonData {
	// Open json file
	file := ReadFile()

	// Get data and store it in data variable
	data := ConvertToJsonDataStruct(file)

	return data
}

// Make dir and data.json if not exist
func MakeDirAndJson(){
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
	fmt.Print("--> ID: ")
	var ID int
	fmt.Scanln(&ID)

	return ID
}

// Encrypt new data and construct a WebsiteData struct for adding it to json file
func ConvertAnswersToJsonData(Activity_Name string, Activity_Name_short string, GetLastid bool, function string) JsonData {

	// Declare variables
	Id := 0

	if GetLastid {
		Id = GetLastId(filename)
	}

	var ValuesToAdd JsonData

	ValuesToAdd = JsonData{
		Id: Id,
		Activity: Activity_Name,
		Short: Activity_Name_short,
		Hours: 0,
		Minutes: 0,
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
		fmt.Println(location+":", err.Error())
	}
}

// quit
func quit(){
	ClearScreen()
	os.Exit(0)		
}