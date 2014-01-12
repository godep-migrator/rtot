package rtot

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/codegangsta/martini"
	"github.com/codegangsta/martini-contrib/render"
)

var (
	logger = log.New(os.Stdout, "[rtot] ", log.LstdFlags|log.Lshortfile)

	pong          = &map[string]string{"message": "still here"}
	notAuthorized = &map[string]string{
		"error":   "not authorized",
		"message": "phooey!",
	}
	noSuchJob = &map[string]string{"error": "no such job"}
)

// ServerMain is the entry point for the server executable
func ServerMain(addr, secret string) {
	app := martini.Classic()
	app.Use(render.Renderer())
	app.Use(func(res http.ResponseWriter, req *http.Request) {
		if req.URL.Path == "/" && req.Method == "GET" {
			return
		}

		if req.Header.Get("Rtot-Secret") != secret {
			http.Error(res, "Not Authorized", http.StatusUnauthorized)
		}
	})

	app.Get("/", root)
	app.Delete("/", die)
	app.Delete("/:i", delJob)
	app.Get("/:i", getJob)
	app.Post("/", createJob)

	logger.Printf("Serving at %s\n", addr)
	http.ListenAndServe(addr, app)
}

func root(r render.Render) {
	r.JSON(200, pong)
}

func die(r render.Render, req *http.Request) {
	go os.Exit(1)
	r.JSON(204, "")
}

func delJob(r render.Render, req *http.Request, params martini.Params) {
	i, err := strconv.Atoi(params["i"])
	if err != nil {
		r.JSON(400, map[string]string{
			"error":   "invalid job number",
			"message": fmt.Sprintf("what is %q?", params["i"]),
		})
		return
	}

	jobsMutex.Lock()
	_, ok := jobs[i]
	jobsMutex.Unlock()
	if !ok {
		r.JSON(404, noSuchJob)
		return
	}

	delete(jobs, i)
	r.JSON(204, "")
}

func getJob(r render.Render, req *http.Request, params martini.Params) {
	i, err := strconv.Atoi(params["i"])
	if err != nil {
		r.JSON(400, map[string]string{
			"error":   "invalid job number",
			"message": fmt.Sprintf("what is %q?", params["i"]),
		})
		return
	}

	jobsMutex.Lock()
	j, ok := jobs[i]
	jobsMutex.Unlock()
	if !ok {
		r.JSON(404, noSuchJob)
		return
	}

	if j.state != "complete" {
		r.JSON(202, j)
		return
	}

	r.JSON(200, j)
}

func createJob(r render.Render, req *http.Request) {
	bodyBytes, err := ioutil.ReadAll(req.Body)
	if err != nil {
		r.JSON(400, map[string]string{
			"error": fmt.Sprintf("nope. %v", err),
		})
		return
	}

	body := string(bodyBytes)

	f, err := ioutil.TempFile("", "rtot")
	if err != nil {
		r.JSON(500, map[string]string{
			"error": fmt.Sprintf("nope. %v", err),
		})
		return
	}

	if !strings.HasPrefix(body, "#!") {
		body = "#!/bin/bash\n" + body
	}

	_, err = f.WriteString(body)
	if err != nil {
		r.JSON(500, map[string]string{
			"error": fmt.Sprintf("nope. %v", err),
		})
		return
	}

	err = f.Chmod(0755)

	jobNumMutex.Lock()
	defer jobNumMutex.Unlock()
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

		j := &job{
			i:      i,
			cmd:    cmd,
			state:  "new",
			outBuf: &outbuf,
			errBuf: &errbuf,
		}
		jobsMutex.Lock()
		jobs[i] = j
		jobsMutex.Unlock()

		j.state = "running"
		j.exit = cmd.Run()
		j.state = "complete"

		os.Remove(fname)
	}()

	r.JSON(201, map[string]string{
		"message": "created job",
		"href":    fmt.Sprintf("/%v", i),
	})
}
