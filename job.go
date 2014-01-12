package rtot

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
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
	filename     string
	exit         error
}

func newJob(script string) (*job, error) {
	var err error

	f, err := ioutil.TempFile("", "rtot-job-")
	if err != nil {
		return nil, err
	}

	defer func() {
		if f != nil {
			if err != nil {
				os.Remove(f.Name())
			}
		}
	}()

	if !strings.HasPrefix(script, "#!") {
		script = "#!/bin/bash\n" + script
	}

	_, err = f.WriteString(script)
	if err != nil {
		return nil, err
	}

	err = f.Chmod(0755)
	if err != nil {
		return nil, err
	}

	filename := f.Name()
	f.Close()

	var (
		outbuf bytes.Buffer
		errbuf bytes.Buffer
	)

	cmd := exec.Command(filename)
	cmd.Stdout = &outbuf
	cmd.Stderr = &errbuf

	return &job{
		cmd:        cmd,
		state:      "new",
		outBuf:     &outbuf,
		errBuf:     &errbuf,
		createTime: time.Now().UTC(),
		filename:   filename,
	}, nil
}

func (j *job) Run() {
	j.state = "running"
	j.startTime = time.Now().UTC()
	j.exit = j.cmd.Run()
	j.state = "complete"
	j.completeTime = time.Now().UTC()
}

func (j *job) Cleanup() {
	j.cmd.Process.Release()
	os.Remove(j.filename)
}

func (j *job) toJSON(fields *map[string]int) *jobJSON {
	fieldsMap := *fields

	exitString := ""
	outStr := ""
	errStr := ""
	startString := ""
	completeString := ""
	createString := ""
	filenameString := ""

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

	if _, ok := fieldsMap["filename"]; ok {
		filenameString = j.filename
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
		Filename: filenameString,
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
	Filename string `json:"filename,omitempty"`
	Href     string `json:"href"`
}
