package actor

import (
	"github.com/funkygao/golib/gofmt"
	"github.com/funkygao/golib/server"
	log "github.com/funkygao/log4go"
	"github.com/funkygao/metrics"
	"github.com/gorilla/mux"
	"io"
	logger "log"
	"net/http"
	"os"
	"runtime"
	"syscall"
	"time"
)

type StatsRunner struct {
	actor     *Actor
	scheduler *Scheduler
	quit      chan bool
}

func NewStatsRunner(actor *Actor, scheduler *Scheduler) *StatsRunner {
	this := new(StatsRunner)
	this.actor = actor
	this.scheduler = scheduler
	this.quit = make(chan bool)
	return this
}

func (this *StatsRunner) Stop() {
	close(this.quit)
}

func (this *StatsRunner) Run() {
	this.launchHttpServ()
	defer this.stopHttpServ()

	var (
		metricsWriter io.Writer
		err           error
	)
	if this.actor.config.MetricsLogfile == "" ||
		this.actor.config.MetricsLogfile == "stdout" {
		metricsWriter = os.Stdout
	} else {
		metricsWriter, err = os.OpenFile(this.actor.config.MetricsLogfile,
			os.O_APPEND|os.O_CREATE|os.O_RDWR, 0644)
		if err != nil {
			panic(err)
		}
	}
	go metrics.Log(metrics.DefaultRegistry, this.actor.config.ConsoleStatsInterval,
		logger.New(metricsWriter, "", logger.LstdFlags))

	ticker := time.NewTicker(this.actor.config.ConsoleStatsInterval)
	defer ticker.Stop()

	var (
		ms           = new(runtime.MemStats)
		rusage       = &syscall.Rusage{}
		lastUserTime int64
		lastSysTime  int64
		userTime     int64
		sysTime      int64
		userCpuUtil  float64
		sysCpuUtil   float64
		nsInMs       uint64 = 1000 * 1000
	)

	for {
		select {
		case <-ticker.C:
			runtime.ReadMemStats(ms)

			syscall.Getrusage(syscall.RUSAGE_SELF, rusage)
			syscall.Getrusage(syscall.RUSAGE_SELF, rusage)
			userTime = rusage.Utime.Sec*1000000000 + int64(rusage.Utime.Usec)
			sysTime = rusage.Stime.Sec*1000000000 + int64(rusage.Stime.Usec)
			userCpuUtil = float64(userTime-lastUserTime) * 100 / float64(this.actor.config.ConsoleStatsInterval)
			sysCpuUtil = float64(sysTime-lastSysTime) * 100 / float64(this.actor.config.ConsoleStatsInterval)

			lastUserTime = userTime
			lastSysTime = sysTime

			log.Info("ver:%s, elapsed:%s, backlog:%d, go:%d, gc:%dms/%d=%d, heap:{%s, %s, %s, %s} cpu:{%3.2f%%us, %3.2f%%sy}",
				server.BuildID,
				time.Since(this.actor.server.StartedAt),
				this.scheduler.Outstandings(),
				runtime.NumGoroutine(),
				ms.PauseTotalNs/nsInMs,
				ms.NumGC,
				ms.PauseTotalNs/(nsInMs*uint64(ms.NumGC))+1,
				gofmt.ByteSize(ms.HeapSys),      // bytes it has asked the operating system for
				gofmt.ByteSize(ms.HeapAlloc),    // bytes currently allocated in the heap
				gofmt.ByteSize(ms.HeapIdle),     // bytes in the heap that are unused
				gofmt.ByteSize(ms.HeapReleased), // bytes returned to the operating system, 5m for scavenger
				userCpuUtil,
				sysCpuUtil)

			log.Info("scheduler: %+v", this.scheduler.Stat())

		case <-this.quit:
			break
		}
	}

}

func (this *StatsRunner) launchHttpServ() {
	if this.actor.config.StatsListenAddr == "" {
		return
	}

	server.LaunchHttpServ(this.actor.config.StatsListenAddr, this.actor.config.ProfListenAddr)
	server.RegisterHttpApi("/s/{cmd}",
		func(w http.ResponseWriter, req *http.Request,
			params map[string]interface{}) (interface{}, error) {
			return this.handleHttpQuery(w, req, params)
		}).Methods("GET")
}

func (this *StatsRunner) stopHttpServ() {
	log.Info("stats httpd stopped")
	server.StopHttpServ()
}

func (this *StatsRunner) handleHttpQuery(w http.ResponseWriter, req *http.Request,
	params map[string]interface{}) (interface{}, error) {
	var (
		vars   = mux.Vars(req)
		cmd    = vars["cmd"]
		output = make(map[string]interface{})
	)

	switch cmd {
	case "ping":
		output["status"] = "ok"

	case "trace":
		stack := make([]byte, 1<<20)
		stackSize := runtime.Stack(stack, true)
		output["callstack"] = string(stack[:stackSize])

	case "sys":
		output["goroutines"] = runtime.NumGoroutine()

		memStats := new(runtime.MemStats)
		runtime.ReadMemStats(memStats)
		output["memory"] = *memStats

		rusage := syscall.Rusage{}
		syscall.Getrusage(0, &rusage)
		output["rusage"] = rusage

	case "stat":
		output["ver"] = server.VERSION
		output["build"] = server.BuildID
		output["stats"] = this.scheduler.Stat()
		output["conf"] = *this.actor.config

	default:
		return nil, server.ErrHttp404
	}

	if this.actor.config.ProfListenAddr != "" {
		output["pprof"] = "http://" +
			this.actor.config.ProfListenAddr + "/debug/pprof/"
	}

	return output, nil
}
