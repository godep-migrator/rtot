package rtot

type jobGroupStore interface {
	Add(*job) int
	Get(int) *job
	Getall(string) []*job
	Remove(int) bool
}
