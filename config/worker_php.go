package config

import (
	conf "github.com/nicholaskh/jsconf"
	"time"
)

type ConfigWorkerPhp struct {
	DryRun           bool
	DebugLocking     bool
	Timeout          time.Duration
	MaxFlightEntries int
	LockExpires      time.Duration

	// if use php as worker, it's callback url template
	Job   string
	March string
	Pve   string
}

func (this *ConfigWorkerPhp) loadConfig(cf *conf.Conf) {
	this.DryRun = cf.Bool("dry_run", true)
	this.DebugLocking = cf.Bool("debug_locking", false)
	this.Timeout = cf.Duration("timeout", 5*time.Second)
	this.MaxFlightEntries = cf.Int("max_flight_entries", 100000)
	this.LockExpires = cf.Duration("lock_expires", time.Second*30)
	this.Job = cf.String("job", "")
	this.March = cf.String("march", "")
	this.Pve = cf.String("pve", "")
}
