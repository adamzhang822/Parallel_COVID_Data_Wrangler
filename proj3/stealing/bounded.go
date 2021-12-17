package stealing

import (
	"sync/atomic"
	"unsafe"
)

type Runnable func(arg interface{})

type DEQueue interface {
	PushBottom(task Runnable)
	IsEmpty() bool //returns whether the queue is empty
	PopTop() Runnable
	PopBottom() Runnable
}

type Node struct {
	Payload        Runnable
	Prev           *Node
	Next           *Node
	PositionNumber int
}

/*
The Top pointer will point to the actual top (first element of queue), whereas
BottomSentinel is a sentinel node pointer to a dummy node attached to the actual bottom,
with position number that's one extra than the actual bottom's number right now.
*/
type BoundedDEQueue struct {
	Top            *Node
	BottomSentinel *Node
}

func NewBoundedDEQueue() DEQueue {
	bottomSentinel := Node{Payload: nil, Prev: nil, Next: nil, PositionNumber: 0}
	newDequeue := BoundedDEQueue{Top: nil, BottomSentinel: &bottomSentinel}
	return &newDequeue
}

/*
Assumption:
Based on the instruction for Q2.3 step no.3, which tells us that the main goroutine will fill up
all queues before calling Run for or Exit for any of the workers, I assume that PushBottom can be
thread-unsafe as we are merely filling up each queues sequentially without concurrently dequeuing / enquing it
from either top or bottom side
This also means Top can be nil at initialization and then reference the first node pushed, since
it's guaranteed that we will push something to a queue for a given worker
*/
func (queue *BoundedDEQueue) PushBottom(task Runnable) {
	// If bottom sentinel number is 0, that means queue is empty
	// Top is nil (uninitlaized), so we assign the new node to be the top
	// Rewire the bottom sentinel accordingly (the number held by sentinel should be the next position number to be filled)
	if queue.BottomSentinel.PositionNumber == 0 {
		newBottom := &Node{Payload: task, Next: queue.BottomSentinel, Prev: nil, PositionNumber: 0}
		queue.Top = newBottom
		queue.BottomSentinel.Prev = newBottom
		queue.BottomSentinel.PositionNumber++
		return
	}

	// Otherwise, do regular insert at the end for the double ended queue, top does not change
	currentBottom := queue.BottomSentinel.Prev
	nextPositionNumber := queue.BottomSentinel.PositionNumber
	newBottom := &Node{Payload: task, Next: queue.BottomSentinel, Prev: currentBottom,
		PositionNumber: nextPositionNumber}
	currentBottom.Next = newBottom
	queue.BottomSentinel.Prev = newBottom
	queue.BottomSentinel.PositionNumber += 1
}

/*
Assumption:
IsEmpty() only returns whether the queue is empty at the time the references for Top and Bottom sentinels are gotten.
If function returns empty, the queue will stay empty since no additional tasks will be enqueued
If function returns not empty (false), the thief will move on to steal,
but it could be that when it actually starts stealing, the
victim queue has turned empty since last time it was checked. this case will be dealt by PopTop() separately
*/
func (queue *BoundedDEQueue) IsEmpty() bool {
	return queue.BottomSentinel.PositionNumber <= queue.Top.PositionNumber
}

/*
In this implementation Top pointer will only go up in number (go down in queue towards bottom)
If PopBottom() detects that after its operation the queue becomes empty,
it will set Top to a dummy node with position number 999.
This is valid for our program, because we know once queue is emptied, it will not be refilled,
and we also know that there are only 500 tasks in total, so no matter what thread count we have,
position number 999 will be the end of the queue
*/
func (queue *BoundedDEQueue) PopTop() Runnable {
	oldTop := queue.Top
	newTop := oldTop.Next
	oldTopNumber := oldTop.PositionNumber
	task := oldTop.Payload
	if queue.BottomSentinel.PositionNumber <= oldTopNumber {
		return nil
	}
	if atomic.CompareAndSwapPointer((*unsafe.Pointer)(unsafe.Pointer(&queue.Top)),
		unsafe.Pointer(oldTop), unsafe.Pointer(newTop)) {
		return task
	}
	return nil
}

/*
Similar to implementation of the book, when PopBottom() detects that it's trying to
dequeue the last node in queue, it enters into safety mode by trying to also reset the
Top pointer atomically.
In this implementation, we do not have to reset BottomSentinel or Top to 0, since
once queue is emptied, it will never be refilled.
Therefore, it is sufficient that the Top be set to a number such that any thief trying to
steal from the queue will be notified that the Top number is large enough to indicate that
the queue has been emptied. We only have 500 files, so 999 will work.
Furthermore, we know that if PopBottom() fails, it must be that PopTop() has stolen the last
task, so once PopBottom() returns nil, it must be that the queue is emptied and will never be refilled.
*/
func (queue *BoundedDEQueue) PopBottom() Runnable {
	if queue.BottomSentinel.PositionNumber == 0 {
		return nil
	}
	bottom := queue.BottomSentinel.Prev
	queue.BottomSentinel = bottom
	task := queue.BottomSentinel.Payload
	oldTop := queue.Top
	newTop := &Node{PositionNumber: 999}
	oldTopNumber := oldTop.PositionNumber
	if bottom.PositionNumber > oldTopNumber {
		return task
	}
	if bottom.PositionNumber == oldTopNumber {
		if atomic.CompareAndSwapPointer((*unsafe.Pointer)(unsafe.Pointer(&queue.Top)),
			unsafe.Pointer(oldTop), unsafe.Pointer(newTop)) {
			return task
		}
	}
	return nil
}
