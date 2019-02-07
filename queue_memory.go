package goworker

import (
	"errors"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// MemoryQueue
type memQueue struct {
	storage chan interface{}
	close   chan os.Signal
}

func (m *memQueue) Acknowledge(message interface{}) {
	return
}

func NewMemoryQueue(length int) Queue {
	q := &memQueue{
		storage: make(chan interface{}, length),
		close:   make(chan os.Signal),
	}

	signal.Notify(q.close, syscall.SIGTERM)

	return q
}

func NewMemoryQueueFactory(length int) QueueFactory {
	return func(name string) (queue Queue, e error) {
		return NewMemoryQueue(length), nil
	}
}

func (m *memQueue) Push(entry interface{}, timeout time.Duration) error {
	timeoutReached := time.After(timeout)
	if timeout == 0 {
		timeoutReached = make(chan time.Time)
	}
	select {
	case m.storage <- entry:
	case <-m.close:
		close(m.storage)
	case <-timeoutReached:
		return errors.New("timeout while pushing to queueManager")
	}

	return nil
}

func (m *memQueue) Pop(timeout time.Duration) (interface{}, error) {
	timeoutReached := time.After(timeout)
	if timeout == 0 {
		timeoutReached = make(chan time.Time)
	}
	select {
	case e := <-m.storage:
		return e, nil
	case <-m.close:
		close(m.storage)
	case <-timeoutReached:
		return nil, errors.New("timeout while poping queueManager")
	}

	return nil, nil
}

func (m *memQueue) Channel() chan interface{} {
	return m.storage
}

func (m *memQueue) Shutdown() error {
	close(m.storage)
	return nil
}