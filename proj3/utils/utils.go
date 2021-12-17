package utils

import (
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// total number of unique files. If total file to be processed is greater, the program
// will simply cycle through these files, which is computationally equiv to processing different files
const NUM_FILES = 500

// global variables to kee[ track of column numbers
const ZipcodeCol = 0
const WeekStart = 2
const CasesWeek = 4
const TestsWeek = 8
const DeathsWeek = 14

type Arguments struct {
	Zipcode string
	Month   int
	Year    int
}

func ValidateLine(args *Arguments, line []string) bool {
	zipcode := args.Zipcode
	month := args.Month
	year := args.Year

	// check for zipcode
	if strings.Compare(line[ZipcodeCol], zipcode) != 0 {
		return false
	}

	// check for month and year
	time := strings.Split(line[WeekStart], "/")
	dataMonth, _ := strconv.Atoi(time[0])
	dataYear, _ := strconv.Atoi(time[2])
	if dataMonth != month || dataYear != year {
		return false
	}

	// check for miissing value
	if strings.Compare(line[CasesWeek], "") == 0 ||
		strings.Compare(line[TestsWeek], "") == 0 ||
		strings.Compare(line[DeathsWeek], "") == 0 {
		return false
	}

	// all clear
	return true
}

func ParseFile(args *Arguments, fileNum int) map[string][]int {

	// start counter and set up file path
	fileRecords := make(map[string][]int)
	filePath := fmt.Sprintf("../data/covid_%v.csv", fileNum)

	// read the csv file, skip the header
	csvFile, _ := os.Open(filePath)
	defer csvFile.Close()
	csvLines, _ := csv.NewReader(csvFile).ReadAll()
	for _, line := range csvLines {

		if !ValidateLine(args, line) {
			continue
		}

		cases, _ := strconv.Atoi(line[CasesWeek])
		tests, _ := strconv.Atoi(line[TestsWeek])
		deaths, _ := strconv.Atoi(line[DeathsWeek])
		key := fmt.Sprintf("zipcode:%v,time:%v", line[ZipcodeCol], line[WeekStart])
		fileRecords[key] = append(fileRecords[key], cases)
		fileRecords[key] = append(fileRecords[key], tests)
		fileRecords[key] = append(fileRecords[key], deaths)
	}

	return fileRecords
}

func UpdateGlobal(localRecord map[string][]int, globalRecord map[string]bool, totalCases *int, totalTests *int, totalDeaths *int) {
	for key, val := range localRecord {
		if _, contains := globalRecord[key]; contains {
			continue
		} // skip duplicate
		globalRecord[key] = true // add the record
		// add to the tallies
		*totalCases += val[0]
		*totalTests += val[1]
		*totalDeaths += val[2]
	}
}

func GetFileNum(num int) int {
	if num <= NUM_FILES {
		return num
	} else if num%NUM_FILES == 0 {
		return NUM_FILES
	} else {
		return num % NUM_FILES
	}
}
