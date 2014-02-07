package rtot

import (
	"bytes"
	"flag"
	"log"
	"testing"
	"time"
)

var (
	logBuf bytes.Buffer

	testServerContext = &serverContext{
		logger:        log.New(&logBuf, "[rtot-test]", log.LstdFlags),
		theBeginning:  time.Now(),
		notAuthorized: defaultNotAuthorized,
		rootMap:       defaultRootMap,
		noSuchJob:     defaultNoSuchJob,

		fl:   flag.NewFlagSet("rtot-test", flag.ContinueOnError),
		args: []string{},
		env:  []string{},

		noop: true,
	}
)

func TestServerMainDoesNotExplode(t *testing.T) {
	if exit := ServerMain(testServerContext); exit != 0 {
		t.Fail()
	}
}
