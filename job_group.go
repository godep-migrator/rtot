package rtot

import (
	"fmt"
	"sync"
)

var (
	jobGroups      = map[string]*jobGroup{}
	jobGroupsMutex sync.Mutex
	errNoSuchJob   = fmt.Errorf("no such job")
)

type jobGroup struct {
	cur   int
	store jobGroupStore
}

// GetJobGroup is how you get a job group, assuming it exists
func GetJobGroup(name string) *jobGroup {
	jobGroupsMutex.Lock()
	defer jobGroupsMutex.Unlock()

	g, ok := jobGroups[name]
	if !ok {
		return nil
	}

	return g
}

// NewJobGroup is used to initialize members of the jobGroups var
func NewJobGroup(name, storeType string) (*jobGroup, error) {
	var store jobGroupStore
	switch storeType {
	case "memory":
		store = newMemoryJobGroupStore()
	default:
		return nil, fmt.Errorf("invalid storeType %v", storeType)
	}
	jobGroupsMutex.Lock()
	defer jobGroupsMutex.Unlock()
	jobGroups[name] = &jobGroup{
		store: store,
		cur:   0,
	}
	return jobGroups[name], nil
}

func (g *jobGroup) Add(j *job) int {
	i := g.cur
	j.id = i
	g.store.Add(j)
	g.cur += 1
	return i
}

func (g *jobGroup) Get(i int) *job {
	return g.store.Get(i)
}

func (g *jobGroup) Kill(i int) error {
	job := g.store.Get(i)
	if job != nil {
		return job.cmd.Process.Kill()
	}
	return errNoSuchJob
}

func (g *jobGroup) Getall(state string) []*job {
	return g.store.Getall(state)
}

func (g *jobGroup) Remove(i int) bool {
	job := g.store.Get(i)
	if job != nil {
		job.Cleanup()
	}
	return g.store.Remove(i)
}
