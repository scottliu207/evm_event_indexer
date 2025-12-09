package background

import (
	"context"
	"sync"
)

type BGManager struct {
	workers []Worker
	wg      *sync.WaitGroup
}

func NewBGManager() *BGManager {
	return &BGManager{
		wg: &sync.WaitGroup{},
	}
}

func (m *BGManager) AddWorker(w Worker) {
	m.workers = append(m.workers, w)
}

func (m *BGManager) Start(ctx context.Context) {
	for _, w := range m.workers {
		m.wg.Add(1)
		go func(w Worker) {
			defer m.wg.Done()
			w.Run(ctx)
		}(w)
	}
}

func (m *BGManager) Stop() {
	m.wg.Wait()
}
