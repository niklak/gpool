package main

import (
	"log"
	"math/rand"
	"sync"
	"time"

	"github.com/niklak/gpool"
)

type dummyLazyHandler struct {
	/*imagine some heavy, complex fields*/
}

func (s *dummyLazyHandler) Handle() (num int, err error) {
	log.Printf("[INFO] dummyLazyHandler.Handle: Doing the job. Do no hurry.\n")
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	num = r.Intn(100)
	time.Sleep(3 * time.Second)
	return
}

func (s *dummyLazyHandler) Close() {
	log.Printf("[WARNING] dummyLazyHandler.Close: Closing the handler")
}

type dummyLazyHandlerManager struct {
	pool     *gpool.Pool[*dummyLazyHandler]
	poolSize int
}

func (s *dummyLazyHandlerManager) Handle() (num int, err error) {
	handler, err := s.pool.Acquire()
	if err != nil {
		return

	}
	defer s.pool.Release(handler)
	return handler.Handle()
}

func (s *dummyLazyHandlerManager) Close() (err error) {

	for i := 0; i < s.poolSize; i++ {
		handler, err := s.pool.Acquire()
		if err != nil {
			return err
		}
		handler.Close()
	}

	s.pool.Close()
	return
}

func NewDummyLazyHandlerManager(size int) *dummyLazyHandlerManager {
	return &dummyLazyHandlerManager{
		pool: gpool.New(size, func() *dummyLazyHandler {
			return &dummyLazyHandler{}
		}),
		poolSize: size,
	}
}

func main() {

	totalTasks := 20
	concurrency := 10
	poolSize := 5

	wg := sync.WaitGroup{}

	manager := NewDummyLazyHandlerManager(poolSize)

	tasks := make(chan int)

	for j := 0; j < concurrency; j++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for _ = range tasks {

				if num, err := manager.Handle(); err != nil {
					log.Printf("Error: %s\n", err)
				} else {
					log.Printf("Got result: %d\n", num)
				}

			}
		}()
	}

	for i := 0; i < totalTasks; i++ {
		tasks <- i
	}
	close(tasks)

	//ensure that manager is not used anymore
	wg.Wait()

	manager.Close()

}
