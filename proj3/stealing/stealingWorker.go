package stealing

import (
	"math/rand"
	"proj3/utils"
	"runtime"
	"sync"
	"sync/atomic"
)

type StealingWorkerContext struct {
	TotalCases  int
	TotalTests  int
	TotalDeaths int
	Records     map[string]bool
	Flag        int32
	Group       *sync.WaitGroup
	Args        *utils.Arguments
	Queues      []DEQueue
	Workers     []*StealingWorker
	NumEmptied  int32
	NumThreads  int32
}

type StealingWorker struct {
	LocalQueue DEQueue
	ID         int
	Emptied    bool
	NoMoreTask bool
	Ctx        *StealingWorkerContext
}

func NewStealingWorker(assignedQueue int, ctx interface{},
	queues []DEQueue, id int) *StealingWorker {
	newWorker := StealingWorker{LocalQueue: queues[assignedQueue], Ctx: ctx.(*StealingWorkerContext), ID: id}
	return &newWorker
}

func (worker *StealingWorker) Run() {
	// Loop runs when there are more tasks to be enqueued to local queue
	// Or when there are still other threads with work left to steal
	for !worker.NoMoreTask || worker.Ctx.NumEmptied < worker.Ctx.NumThreads {
		// Still has own tasks to do, so do not steal
		if !worker.Emptied {
			task := worker.LocalQueue.PopBottom()
			if task == nil {
				// emptied
				atomic.AddInt32(&worker.Ctx.NumEmptied, 1)
				worker.Emptied = true
			} else {
				// Execute the task
				task(worker.Ctx)
			}
		} else {
			// No more work in local queue, try to steal after yielding CPU for any other
			// thread who hasn't finished
			runtime.Gosched()
			victimID := rand.Intn(int(worker.Ctx.NumThreads))
			if victimID == worker.ID || worker.Ctx.Queues[victimID].IsEmpty() {
				continue
			}
			task := worker.Ctx.Queues[victimID].PopTop()
			if task != nil {
				task(worker.Ctx)
			}
		}
	}
	// Exit the loop after getting signal to exit, and calling Done on waitgroup
	worker.Ctx.Group.Done()
}

/*
In our implementation Exit() doesn't do much, because the Workers already know that once
queue is emptied, it will not be refilled. If we have an implementation where new tasks come into
the queues concurrently, then we would need this function.
*/
func (worker *StealingWorker) Exit() {
	worker.NoMoreTask = true
}
