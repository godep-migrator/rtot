package rtot

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/codegangsta/martini"
)

var (
	logBuf bytes.Buffer

	testServerContext = &serverContext{
		logger:        log.New(&logBuf, "[rtot-test]", log.LstdFlags),
		theBeginning:  time.Now(),
		notAuthorized: defaultNotAuthorized,
		rootMap:       defaultRootMap,
		noSuchJob:     defaultNoSuchJob,
		secret:        "swordfish",

		fl:   flag.NewFlagSet("rtot-test", flag.ContinueOnError),
		args: []string{},
		env:  []string{},

		noop: true,
	}
)

func setupServer() (*httptest.ResponseRecorder, *martini.ClassicMartini) {
	m := NewServer(testServerContext)
	hr := httptest.NewRecorder()
	m.MapTo(hr, (*http.Handler)(nil))
	return hr, m
}

func getResponse(verb, path, ctype string, body io.Reader, authd bool) *httptest.ResponseRecorder {
	hr, m := setupServer()

	req, err := http.NewRequest(verb, path, body)
	if err != nil {
		panic(err)
	}

	if ctype != "" {
		req.Header.Set("Content-Type", ctype)
	}

	if authd {
		req.Header.Set("Rtot-Secret", testServerContext.secret)
	}

	m.ServeHTTP(hr, req)
	return hr
}

func createTestJob(t *testing.T, script string) string {
	createResp := getResponse("POST", "/jobs", "application/octet-stream",
		strings.NewReader(script), true)
	if createResp.Code != 201 {
		testDumpFail(t, createResp)
	}
	return createResp.Body.String()
}

func testDumpFail(t *testing.T, resp *httptest.ResponseRecorder) {
	fmt.Println(resp.Body.String())
	t.Fail()
}

func TestServerMainDoesNotExplode(t *testing.T) {
	if exit := ServerMain(testServerContext); exit != 0 {
		t.Fail()
	}
}

func TestServerRespondsToPingUnauthorized(t *testing.T) {
	resp := getResponse("GET", "/ping", "", nil, false)
	if resp.Code != 200 {
		testDumpFail(t, resp)
	}
}

func TestServerRejectsUnauthorized(t *testing.T) {
	resp := getResponse("GET", "/", "", nil, false)
	if resp.Code != 401 {
		testDumpFail(t, resp)
	}
}

func TestServerRespondsToRoot(t *testing.T) {
	resp := getResponse("GET", "/", "", nil, true)
	if resp.Code != 200 {
		testDumpFail(t, resp)
	}
}

func TestServerRespondsToDie(t *testing.T) {
	resp := getResponse("DELETE", "/", "", nil, true)
	if resp.Code != 204 {
		testDumpFail(t, resp)
	}
}

func TestServerCreateJob(t *testing.T) {
	resp := getResponse("POST", "/jobs", "application/octet-stream",
		strings.NewReader("echo something already"), true)
	if resp.Code != 201 {
		testDumpFail(t, resp)
	}
}

func TestServerGetAllJobs(t *testing.T) {
	createTestJob(t, "echo another thing")
	resp := getResponse("GET", "/jobs", "", nil, true)
	if resp.Code != 200 {
		testDumpFail(t, resp)
	}
}

func TestServerGetJobByID(t *testing.T) {
	createTestJob(t, "echo canyon")
	resp := getResponse("GET", "/jobs/0", "", nil, true)
	if resp.Code != 202 {
		testDumpFail(t, resp)
	}
}

func TestServerDeleteAllJobs(t *testing.T) {
	createTestJob(t, "echo chamber")
	resp := getResponse("DELETE", "/jobs", "", nil, true)
	if resp.Code != 204 {
		testDumpFail(t, resp)
	}
}

func TestServerDeleteJobByID(t *testing.T) {
	out := createTestJob(t, "echo chamber")
	dest := &jobResponse{}
	err := json.Unmarshal([]byte(out), &dest)
	if err != nil {
		t.Error(err)
	}

	if dest.Jobs == nil {
		t.Fail()
	}

	for _, job := range dest.Jobs {
		resp := getResponse("DELETE", fmt.Sprintf("/jobs/%v", job.ID), "", nil, true)
		if resp.Code != 204 {
			testDumpFail(t, resp)
		}
	}
}
