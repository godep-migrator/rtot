package rtot

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"

	"github.com/codegangsta/martini"
)

var (
	logger = log.New(os.Stdout, "[rtot] ", log.LstdFlags|log.Lshortfile)
	jobs   = map[int]*job{}
	jobNum = 1

	jobMutex sync.Mutex
)

type job struct {
	i    int
	Out  string `json:"out"`
	Err  string `json:"err"`
	cmd  *exec.Cmd
	done bool
	Exit error `json:"exit"`
}

// ServerMain is the entry point for the server executable
func ServerMain(addr, secret string) {
	app := martini.Classic()

	app.Get("/", func() string {
		return "still here\n"
	})

	app.Delete("/:i", func(req *http.Request, params martini.Params) (int, string) {
		if req.Header.Get("Rtot-Secret") != secret {
			return 401, "phooey!\n"
		}

		i, err := strconv.Atoi(params["i"])
		if err != nil {
			return 400, fmt.Sprintf("what is %q?\n", params["i"])
		}

		_, ok := jobs[i]
		if !ok {
			return 404, "no such job\n"
		}

		delete(jobs, i)
		return 204, "\n"
	})

	app.Get("/:i", func(res http.ResponseWriter, req *http.Request,
		params martini.Params) (int, string) {

		if req.Header.Get("Rtot-Secret") != secret {
			return 401, "phooey!\n"
		}

		i, err := strconv.Atoi(params["i"])
		if err != nil {
			return 400, fmt.Sprintf("what is %q?\n", params["i"])
		}
		j, ok := jobs[i]
		if !ok {
			return 404, "no such job\n"
		}

		if !j.done {
			return 202, "still at it\n"
		}

		retBytes, err := json.Marshal(j)
		if err != nil {
			return 500, fmt.Sprintf("bork: %v\n", err)
		}

		res.Header().Set("Content-Type", "application/json; charset=utf-8")
		return 200, string(retBytes) + "\n"
	})

	app.Post("/", func(req *http.Request) (int, string) {
		if req.Header.Get("Rtot-Secret") != secret {
			return 401, "phooey!\n"
		}

		bodyBytes, err := ioutil.ReadAll(req.Body)
		if err != nil {
			return 400, fmt.Sprintf("nope. %v\n", err)
		}

		body := string(bodyBytes)

		f, err := ioutil.TempFile("", "rtot")
		if err != nil {
			return 500, fmt.Sprintf("nope. %v\n", err)
		}

		if !strings.HasPrefix(body, "#!") {
			body = "#!/bin/bash\n" + body
		}

		_, err = f.WriteString(body)
		if err != nil {
			return 500, fmt.Sprintf("nope. %v\n", err)
		}

		err = f.Chmod(0755)

		jobMutex.Lock()
		defer jobMutex.Unlock()
		i := jobNum
		jobNum += 1

		go func() {
			var outbuf bytes.Buffer
			var errbuf bytes.Buffer

			fname := f.Name()
			f.Close()

			cmd := exec.Command(fname)
			cmd.Stdout = &outbuf
			cmd.Stderr = &errbuf

			j := &job{i: i, cmd: cmd}
			jobs[i] = j

			j.Exit = cmd.Run()
			j.Out = string(outbuf.Bytes())
			j.Err = string(errbuf.Bytes())
			j.done = true

			os.Remove(fname)
		}()

		return 201, fmt.Sprintf("/%v\n", i)
	})

	logger.Printf("Serving at %s\n", addr)
	http.ListenAndServe(addr, app)
}
