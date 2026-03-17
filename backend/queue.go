package checkered

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
  - value T: The value to add
*/
func (queue *Queue[T]) enqueue(value T) {
	if queue.size == queue.maxSize {
		panic("Queue Overflow!")
	}
	queue.arr[queue.tail] = value
	queue.tail = (queue.tail + 1) % queue.maxSize
	queue.size++
}

/*
Remove an element from the queue
Args:
  - queue *Queue[T]: Pointer to the queue to remove an element from

Returns:
  - T: The element removed
*/
func (queue *Queue[T]) dequeue() T {
	if queue.size == 0 {
		panic("Queue Empty!")
	}
	res := queue.arr[queue.head]
	queue.head = (queue.head + 1) % queue.maxSize
	queue.size--
	return res
}

func (queue *Queue[T]) forEach(callback func(T, int)) {
	for i := 0; i < queue.size; i++ {
		callback(queue.arr[(queue.head+i)%queue.maxSize], i)
	}
}

func RemoveValue(q *Queue[*Client], value *Client) {
	originalSize := q.size

	for i := 0; i < originalSize; i++ {
		v := q.dequeue()

		if v.username != value.username {
			q.enqueue(v)
		}
	}
}
