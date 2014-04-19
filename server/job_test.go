package server

import (
	"os"
	"testing"
)

func TestNewJob(t *testing.T) {
	j, err := newJob("echo foop")
	if err != nil {
		t.Error(err)
	}

	if j.state != "new" {
		t.Fail()
	}
}

func TestJobRunSetsStateToComplete(t *testing.T) {
	j, err := newJob("echo foop")
	if err != nil {
		t.Error(err)
	}

	j.Run()
	if j.state != "complete" {
		t.Fail()
	}
}

func TestJobCleanupRemovesJobFile(t *testing.T) {
	j, err := newJob("echo boop")
	if err != nil {
		t.Error(err)
	}

	err = j.Cleanup()
	if err != nil {
		t.Error(err)
	}

	if _, err := os.Stat(j.filename); err == nil {
		t.Fail()
	}
}
