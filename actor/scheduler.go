package actor

import (
	"github.com/funkygao/actor/config"
	log "github.com/funkygao/log4go"
	"time"
)

type Scheduler struct {
	config *config.ConfigActor

	stopCh chan bool

	jobN, marchN, pnbN, rtmN int64

	// Poller -> WakeableChannel -> Worker
	backlog chan Wakeable

	mysqlPollers     map[string]Poller // key is db pool name
	beanstalkPollers map[string]Poller // key is tube

	phpWorker Worker
	pnbWorker Worker
	rtmWorker Worker
}

func NewScheduler(cf *config.ConfigActor) *Scheduler {
	this := new(Scheduler)
	this.config = cf
	this.stopCh = make(chan bool)
	this.backlog = make(chan Wakeable, cf.SchedulerBacklog)

	this.mysqlPollers = make(map[string]Poller)
	this.beanstalkPollers = make(map[string]Poller)

	this.phpWorker = NewPhpWorker(&cf.Worker.Php)
	this.pnbWorker = NewPnbWorker(&cf.Worker.Pnb)
	this.rtmWorker = NewRtmWorker(&cf.Worker.Rtm)

	var err error
	for pool, my := range this.config.Poller.Mysql.Servers {
		this.mysqlPollers[pool], err = NewMysqlPoller(my,
			&this.config.Poller.Mysql)
		if err != nil {
			log.Error("mysql poller[%s]: %s", pool, err)
			continue
		}

		log.Info("Started poller[mysql]: %s", pool)

		go this.mysqlPollers[pool].Poll(this.backlog)
	}

	for tube, beanstalk := range this.config.Poller.Beanstalk.Servers {
		this.beanstalkPollers[tube], err = NewBeanstalkdPoller(beanstalk.ServerAddr, tube)
		if err != nil {
			log.Error("beanstalk poller[%s]: %s", tube, err)
			continue
		}

		log.Info("Started poller[beanstalk]: %s", tube)

		go this.beanstalkPollers[tube].Poll(this.backlog)
	}

	return this
}

func (this *Scheduler) Outstandings() int {
	return len(this.backlog)
}

func (this *Scheduler) Stat() map[string]interface{} {
	return map[string]interface{}{
		"backlog": this.Outstandings(),
		"job":     this.jobN,
		"march":   this.marchN,
		"pnb":     this.pnbN,
		"rtm":     this.rtmN,
	}
}

func (this *Scheduler) Stop() {
	for _, p := range this.beanstalkPollers {
		p.Stop()
	}
	for _, p := range this.mysqlPollers {
		if p != nil {
			p.Stop()
		}
	}

	close(this.stopCh)
}

func (this *Scheduler) Run() {
	const (
		typePubnub = "p"
		typeRtm    = "r"
	)

	this.phpWorker.Start()
	this.pnbWorker.Start()
	this.rtmWorker.Start()

	log.Info("scheduler ready")

	for {
		select {
		case w, open := <-this.backlog:
			if !open {
				log.Critical("Scheduler chan closed: aborted")
				return
			}

			if w.Ignored() {
				log.Debug("ignored: %+v", w)
				continue
			}

			elapsed := time.Since(w.DueTime())
			if elapsed.Seconds() > this.config.Poller.Mysql.Interval.Seconds()+1 {
				log.Debug("late %s for %+v", elapsed, w)
			}

			switch w := w.(type) {
			case *Push:
				switch w.Type() {
				case typePubnub:
					this.pnbN++
					go this.pnbWorker.Wake(w)

				case typeRtm:
					this.rtmN++
					go this.rtmWorker.Wake(w)

				default:
					// should never happen
					log.Error("Unknown push type[%s]: %s", w.Type(), string(w.Body))
				}

			case *Job:
				this.jobN++
				go this.phpWorker.Wake(w)

			case *March:
				this.marchN++
				go this.phpWorker.Wake(w)
			}

		case <-this.stopCh:
			log.Info("Scheduler stopped")
			return

		}
	}
}
