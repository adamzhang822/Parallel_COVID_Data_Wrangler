package modes

import (
	"fmt"
	"proj3/utils"
)

func RunSequential(args *utils.Arguments, size int) {
	totalCases := 0
	totalTests := 0
	totalDeaths := 0
	allRecords := make(map[string]bool)

	for i := 1; i <= size; i++ {
		fileNum := utils.GetFileNum(i)
		fileRecord := utils.ParseFile(args, fileNum)
		utils.UpdateGlobal(fileRecord, allRecords, &totalCases, &totalTests, &totalDeaths)
	}
	result := fmt.Sprintf("%v,%v,%v", totalCases, totalTests, totalDeaths)
	fmt.Println(result)
	return
}
