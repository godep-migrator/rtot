package rtot

import (
	"sync"
)

type memoryJobGroupStore struct {
	sync.Mutex
	group map[int]*job
}

func newMemoryJobGroupStore() *memoryJobGroupStore {
	return &memoryJobGroupStore{
		group: map[int]*job{},
	}
}

func (m *memoryJobGroupStore) Add(j *job) int {
	m.Lock()
	defer m.Unlock()

	m.group[j.id] = j
	return j.id
}

func (m *memoryJobGroupStore) Get(i int) *job {
	m.Lock()
	defer m.Unlock()

	j, ok := m.group[i]
	if !ok {
		return nil
	}

	return j
}

func (m *memoryJobGroupStore) Getall() []*job {
	m.Lock()
	defer m.Unlock()

	ret := []*job{}
	for _, job := range m.group {
		ret = append(ret, job)
	}

	return ret
}

func (m *memoryJobGroupStore) Remove(i int) bool {
	m.Lock()
	defer m.Unlock()

	_, ok := m.group[i]
	if !ok {
		return false
	}

	delete(m.group, i)
	return true
}
