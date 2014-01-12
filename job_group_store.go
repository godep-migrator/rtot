package rtot

type jobGroupStore interface {
	Add(*job) int
	Get(int) *job
	Getall() []*job
	Remove(int) bool
}
