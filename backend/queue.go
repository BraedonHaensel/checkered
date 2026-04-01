package checkered

/*
Cyclic queue with fixed capacity
Use:
  - `InitQueue` to properly initialize the queue
  - `Enqueue` to add elements to the queue; and
  - `Dequeue` to remove elements from the queue
*/
type Queue[T any] struct {
	MaxSize int `json:"max_size"` // The maximum capacity of the queue
	Size    int `json:"size"`     // The current size of the queue (i.e. the number of elements in the queue)
	Head    int `json:"head"`     // The pointer to the head of the queue
	Tail    int `json:"tail"`     // The pointer to the tail of the queue
	Arr     []T `json:"data"`     // The underlying container of the queue
}

/*
Initialize a queue
Args:
  - queue *Queue[T]: Pointer to the queue to initialize
  - size int: The max size of the queue
*/
func InitQueue[T any](queue *Queue[T], size int) {
	queue.Arr = make([]T, size)
	queue.MaxSize = size
	queue.Size = 0
	queue.Head = 0
	queue.Tail = 0
}

/*
Add an element to the queue
Args:
  - value T: The value to add
*/
func (queue *Queue[T]) enqueue(value T) {
	if queue.Size == queue.MaxSize {
		panic("Queue Overflow!")
	}
	queue.Arr[queue.Tail] = value
	queue.Tail = (queue.Tail + 1) % queue.MaxSize
	queue.Size++
}

/*
Remove an element from the queue
Args:
  - queue *Queue[T]: Pointer to the queue to remove an element from

Returns:
  - T: The element removed
*/
func (queue *Queue[T]) dequeue() T {
	if queue.Size == 0 {
		panic("Queue Empty!")
	}
	res := queue.Arr[queue.Head]
	queue.Head = (queue.Head + 1) % queue.MaxSize
	queue.Size--
	return res
}

func (queue *Queue[T]) forEach(callback func(T, int)) {
	for i := 0; i < queue.Size; i++ {
		callback(queue.Arr[(queue.Head+i)%queue.MaxSize], i)
	}
}

func RemoveValue(q *Queue[*Client], value *Client) {
	originalSize := q.Size

	for i := 0; i < originalSize; i++ {
		v := q.dequeue()

		if v.username != value.username {
			q.enqueue(v)
		}
	}
}
