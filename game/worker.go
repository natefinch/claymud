package game

import (
	"sync"
	"time"
)

const tickLen = 100 * time.Millisecond

// SpawnWorker creates a long-lived goroutine that handles work until the
// shutdown channel is closed.
// Zone-local workers are spawned with a readlocker
// so they can run in parallel.  A single global worker is spawned with a write
// locker to ensure that none of the other workers are processing events when it
// is (to avoid data races).  Closing the shutdown channel will stop the worker
// as soon as possible.  Waiting on the waitgroup will unblock when all workers
// have exited.
func SpawnWorker(runLock sync.Locker, shutdown <-chan struct{}, wg *sync.WaitGroup) *Worker {
	w := &Worker{
		runLock:   runLock,
		shutdown:  shutdown,
		wg:        wg,
		events:    make(chan func()),
		next:      time.Now().Add(tickLen),
		eventGate: &sync.RWMutex{},
	}
	wg.Add(1)
	go w.run()
	return w
}

// Worker is a single-threaded event loop that serializes actions in the world.
// All actions in a zone are handled by the same worker, eliminating race
// conditions and obviating the need for most locks.  Coordination of actions
// that take place across zones is done by a single "global" worker.  Since few
// actions have that trait, this is less of a performance burden.
type Worker struct {
	shutdown  <-chan struct{}
	wg        *sync.WaitGroup
	next      time.Time     // Next tick
	runLock   sync.Locker   // exclusive lock between zone and global workers
	eventGate *sync.RWMutex // exclusive lock between worker and event sources
	events    chan func()
}

// Handle takes an event from somewhere in the world and executes it.  This
// method is thread safe, so users on their own threads can call it without
// worry.
func (w *Worker) Handle(event func()) {
	// This is a gate that ensures each event source can only put one event on
	// the worker's queue.
	//
	// The worker write locks when it starts to process so that no new items can
	// be put on the goroutine until it is finished processing. This prevents a
	// fast thread from having its event handled and then getting a second one
	// put back on the queue before the worker is finished. We use a reader lock
	// for other threads so they won't contend. We unlock immediately, otherwise
	// we'd have a deadlock between this readlock being locked, waiting on the
	// channel, and the worker trying to write lock before it drains the
	// channel.
	w.eventGate.RLock()
	w.eventGate.RUnlock()
	w.events <- event
}

// run is the goroutine for the worker.
func (w *Worker) run() {
	defer w.wg.Done()

	for {
		if w.closed() {
			return
		}
		func() {
			defer w.runLock.Unlock()
			w.runLock.Lock()
			defer w.eventGate.Unlock()
			w.eventGate.Lock()
			for {
				select {
				case e := <-w.events:
					e()
				default:
					return
				}
			}
		}()
		if w.closed() {
			return
		}
		time.Sleep(time.Until(w.next))
		if w.closed() {
			return
		}
		// we recalculate next based on when we *should* have woken up, since
		// the actual wake up time may vary slightly.
		w.next = w.next.Add(tickLen)
	}
}

// closed returns true if we should exit.
func (w *Worker) closed() bool {
	select {
	case <-w.shutdown:
		return true
	default:
		return false
	}
}
