package rtot

import (
	"bytes"
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

func (j *job) toJSON(fields *map[string]int) *jobJSON {
	fieldsMap := *fields

	exitString := ""
	outStr := ""
	errStr := ""
	startString := ""
	completeString := ""
	createString := ""

	if j.exit != nil {
		exitString = j.exit.Error()
	}

	if _, ok := fieldsMap["out"]; ok {
		outStr = string(j.outBuf.Bytes())
	}

	if _, ok := fieldsMap["err"]; ok {
		errStr = string(j.errBuf.Bytes())
	}

	if _, ok := fieldsMap["create"]; ok {
		createString = j.createTime.String()
	}

	if _, ok := fieldsMap["start"]; ok {
		startString = j.startTime.String()
	}

	if _, ok := fieldsMap["complete"]; ok {
		completeString = j.completeTime.String()
	}

	return &jobJSON{
		ID:       j.id,
		Out:      outStr,
		Err:      errStr,
		State:    j.state,
		Exit:     exitString,
		Start:    startString,
		Complete: completeString,
		Create:   createString,
		Href:     fmt.Sprintf("/%v", j.id),
	}
}

type jobJSON struct {
	ID       int    `json:"id"`
	Out      string `json:"out,omitempty"`
	Err      string `json:"err,omitempty"`
	State    string `json:"state"`
	Exit     string `json:"exit,omitempty"`
	Start    string `json:"start,omitempty"`
	Complete string `json:"complete,omitempty"`
	Create   string `json:"create,omitempty"`
	Href     string `json:"href"`
}
