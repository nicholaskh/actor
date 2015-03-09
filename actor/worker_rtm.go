package actor

import (
	"github.com/nicholaskh/actor/config"
	"github.com/nicholaskh/golib/idgen"
	log "github.com/nicholaskh/log4go"
	"github.com/nicholaskh/metrics"
	"math/rand"
	"strconv"
	"sync"
	"time"
)

type WorkerRtm struct {
	config  *config.ConfigWorkerRtm
	backlog chan *Push
	mutex   sync.Mutex
	latency metrics.Histogram

	rtm *rtmPool // pool capacity=max procs
}

func NewRtmWorker(config *config.ConfigWorkerRtm) *WorkerRtm {
	this := new(WorkerRtm)
	this.config = config
	this.backlog = make(chan *Push, config.Backlog)
	this.latency = metrics.NewHistogram(
		metrics.NewExpDecaySample(1028, 0.015))
	if this.Enabled() {
		metrics.Register("latency.rtm", this.latency)
	}
	this.rtm = newRtmPool(config)
	return this
}

func (this *WorkerRtm) Enabled() bool {
	return this.config.MaxProcs > 0 && len(this.config.PrimaryHosts) > 0
}

func (this *WorkerRtm) Start() {
	if !this.Enabled() {
		return
	}

	var wg sync.WaitGroup
	t1 := time.Now()
	for i := 0; i < this.config.MaxProcs; i++ {
		wg.Add(1)

		go func() {
			// warm up
			conn, err := this.rtm.Get()
			if conn != nil {
				if err != nil {
					conn.Close()
				}
				conn.Recycle()
			}

			wg.Done()

			for {
				select {
				case push := <-this.backlog:
					this.doPublish(push)
				}
			}
		}()
	}

	wg.Wait()
	log.Info("worker[rtm] ready: %s", time.Since(t1))
}

func (this *WorkerRtm) Wake(w Wakeable) {
	push := w.(*Push)
	select {
	case this.backlog <- push:
	default:
		log.Warn("rtm backlog full, discarded: %s", string(push.Body))
	}
}

func (this *WorkerRtm) doPublish(push *Push) {
	msg, fromId, toIds := push.Unmarshal()
	for _, toIdStr := range toIds {
		go func(toIdStr string) {
			toId, _ := strconv.ParseInt(toIdStr, 0, 0)
			if this.isToGroupId(toId) {
				this.sendGroup(fromId, toId, msg)
			} else {
				this.sendSingle(fromId, toId, msg)
			}
		}(toIdStr)
	}
}

func (this *WorkerRtm) sendSingle(fromId int64, toId int64, msg string) {
	t1 := time.Now()
	for i := 0; i < this.config.MaxRetries; i++ {
		conn := this.validRtmConn()
		_, err := conn.SendMsg(this.config.ProjectId,
			this.config.SecretKey, 100,
			fromId, toId,
			this.nextMid(), msg)
		if err == nil {
			// sent ok
			this.latency.Update(time.Since(t1).Nanoseconds() / 1e6)
			if i == 0 {
				log.Trace("rtm.single %s: {from^%d to^%d msg^%s}",
					time.Since(t1), fromId, toId, msg)
			} else {
				log.Trace("rtm.single#%d %s: {from^%d to^%d msg^%s}",
					i, time.Since(t1), fromId, toId, msg)
			}

			conn.Recycle() // never forget about this
			return
		}

		// failed to send to rtm
		log.Error("rtm.single: %s {from^%d to^%d msg^%s}",
			err, fromId, toId, msg)
		conn.Close()
		conn.Recycle()
	}

	log.Error("rtm.single quit %s: {from^%d to^%d msg^%s}",
		time.Since(t1), fromId, toId, msg)
}

func (this *WorkerRtm) sendGroup(fromId int64, toId int64, msg string) {
	t1 := time.Now()
	for i := 0; i < this.config.MaxRetries; i++ {
		conn := this.validRtmConn()
		_, err := conn.SendGroupMsg(this.config.ProjectId,
			this.config.SecretKey, 100,
			fromId, toId,
			this.nextMid(), msg)
		if err == nil {
			// sent ok
			if i == 0 {
				log.Trace("rtm.group %s: {from^%d to^%d msg^%s}",
					time.Since(t1), fromId, toId, msg)
			} else {
				log.Trace("rtm.group#%d %s: {from^%d to^%d msg^%s}",
					i, time.Since(t1), fromId, toId, msg)
			}

			conn.Recycle() // never forget about this
			return
		}

		// failed to send to rtm
		log.Error("rtm.group: %s {from^%d to^%d msg^%s}",
			err, fromId, toId, msg)
		conn.Close()
		conn.Recycle()
	}

	log.Error("rtm.group quit %s: {from^%d to^%d msg^%s}",
		time.Since(t1), fromId, toId, msg)
}

func (this *WorkerRtm) validRtmConn() *rtmConn {
	for {
		conn, err := this.rtm.Get()
		if err == nil {
			return conn // this conn maybe already broken, we didn't ping
		}

		log.Warn("rtm fetch: %s", err)
	}
}

func (this *WorkerRtm) nextMid() int64 {
	return rand.Int63()
}

func (this *WorkerRtm) isToGroupId(toId int64) bool {
	_, tag, _, _ := idgen.DecodeId(toId)
	return tag != ticket_user
}
