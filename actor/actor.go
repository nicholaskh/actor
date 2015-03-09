/*
   StatsRunner
     |
     |                             +- beantalkds
     |                             |
   Actor --- Schduler --- Pollers --- mysql farm
               |            |
               |            V backlog(channel of Wakeable)
               |            |
             Worker --------+
               |
               | go Wake(Wakeable) with retries in case of lock conflict
               V
    -----------------------------------
               |        |        |
              php    pubnub     RTM

*/
package actor

import (
	"github.com/funkygao/actor/config"
	"github.com/funkygao/etclib"
	"github.com/funkygao/golib/server"
	log "github.com/funkygao/log4go"
)

type Actor struct {
	server *server.Server
	config *config.ConfigActor

	statsRunner *StatsRunner

	scheduler *Scheduler
}

func New(server *server.Server) (this *Actor) {
	this = new(Actor)
	this.server = server

	this.config = new(config.ConfigActor)
	if err := this.config.LoadConfig(server.Conf); err != nil {
		panic(err)
	}

	this.scheduler = NewScheduler(this.config)
	this.statsRunner = NewStatsRunner(this, this.scheduler)

	return
}

func (this *Actor) ServeForever() {
	go this.scheduler.Run()

	if this.config.EtcdSelfAddr != "" {
		err := etclib.Dial(this.config.EtcdServers)
		if err != nil {
			log.Error("etcd[%+v]: %s", this.config.EtcdServers, err)
		} else {
			etclib.BootService(this.config.EtcdSelfAddr, etclib.SERVICE_ACTOR)
		}
	}

	// after zk connected, start fae proxy
	fae = NewFaeExecutor(this.config.FaePoolSize)
	fae.StartCluster()
	//fae.Warmup()  has bug

	this.statsRunner.Run()
}

func (this *Actor) Stop() {
	if this.config.EtcdSelfAddr != "" {
		if !etclib.IsConnected() {
			if err := etclib.Dial(this.config.EtcdServers); err != nil {
				log.Error("etcd[%+v]: %s", this.config.EtcdServers, err)

				etclib.Close()
			}
		}

		if etclib.IsConnected() {
			log.Info("shutdown actor node[%s] in etcd", this.config.EtcdSelfAddr)

			etclib.ShutdownService(this.config.EtcdSelfAddr, etclib.SERVICE_ACTOR)
		}

	}

	this.statsRunner.stopHttpServ()
	this.statsRunner.Stop()

	this.scheduler.Stop()
}
