package modes

import (
	"fmt"
	"proj3/utils"
	"sync"
)

type BSPContext struct {
	// Number of Threads
	numThreads int

	// Keep track of current task and effect
	iterIdx      int // idx of current iteration of superstep - synchronization step
	numTasks     int // total number of tasks
	localRecords []map[string][]int
	args         *utils.Arguments

	// For global synchronization
	globalRecords map[string]bool
	totalCases    int
	totalTests    int
	totalDeaths   int
	mutex         *sync.Mutex
	cond          *sync.Cond
	workersIdle   int
	synchronizing bool
	done          bool
}

func initBSPContext(numThreads int, args *utils.Arguments, size int) *BSPContext {

	// Initialize the basic task information (threads, number of tasks )
	newContext := &BSPContext{numThreads: numThreads, args: args, iterIdx: 0, totalCases: 0, totalTests: 0, totalDeaths: 0}
	newContext.numTasks = size
	// Initialize the local records slice for workers
	localRecords := make([]map[string][]int, numThreads)
	newContext.localRecords = localRecords
	newContext.globalRecords = make(map[string]bool)

	// Initialize the synchronization parameters
	var mutex sync.Mutex
	cond := sync.NewCond(&mutex)
	newContext.cond = cond
	newContext.mutex = &mutex
	newContext.workersIdle = 0
	newContext.synchronizing = false
	newContext.done = false

	return newContext
}

func synchronize(ctx *BSPContext) {
	ctx.mutex.Lock()
	for !ctx.synchronizing {
		ctx.cond.Wait()
	}

	// Update all the records
	for i := 0; i < ctx.numThreads; i++ {
		utils.UpdateGlobal(ctx.localRecords[i], ctx.globalRecords, &ctx.totalCases, &ctx.totalTests, &ctx.totalDeaths)
	}

	// update idx for next iter:
	ctx.iterIdx += 1

	// terminate if exceeds the quota
	if ctx.iterIdx*ctx.numThreads > ctx.numTasks {
		ctx.done = true
		ctx.mutex.Unlock()
		return
	}

	ctx.synchronizing = false
	ctx.mutex.Unlock()

	ctx.cond.Broadcast() // signal all workers to wake up for next round

}

func superStep(idx int, ctx *BSPContext) {

	// Get the correct subtask for this superstep on this worker
	curIterIdx := ctx.iterIdx
	fileIdx := (curIterIdx*ctx.numThreads + (idx + 1))

	if fileIdx > ctx.numTasks {
		ctx.localRecords[idx] = make(map[string][]int)
	} else {
		fileNum := utils.GetFileNum(fileIdx)
		ctx.localRecords[idx] = utils.ParseFile(ctx.args, fileNum)
	}

	// Synchronize
	ctx.mutex.Lock()
	ctx.workersIdle++
	if ctx.workersIdle == ctx.numThreads {
		ctx.synchronizing = true
		ctx.cond.Broadcast() // wake up synchronizer routine
	}
	for ctx.synchronizing || (ctx.iterIdx == curIterIdx) {
		ctx.cond.Wait()
	}
	ctx.workersIdle--
	ctx.mutex.Unlock()
}

func ExecuteBSP(idx int, ctx *BSPContext) {
	for {
		if ctx.done {
			return
		}
		if idx == ctx.numThreads {
			synchronize(ctx)
		} else {
			superStep(idx, ctx)
		}
	}
}

func RunBSP(numThreads int, args *utils.Arguments, size int) {
	ctx := initBSPContext(numThreads-1, args, size) // Initialize your BSP context
	for idx := 0; idx < numThreads-1; idx++ {
		go ExecuteBSP(idx, ctx)
	}
	ExecuteBSP(numThreads-1, ctx)
	// final processing of the result and print out to console
	result := fmt.Sprintf("%v,%v,%v", ctx.totalCases, ctx.totalTests, ctx.totalDeaths)
	fmt.Println(result)
}
