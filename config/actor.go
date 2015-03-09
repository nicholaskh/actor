package config

import (
	"github.com/nicholaskh/golib/ip"
	conf "github.com/nicholaskh/jsconf"
	log "github.com/nicholaskh/log4go"
	"strings"
	"time"
)

type ConfigActor struct {
	EtcdServers       []string
	EtcdSelfAddr      string // calculated based upon HttpApiListenAddr, not configured
	HttpApiListenAddr string
	StatsListenAddr   string
	ProfListenAddr    string
	MetricsLogfile    string
	SchedulerBacklog  int
	FaePoolSize       int

	ConsoleStatsInterval time.Duration

	Poller ConfigPoller
	Worker ConfigWorker
}

func (this *ConfigActor) LoadConfig(cf *conf.Conf) (err error) {
	this.EtcdServers = cf.StringList("etcd_servers", nil)
	this.HttpApiListenAddr = cf.String("http_api_listen_addr", ":9898")
	if len(this.EtcdServers) > 0 {
		this.EtcdSelfAddr = this.HttpApiListenAddr
		if strings.HasPrefix(this.EtcdSelfAddr, ":") {
			// automatically get local ip addr
			myIp := ip.LocalIpv4Addrs()[0]
			this.EtcdSelfAddr = myIp + this.EtcdSelfAddr
		}
	}
	this.StatsListenAddr = cf.String("stats_listen_addr", "127.0.0.1:9010")
	this.ProfListenAddr = cf.String("prof_listen_addr", "")
	this.MetricsLogfile = cf.String("metrics_logfile", "")
	this.SchedulerBacklog = cf.Int("sched_backlog", 10000)
	this.FaePoolSize = cf.Int("fae_pool_size", 200)

	this.ConsoleStatsInterval = cf.Duration("stats_interval", time.Minute*10)

	var section *conf.Conf
	section, err = cf.Section("poller")
	if err != nil {
		return
	}
	this.Poller.loadConfig(section)

	section, err = cf.Section("worker")
	if err != nil {
		return
	}
	this.Worker.loadConfig(section)

	log.Debug("actor config %+v", *this)
	return
}
