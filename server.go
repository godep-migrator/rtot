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
	"time"

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
	app.Post("/", createJob)
	app.Get("/all", allJobs)
	app.Get("/:i", getJob)
	app.Delete("/all", delAllJobs)
	app.Delete("/:i", delJob)

	logger.Printf("Serving at %s\n", addr)
	http.Handle("/", app)
	http.ListenAndServe(addr, nil)
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
		sendInvalidJob400(r, params["i"])
		return
	}

	jobs, ok := getMainJobGroupOr500(r)
	if !ok {
		return
	}

	if !jobs.Remove(i) {
		r.JSON(404, noSuchJob)
		return
	}

	r.JSON(204, "")
}

func getJob(r render.Render, req *http.Request, params martini.Params) {
	i, err := strconv.Atoi(params["i"])
	if err != nil {
		sendInvalidJob400(r, params["i"])
		return
	}

	jobs, ok := getMainJobGroupOr500(r)
	if !ok {
		return
	}

	j := jobs.Get(i)
	if j == nil {
		r.JSON(404, &jobResponse{Jobs: []*job{}})
		return
	}

	if j.state != "complete" {
		r.JSON(202, &jobResponse{Jobs: []*job{j}})
		return
	}

	r.JSON(200, &jobResponse{Jobs: []*job{j}})
}

func createJob(r render.Render, req *http.Request) {
	bodyBytes, err := ioutil.ReadAll(req.Body)
	if err != nil {
		send500(r, err)
		return
	}

	jobs, ok := getMainJobGroupOr500(r)
	if !ok {
		return
	}

	body := string(bodyBytes)

	f, err := ioutil.TempFile("", "rtot")
	if err != nil {
		send500(r, err)
		return
	}

	if !strings.HasPrefix(body, "#!") {
		body = "#!/bin/bash\n" + body
	}

	_, err = f.WriteString(body)
	if err != nil {
		send500(r, err)
		return
	}

	err = f.Chmod(0755)

	var (
		outbuf bytes.Buffer
		errbuf bytes.Buffer
	)

	fname := f.Name()
	f.Close()

	cmd := exec.Command(fname)
	cmd.Stdout = &outbuf
	cmd.Stderr = &errbuf

	j := &job{
		cmd:        cmd,
		state:      "new",
		outBuf:     &outbuf,
		errBuf:     &errbuf,
		createTime: time.Now().UTC(),
	}
	jobs.Add(j)

	go func(j *job, fname string) {
		j.state = "running"
		j.startTime = time.Now().UTC()
		j.exit = cmd.Run()
		j.state = "complete"
		j.completeTime = time.Now().UTC()

		os.Remove(fname)
	}(j, fname)

	r.JSON(201, &jobResponse{Jobs: []*job{j}})
}

func delAllJobs(r render.Render) {
	jobs, ok := getMainJobGroupOr500(r)
	if !ok {
		return
	}

	for _, job := range jobs.Getall() {
		jobs.Remove(job.id)
	}
	r.JSON(204, "")
}

func allJobs(r render.Render) {
	jobs, ok := getMainJobGroupOr500(r)
	if !ok {
		return
	}

	r.JSON(200, &jobResponse{Jobs: jobs.Getall()})
}

func send500(r render.Render, err error) {
	r.JSON(500, map[string]string{
		"error": fmt.Sprintf("nope. %v", err),
	})
	return
}

func sendInvalidJob400(r render.Render, i string) {
	r.JSON(400, map[string]string{
		"error":   "invalid job number",
		"message": fmt.Sprintf("what is %q?", i),
	})
	return
}

func getMainJobGroupOr500(r render.Render) (*jobGroup, bool) {
	jobs := GetJobGroup("main")
	if jobs == nil {
		r.JSON(500, map[string]string{
			"error": "missing main job group",
		})
		return nil, false
	}
	return jobs, true
}
