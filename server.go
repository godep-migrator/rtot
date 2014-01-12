package rtot

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"runtime"
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
	noSuchJob        = &map[string]string{"error": "no such job"}
	defaultJobFields = "out,err,create,start,complete,filename"
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
	app.Get("/all/:state", allJobs)
	app.Get("/:i", getJob)
	app.Delete("/all", delAllJobs)
	app.Delete("/all/:state", delAllJobs)
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

	fields := fieldsMapFromRequest(req)

	j := jobs.Get(i)
	if j == nil {
		r.JSON(404, newJobResponse([]*job{}, fields))
		return
	}

	if j.state != "complete" {
		r.JSON(202, newJobResponse([]*job{j}, fields))
		return
	}

	r.JSON(200, newJobResponse([]*job{j}, fields))
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

	j, err := newJob(string(bodyBytes))
	if err != nil {
		send500(r, err)
		return
	}

	jobs.Add(j)
	go func() {
		j.Run()
		runtime.Goexit()
	}()

	r.JSON(201, newJobResponse([]*job{j}, fieldsMapFromRequest(req)))
}

func delAllJobs(r render.Render, params martini.Params) {
	jobs, ok := getMainJobGroupOr500(r)
	if !ok {
		return
	}

	p := *defaultedParams(params)

	for _, job := range jobs.Getall(p["state"]) {
		jobs.Kill(job.id)
		jobs.Remove(job.id)
	}
	r.JSON(204, "")
}

func allJobs(r render.Render, params martini.Params, req *http.Request) {
	jobs, ok := getMainJobGroupOr500(r)
	if !ok {
		return
	}

	p := *defaultedParams(params)
	r.JSON(200, newJobResponse(jobs.Getall(p["state"]),
		fieldsMapFromRequest(req)))
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

func fieldsMapFromRequest(req *http.Request) *map[string]int {
	fieldsSlice, ok := req.URL.Query()["fields"]
	if !ok {
		fieldsSlice = []string{defaultJobFields}
	}
	return fieldsMapFromString(fieldsSlice[0])
}

func fieldsMapFromString(f string) *map[string]int {
	fields := &map[string]int{}
	for _, part := range strings.Split(f, ",") {
		(*fields)[part] = 1
	}
	return fields
}

func defaultedParams(params martini.Params) *map[string]string {
	state, ok := params["state"]
	if !ok {
		state = ""
	}

	return &map[string]string{
		"state": state,
	}
}
