package tasky

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"sync"
)

type Action uint64

const (
	Cancel Action = iota
	Pause
	Resume
	Restart
)

type Worker interface {
	// Description of the worker and it's usage
	Info() []byte

	// List of available tasks it can service
	Services() []byte

	// Execute the task
	Perform([]byte) ([]byte, bool)

	// Worker status
	Status() []byte

	// Action to be taken on ongoing task
	Signal(Action) bool

	// Worker statistics like number of tasks performed, failure rate,
	// average time per task etc
	Statistics() []byte
}

var (
	wMut    sync.RWMutex
	workers map[string]Worker
)

func init() {
	workers = make(map[string]Worker)
}

func uuid() string {
	b := make([]byte, 16)
	_, err := io.ReadFull(rand.Reader, b)
	if err != nil {
		log.Fatal(err)
	}
	b[6] = (b[6] & 0x0F) | 0x40
	b[8] = (b[8] &^ 0x40) | 0x80
	return fmt.Sprintf("%x%x%x%x%x", b[:4], b[4:6], b[6:8], b[8:10], b[10:])
}

type taskyWorker struct {
	Id string
	w  Worker
}

func (tw *taskyWorker) Info() []byte {
	return tw.w.Info()
}

func (tw *taskyWorker) Services() []byte {
	return tw.w.Services()
}

func (tw *taskyWorker) Perform(job []byte) ([]byte, bool) {
	return tw.w.Perform(job)
}

func (tw *taskyWorker) Status() []byte {
	return tw.w.Status()
}

func (tw *taskyWorker) Signal(act Action) bool {
	return tw.w.Signal(act)
}

func (tw *taskyWorker) Statistics() []byte {
	return tw.w.Statistics()
}

func NewWorker(w Worker) (Worker, error) {
	tw := &taskyWorker{}

	tw.Id = uuid()

	tw.w = w

	wMut.Lock()
	workers[tw.Id] = tw
	wMut.Unlock()

	return tw, nil
}

type worker struct {
	Id   string
	Info []byte
}

type ws struct {
	Workers []worker
}

func listWorkers() ([]byte, error) {
	w := ws{}

	wMut.RLock()
	for k, v := range workers {
		t := worker{}
		t.Id = k
		t.Info = v.Info()

		if len(w.Workers) <= 0 {
			w.Workers = make([]worker, 1)
			w.Workers[0] = t
		} else {
			w.Workers = append(w.Workers, t)
		}
	}
	wMut.RUnlock()

	jsonStr, err := json.Marshal(w)
	if err != nil {
		return nil, err
	}

	return jsonStr, nil
}
