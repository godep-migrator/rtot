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
	"time"

	"github.com/codegangsta/martini"
	"github.com/codegangsta/martini-contrib/render"
)

var (
	logger = log.New(os.Stdout, "[rtot] ", log.LstdFlags|log.Lshortfile)

	theBeginning     time.Time
	defaultJobFields = "out,err,create,start,complete,filename"

	notAuthorized = &map[string]string{
		"error":   "not authorized",
		"message": "phooey!",
	}
	rootMap = &map[string]*map[string]string{
		"links": &map[string]string{
			"jobs":      "/jobs{?state}",
			"jobs.byID": "/jobs/{jobs.id}",
			"ping":      "/ping",
		},
	}
	noSuchJob = &map[string]string{"error": "no such job"}
)

func init() {
	theBeginning = time.Now()
}

// ServerMain is the entry point for the server executable
func ServerMain(addr, secret string) {
	app := martini.Classic()
	app.Use(render.Renderer())
	app.Use(func(res http.ResponseWriter, req *http.Request) {
		if req.URL.Path == "/ping" && req.Method == "GET" {
			return
		}

		if req.Header.Get("Rtot-Secret") != secret {
			http.Error(res, "Not Authorized", http.StatusUnauthorized)
		}
	})
	app.Use(func(res http.ResponseWriter) {
		res.Header().Set("Rtot-Version", VersionString)
	})

	app.Get("/", root)
	app.Delete("/", die)

	app.Get("/ping", ping)

	app.Post("/jobs", createJob)
	app.Get("/jobs", allJobs)
	app.Get("/jobs/:id", getJob)
	app.Delete("/jobs", delAllJobs)
	app.Delete("/jobs/:id", delJob)

	logger.Printf("Serving at %s\n", addr)
	http.Handle("/", app)
	http.ListenAndServe(addr, nil)
}

func root(r render.Render) {
	r.JSON(200, rootMap)
}

func ping(r render.Render) {
	r.JSON(200, &map[string]string{
		"message": "still here",
		"uptime":  time.Now().Sub(theBeginning).String(),
	})
}

func die(r render.Render, req *http.Request) {
	go os.Exit(1)
	r.JSON(204, "")
}

func delJob(r render.Render, req *http.Request, params martini.Params) {
	i, err := strconv.Atoi(params["id"])
	if err != nil {
		sendInvalidJob400(r, params["id"])
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

func getJob(r render.Render, res http.ResponseWriter,
	req *http.Request, params martini.Params) {

	i, err := strconv.Atoi(params["id"])
	if err != nil {
		sendInvalidJob400(r, params["id"])
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

	res.Header().Set("Location", j.Href())

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

func delAllJobs(r render.Render, req *http.Request) {
	jobs, ok := getMainJobGroupOr500(r)
	if !ok {
		return
	}

	for _, job := range jobs.Getall(req.URL.Query().Get("state")) {
		jobs.Kill(job.id)
		jobs.Remove(job.id)
	}

	r.JSON(204, "")
}

func allJobs(r render.Render, req *http.Request) {
	jobs, ok := getMainJobGroupOr500(r)
	if !ok {
		return
	}

	r.JSON(200, newJobResponse(jobs.Getall(req.URL.Query().Get("state")),
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
