package rtot

import (
	"bytes"
	"encoding/json"
	"os/exec"
	"sync"
)

var (
	jobs   = map[int]*job{}
	jobNum = 1

	jobsMutex   sync.Mutex
	jobNumMutex sync.Mutex
)

type job struct {
	i      int
	outBuf *bytes.Buffer
	errBuf *bytes.Buffer
	cmd    *exec.Cmd
	state  string
	exit   error
}

func (j *job) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Out   string `json:"out"`
		Err   string `json:"err"`
		State string `json:"state"`
		Exit  error  `json:"exit"`
	}{
		Out:   string(j.outBuf.Bytes()),
		Err:   string(j.errBuf.Bytes()),
		State: j.state,
		Exit:  j.exit,
	})
}
