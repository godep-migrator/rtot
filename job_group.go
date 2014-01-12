package rtot

import (
	"sync"
)

var (
	jobs = newJobGroup()
)

type jobGroup struct {
	sync.Mutex
	cur   int
	group map[int]*job
}

func newJobGroup() *jobGroup {
	return &jobGroup{
		group: map[int]*job{},
		cur:   0,
	}
}

func (g *jobGroup) Add(j *job) int {
	g.Lock()
	defer g.Unlock()

	i := g.cur
	j.id = i
	g.group[g.cur] = j
	g.cur += 1
	return i
}

func (g *jobGroup) Get(i int) *job {
	g.Lock()
	defer g.Unlock()

	j, ok := g.group[i]
	if !ok {
		return nil
	}

	return j
}

func (g *jobGroup) Getall() []*job {
	g.Lock()
	defer g.Unlock()

	ret := []*job{}
	for _, job := range g.group {
		ret = append(ret, job)
	}

	return ret
}

func (g *jobGroup) Remove(i int) bool {
	g.Lock()
	defer g.Unlock()

	_, ok := g.group[i]
	if !ok {
		return false
	}

	delete(g.group, i)
	return true
}

func (g *jobGroup) MarshalJSON() ([]byte, error) {
	return []byte("{}"), nil
}
