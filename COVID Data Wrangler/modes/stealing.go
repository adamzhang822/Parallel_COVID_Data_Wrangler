package modes

import (
	"fmt"
	"proj3/stealing"
	"proj3/utils"
	"sync"
	"sync/atomic"
)

func generateTask(fileNum int) func(interface{}) {

	return func(arg interface{}) {
		ctx := arg.(*stealing.StealingWorkerContext)
		args := ctx.Args
		// start counter and set up file path
		fileRecords := utils.ParseFile(args, fileNum)
		// finished parsing the file, try to update the global context
		// enter the critical section by updating the global values
		// exit the critical section after finishing work
		// use a TTAS lock
		for true {
			for ctx.Flag == 1 {
			} // spin while lock is taken
			if atomic.CompareAndSwapInt32(&(ctx.Flag), 0, 1) {
				utils.UpdateGlobal(fileRecords, ctx.Records, &ctx.TotalCases, &ctx.TotalTests, &ctx.TotalDeaths)
				atomic.StoreInt32(&(ctx.Flag), 0)
				return
			}
		}
	}
}

func RunStealing(args *utils.Arguments, size int, numThreads int) {
	// Parallel mode:
	/*
		Assumptions:
		The runnable tasks consist of parsing a file with specified file number, creating a local record,
		and then update the global record.

		At first step, we will have  static distribution (which will be distributed in run-time)
		by enqueing roughly equal number of runnables to each thread's local queue before
		calling them to run.

		During run-time, if for some reason some thread finishes earlier than others, it will attempt
		to steal from others so that it doesn't lay idle.

		For Step 3 and Step 4, we are implementing mechanisms for the program to wait until everything
		has been completed. Calling Exit() will notify each worker that there will be no more works
		enqueued to its local queues so it should focus on stealing work.
		Threads will keep attempting work stealing until it is notified that all other threads are
		also emptied. This is done by each thread atomically updating the global context's numEmptied
		number when its local queue is exhausted, and letting the loop in Run() function detect that change.
	*/
	// Step 0: Initialize the global context
	var group sync.WaitGroup
	context := stealing.StealingWorkerContext{Group: &group}
	context.Group.Add(numThreads)
	context.Records = make(map[string]bool)
	context.Args = args
	context.NumThreads = int32(numThreads)
	context.Queues = make([]stealing.DEQueue, numThreads)
	context.Workers = make([]*stealing.StealingWorker, numThreads)

	// Step 1: Initializing the stealing workers and their queues and filling them up

	workAmount := size / numThreads
	remWork := size % numThreads
	for i := 0; i < numThreads; i++ {
		startPt := i*workAmount + 1
		endPt := (i + 1) * workAmount
		if i == numThreads-1 {
			endPt += remWork
		}
		context.Queues[i] = stealing.NewBoundedDEQueue()
		for j := startPt; j <= endPt; j++ {
			fileNum := utils.GetFileNum(j)
			task := generateTask(fileNum)
			context.Queues[i].PushBottom(task)
		}
		context.Workers[i] = stealing.NewStealingWorker(i, &context, context.Queues, i)
	}

	// Step 2: Call Run on each of the worker inside the workers slice
	for i := 0; i < numThreads; i++ {
		go context.Workers[i].Run()
	}

	// Step 3: Call Exit on each of the workers after distributing all works
	for i := 0; i < numThreads; i++ {
		context.Workers[i].Exit()
	}

	// Step 4: Wait till all workers have completed
	group.Wait()
	// final processing of the result and print out to console
	result := fmt.Sprintf("%v,%v,%v", context.TotalCases, context.TotalTests, context.TotalDeaths)
	fmt.Println(result)
	return
}
