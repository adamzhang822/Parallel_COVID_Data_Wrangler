package modes

import (
	"fmt"
	"proj3/utils"
	"sync"
	"sync/atomic"
)

type WorkerContext struct {
	totalCases  int
	totalTests  int
	totalDeaths int
	records     map[string]bool
	flag        int32
	group       *sync.WaitGroup
	args        *utils.Arguments
}

func worker(context *WorkerContext, args *utils.Arguments, start int, end int) {

	// compute the total cases, tests, and deaths for the portion assigned
	workerRecords := make(map[string][]int)

	for i := start; i <= end; i++ {
		fileNum := utils.GetFileNum(i)
		fileRecords := utils.ParseFile(args, fileNum)
		for key, val := range fileRecords {
			if _, contains := workerRecords[key]; contains {
				continue
			} // skip duplicate
			workerRecords[key] = val
		}
	}

	// enter the critical section by updating the global values
	// exit the critical section after finishing work
	// use a TTAS lock
	for true {
		for context.flag == 1 {
		} // spin while lock is taken
		if atomic.CompareAndSwapInt32(&(context.flag), 0, 1) {
			utils.UpdateGlobal(workerRecords, context.records, &context.totalCases, &context.totalTests, &context.totalDeaths)
			atomic.StoreInt32(&(context.flag), 0)
			context.group.Done()
			return
		}
	}
}

func RunStatic(args *utils.Arguments, size int, numThreads int) {
	// Parallel mode:
	var group sync.WaitGroup
	context := WorkerContext{group: &group}
	context.records = make(map[string]bool)
	workAmount := size / numThreads // static distribution
	remWork := size % numThreads    // last thread does extra work

	for i := 0; i < numThreads; i++ {
		context.group.Add(1)
		startPt := i*workAmount + 1
		endPt := (i + 1) * workAmount
		if i == numThreads-1 {
			endPt += remWork
		}
		go worker(&context, args, startPt, endPt)
	}
	group.Wait()

	// final processing of the result and print out to console
	result := fmt.Sprintf("%v,%v,%v", context.totalCases, context.totalTests, context.totalDeaths)
	fmt.Println(result)
}
