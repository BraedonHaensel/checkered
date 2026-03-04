package main

/*
Cyclic queue with fixed capacity
Use:
	- `InitQueue` to properly initialize the queue
	- `Enqueue` to add elements to the queue; and
	- `Dequeue` to remove elements from the queue
*/
type Queue[T any] struct {
	maxSize int // The maximum capacity of the queue
	size    int // The current size of the queue (i.e. the number of elements in the queue)
	head    int // The pointer to the head of the queue
	tail    int // The pointer to the tail of the queue
	arr     []T // The underlying container of the queue
}

/*
Initialize a queue
Args:
	- queue *Queue[T]: Pointer to the queue to initialize
	- size int: The max size of the queue
*/
func InitQueue[T any](queue *Queue[T], size int) {
	queue.arr = make([]T, size)
	queue.maxSize = size
	queue.size = 0
	queue.head = 0
	queue.tail = 0
}

/*
Add an element to the queue
Args:
	- queue *Queue[T]: Pointer to the queue to add the element to
	- value T: The value to add
*/
func Enqueue[T any](queue *Queue[T], value T) {
	if queue.size == queue.maxSize {
		panic("Queue Overflow!")
	}
	queue.arr[queue.tail] = value
	queue.tail++
	queue.size++
}

/*
Remove an element from the queue
Args:
	- queue *Queue[T]: Pointer to the queue to remove an element from
Returns:
	- T: The element removed
*/
func Dequeue[T any](queue *Queue[T]) T {
	if queue.size == 0 {
		panic("Queue Empty!")
	}
	res := queue.arr[queue.head]
	queue.head++
	queue.size--
	return res
}
