package config

import (
	conf "github.com/funkygao/jsconf"
	"time"
)

type ConfigBreaker struct {
	FailureAllowance uint
	RetryTimeout     time.Duration
}

func (this *ConfigBreaker) loadConfig(cf *conf.Conf) {
	this.FailureAllowance = uint(cf.Int("failure_allowance", 5))
	this.RetryTimeout = cf.Duration("retry_timeout", time.Second*10)
}

func (this *ConfigBreaker) Enabled() bool {
	return this.FailureAllowance > 0 && this.RetryTimeout.Seconds() > 0
}
