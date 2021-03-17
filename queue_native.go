package queue

import (
	"errors"
	"fmt"
	"sync"
)

// NativeIntQ is a type of queue that uses a
// condition variable and a circular buffer
// BoundedQueue interface. this implementation
// is intended to be thread safe
type NativeIntQ struct {
	queue []int
	head     int
	tail     int
	length   int
	capacity int
	mtx sync.Mutex      // a mutex for mutual exclusion
	putcv *sync.Cond    // a condition variable for controlling Puts
	getcv *sync.Cond    // a condition variable for controlling Gets
}

// TryPut adds an element onto the tail queue
// if the queue is full, an error is returned
func (nvq *NativeIntQ) TryPut(value int) error {
	// lock the mutex
	nvq.putcv.L.Lock();
	defer nvq.putcv.L.Unlock()

	// is queue full ?
	if nvq.length == nvq.capacity {
		// return an error
		e := errors.New("queue is full")
		return e;
	}

	// queue had room, add it at the tail
	nvq.queue[nvq.tail] = value
	nvq.tail = (nvq.tail+1) % nvq.capacity
	nvq.length++

	// signal a Get to wake up
	nvq.getcv.Signal()

	// no error
	return nil
} 

// Put adds an element onto the tail queue
// if the queue is full the function blocks
func (nvq *NativeIntQ) Put(value int)  {
	// lock the mutex
	nvq.putcv.L.Lock()
	defer nvq.putcv.L.Unlock()

	// block until a value is in the queue
	for nvq.length == nvq.capacity {
		// release and wait
		nvq.putcv.Wait()
	}
	
	// queue has room, add it at the tail
	nvq.queue[nvq.tail] = value
	nvq.tail = (nvq.tail+1) % nvq.capacity
	nvq.length++

	// signal a Get to wake up
	nvq.getcv.Signal()
} 


// Get returns an element from the head of the queue
// if the queue is empty,the caller blocks
func (nvq *NativeIntQ) Get() int {
	// lock the mutex
	nvq.getcv.L.Lock()
	defer nvq.getcv.L.Unlock()

	// block until a value is in the queue
	for nvq.length == 0 {
		// release and wait
		nvq.getcv.Wait()
	}

	// at this point there is at least one item in the queue
	// remove the head
	value := nvq.queue[nvq.head]
	nvq.head = (nvq.head + 1)  % nvq.capacity
	nvq.length--

	// signal a Put to wake up
	nvq.putcv.Signal()

	return value
}

// TryGet gets a value or returns an error if the queue is empty
func (nvq *NativeIntQ) TryGet() (int, error) {
	var value int
	var err error

	// lock the mutex
	nvq.getcv.L.Lock()
	defer nvq.getcv.L.Unlock()

	// is the queue empty?
	if nvq.length > 0 {
		value = nvq.queue[nvq.head]
		nvq.head = (nvq.head + 1)  % nvq.capacity
		nvq.length--
	} else {
		value = 0
		err = errors.New("queue is empty");
	}
	
	// signal a Put to wake up
	nvq.putcv.Signal()

	return value, err
	
}

// Len is the current number of elements in the queue 
func (nvq *NativeIntQ) Len() int {
	return nvq.length
}

// Cap is the maximum number of elements the queue can hold
func (nvq *NativeIntQ) Cap() int {
	return cap(nvq.queue)
}

// Close is for cleanup
func (nvq *NativeIntQ) Close() {
	// noop
}

// String
func (nvq *NativeIntQ) String() string {
	return fmt.Sprintf("Native Len:%v Cap:%v",nvq.Len(),nvq.Cap())
}

// NewNativeQueue is a factory for creating bounded queues
// that use a condition variable and circular buffer. It returns
// an instance of pointer to BoundedQueue
func NewNativeQueue(size int) *NativeIntQ {
	var nvq NativeIntQ
	
	// allocate the whole slice during init
	nvq.queue = make([]int,size,size)
	nvq.head = 0
	nvq.tail = 0
	nvq.length = 0
	nvq.capacity = size
	nvq.mtx = sync.Mutex{} 
	nvq.putcv = sync.NewCond(&nvq.mtx)
	nvq.getcv = sync.NewCond(&nvq.mtx)

	return &nvq
}
