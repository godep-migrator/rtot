package server

import (
	"crypto/md5"
	"crypto/rand"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/go-martini/martini"
	"github.com/martini-contrib/render"
)

var (
	defaultNotAuthorized = &map[string]string{
		"error":   "not authorized",
		"message": "phooey!",
	}
	defaultRootMap = &map[string]*map[string]string{
		"links": &map[string]string{
			"jobs":       "/jobs{?state}",
			"jobs.by_id": "/jobs/{jobs.id}",
			"ping":       "/ping",
		},
	}
	defaultNoSuchJob     = &map[string]string{"error": "no such job"}
	defaultServerContext = &serverContext{
		logger:           logrus.New(),
		theBeginning:     time.Now(),
		defaultJobFields: "out,err,create,start,complete,filename",

		addr:   os.Getenv("RTOT_ADDR"),
		secret: os.Getenv("RTOT_SECRET"),

		notAuthorized: defaultNotAuthorized,
		rootMap:       defaultRootMap,
		noSuchJob:     defaultNoSuchJob,

		fl:   flag.NewFlagSet("rtot", flag.ExitOnError),
		args: os.Args[1:],
		env:  os.Environ(),
	}
)

type serverContext struct {
	logger           *logrus.Logger
	theBeginning     time.Time
	defaultJobFields string
	addr             string
	secret           string
	notAuthorized    *map[string]string
	rootMap          *map[string]*map[string]string
	noSuchJob        *map[string]string

	fl   *flag.FlagSet
	args []string
	env  []string

	noop bool
}

// ServerMain is the entry point for the server executable
func ServerMain(c *serverContext) int {
	if c == nil {
		c = defaultServerContext
	}

	if c.addr == "" {
		c.addr = ":8457"
	}

	logFmt := os.Getenv("RTOT_LOG_FORMAT")
	if logFmt == "" {
		logFmt = "text"
	}

	c.fl.StringVar(&logFmt,
		"f", logFmt,
		"Log output format (text, json) [RTOT_LOG_FORMAT]")
	c.fl.StringVar(&c.addr,
		"a", c.addr, "HTTP Server address [RTOT_ADDR]")
	c.fl.StringVar(&c.secret,
		"s", c.secret, "Secret string for secret stuff [RTOT_SECRET]")
	versionFlag := c.fl.Bool("v", false, "Show version and exit")

	c.fl.Parse(c.args)

	if logFmt == "json" {
		c.logger.Formatter = &logrus.JSONFormatter{}
	}

	if *versionFlag {
		fmt.Printf("rtot %v\n", VersionString)
		os.Exit(0)
	}

	if c.secret == "" {
		c.secret = makeSecret()
		c.logger.WithField("secret", c.secret).Info("No secret given, so generated one.")
	}

	_, err := NewJobGroup("main", "memory")
	if err != nil {
		c.logger.WithField("err", err).Warn("Failed to init job store")
		os.Exit(1)
	}

	m := NewServer(c)

	c.logger.WithField("addr", c.addr).Info("Serving")
	http.Handle("/", m)
	if !c.noop {
		http.ListenAndServe(c.addr, nil)
	}
	return 0
}

// NewServer creates a martini.ClassicMartini based on server context
func NewServer(c *serverContext) *martini.ClassicMartini {
	r := martini.NewRouter()
	m := martini.New()
	m.Use(func(res http.ResponseWriter, req *http.Request, sc *serverContext, c martini.Context) {
		start := time.Now()
		sc.logger.WithFields(logrus.Fields{
			"method": req.Method,
			"path":   req.URL.Path,
		}).Info("started")

		rw := res.(martini.ResponseWriter)
		c.Next()

		sc.logger.WithFields(logrus.Fields{
			"code":     rw.Status(),
			"status":   http.StatusText(rw.Status()),
			"duration": fmt.Sprintf("%v", time.Since(start)),
		}).Info("completed")
	})
	m.Use(martini.Recovery())
	m.MapTo(r, (*martini.Routes)(nil))
	m.Action(r.Handle)

	cm := &martini.ClassicMartini{m, r}
	cm.Use(render.Renderer())
	cm.Use(func(res http.ResponseWriter, req *http.Request) {
		if req.URL.Path == "/ping" && req.Method == "GET" {
			return
		}

		if req.Header.Get("Authorization") != "rtot "+c.secret {
			http.Error(res, "Not Authorized", http.StatusUnauthorized)
		}
	})
	cm.Use(func(res http.ResponseWriter) {
		res.Header().Set("Rtot-Version", VersionString)
	})
	cm.Map(c)

	cm.Get("/", root)
	cm.Delete("/", die)

	cm.Get("/ping", ping)

	cm.Post("/jobs", createJob)
	cm.Get("/jobs", allJobs)
	cm.Get("/jobs/:id", getJob)
	cm.Delete("/jobs", delAllJobs)
	cm.Delete("/jobs/:id", delJob)

	return cm
}

func root(r render.Render, c *serverContext) {
	r.JSON(200, c.rootMap)
}

func ping(r render.Render, c *serverContext) {
	r.JSON(200, &map[string]string{
		"message": "still here",
		"uptime":  time.Now().Sub(c.theBeginning).String(),
	})
}

func die(r render.Render, req *http.Request, c *serverContext) {
	if !c.noop {
		go os.Exit(1)
	}
	r.JSON(204, "")
}

func delJob(r render.Render, req *http.Request, params martini.Params, c *serverContext) {
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
		r.JSON(404, c.noSuchJob)
		return
	}

	r.JSON(204, "")
}

func getJob(r render.Render, res http.ResponseWriter,
	req *http.Request, params martini.Params, c *serverContext) {

	i, err := strconv.Atoi(params["id"])
	if err != nil {
		sendInvalidJob400(r, params["id"])
		return
	}

	jobs, ok := getMainJobGroupOr500(r)
	if !ok {
		return
	}

	fields := fieldsMapFromRequest(req, c)

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

func createJob(r render.Render, req *http.Request, c *serverContext) {
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
	if !c.noop {
		go func() {
			j.Run()
			runtime.Goexit()
		}()
	}

	r.JSON(201, newJobResponse([]*job{j}, fieldsMapFromRequest(req, c)))
}

func delAllJobs(r render.Render, req *http.Request, c *serverContext) {
	jobs, ok := getMainJobGroupOr500(r)
	if !ok {
		return
	}

	for _, job := range jobs.Getall(req.URL.Query().Get("state")) {
		if !c.noop {
			jobs.Kill(job.id)
		}
		jobs.Remove(job.id)
	}

	r.JSON(204, "")
}

func allJobs(r render.Render, req *http.Request, c *serverContext) {
	jobs, ok := getMainJobGroupOr500(r)
	if !ok {
		return
	}

	r.JSON(200, newJobResponse(jobs.Getall(req.URL.Query().Get("state")),
		fieldsMapFromRequest(req, c)))
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

func fieldsMapFromRequest(req *http.Request, c *serverContext) *map[string]int {
	fieldsSlice, ok := req.URL.Query()["fields"]
	if !ok {
		fieldsSlice = []string{c.defaultJobFields}
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

func makeSecret() string {
	buf := make([]byte, 16)
	_, err := rand.Read(buf)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to generate a secret! %v", err)
		os.Exit(1)
	}
	hash := md5.New()
	io.WriteString(hash, string(buf))
	return fmt.Sprintf("%x", hash.Sum(nil))
}
