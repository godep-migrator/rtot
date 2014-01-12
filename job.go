package rtot

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
	"time"
)

type job struct {
	id           int
	outBuf       *bytes.Buffer
	errBuf       *bytes.Buffer
	cmd          *exec.Cmd
	state        string
	createTime   time.Time
	startTime    time.Time
	completeTime time.Time
	exit         error
}

func (j *job) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		ID       int    `json:"id"`
		Out      string `json:"out"`
		Err      string `json:"err"`
		State    string `json:"state"`
		Exit     error  `json:"exit"`
		Start    string `json:"start"`
		Complete string `json:"complete"`
		Create   string `json:"create"`
		Href     string `json:"href"`
	}{
		ID:       j.id,
		Out:      string(j.outBuf.Bytes()),
		Err:      string(j.errBuf.Bytes()),
		State:    j.state,
		Exit:     j.exit,
		Start:    j.startTime.String(),
		Complete: j.completeTime.String(),
		Create:   j.createTime.String(),
		Href:     fmt.Sprintf("/%v", j.id),
	})
}
