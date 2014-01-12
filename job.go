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

type jobJSON struct {
	ID       int    `json:"id"`
	Out      string `json:"out"`
	Err      string `json:"err"`
	State    string `json:"state"`
	Exit     string `json:"exit"`
	Start    string `json:"start"`
	Complete string `json:"complete"`
	Create   string `json:"create"`
	Href     string `json:"href"`
}

func (j *job) MarshalJSON() ([]byte, error) {
	exitString := ""
	if j.exit == nil {
		exitString = ""
	} else {
		exitString = j.exit.Error()
	}
	return json.Marshal(&jobJSON{
		ID:       j.id,
		Out:      string(j.outBuf.Bytes()),
		Err:      string(j.errBuf.Bytes()),
		State:    j.state,
		Exit:     exitString,
		Start:    j.startTime.String(),
		Complete: j.completeTime.String(),
		Create:   j.createTime.String(),
		Href:     fmt.Sprintf("/%v", j.id),
	})
}
