package thread

import "sync"

type Counter struct {
	value uint
	mutex *sync.Mutex
}

func NewCounter() *Counter {
	return &Counter{0, &sync.Mutex{}}
}

func (c *Counter) Inc() uint {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	currentValue := c.value
	c.value += 1

	return currentValue
}

func (c *Counter) Get() uint {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	return c.value
}
