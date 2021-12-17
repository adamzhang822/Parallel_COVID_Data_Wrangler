package main

import (
	"fmt"
	"os"
	"proj3/modes"
	"proj3/utils"
	"strconv"
	"strings"
)

func validArgs(args []string) bool {
	if len(args) < 6 {
		return false
	}
	if strings.Compare(args[0], "bsp") != 0 && strings.Compare(args[0], "static") != 0 && strings.Compare(args[0], "stealing") != 0 {
		return false
	}
	numThreads, _ := strconv.Atoi(args[2])
	if numThreads < 0 {
		return false
	}
	if strings.Compare(args[0], "bsp") == 0 && numThreads <= 2 {
		return false
	}
	size, _ := strconv.Atoi(args[1])
	if size != 500 && size != 1000 && size != 3000 {
		return false
	}
	month, _ := strconv.Atoi(args[4])
	year, _ := strconv.Atoi(args[5])
	if month < 1 || month > 12 {
		return false
	}
	if year != 2020 && year != 2021 {
		return false
	}
	return true
}

func main() {

	const usage = "Usage:	go run proj3/covid mode size threads zipcode month year\n" +
		"	mode = either 'static' or 'stealing' or 'bsp'\n" +
		"	size = 500 or 1000 or 3000, the number of files to be processed\n" +
		"	threads = the number of threads (i.e., goroutines to spawn). If bsp, must be > 2 \n" +
		"	to run sequential mode, specify thread = 0 when the mode is either static or stealing\n" +
		"	zipcode = a possible Chicago zipcode\n" +
		"	month = the month to display for that zipcode, must be between 1-12 \n" +
		"	year  = the year to display for that zipcode, must be 2020 or 2021 \n"

	// Parse the arguments and check if they are valid
	args := os.Args[1:]
	if !validArgs(args) {
		fmt.Println(usage)
		return
	}
	mode := args[0]
	size, _ := strconv.Atoi(args[1])
	numThreads, _ := strconv.Atoi(args[2])
	zipcode := args[3]
	month, _ := strconv.Atoi(args[4])
	year, _ := strconv.Atoi(args[5])

	// Set up worker contexts
	arguments := utils.Arguments{Zipcode: zipcode, Month: month, Year: year}

	// Sequential mode:
	if numThreads == 0 {
		modes.RunSequential(&arguments, size)
		return
	}

	// Static distribution mode:
	if strings.Compare(mode, "static") == 0 {
		modes.RunStatic(&arguments, size, numThreads)
		return
	}

	// Work stealing mode:
	if strings.Compare(mode, "stealing") == 0 {
		modes.RunStealing(&arguments, size, numThreads)
		return
	}

	// BSP mode (thread number must be > 2):
	if strings.Compare(mode, "bsp") == 0 {
		modes.RunBSP(numThreads, &arguments, size)
		return
	}

}
