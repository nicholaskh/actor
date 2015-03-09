package config

import (
	conf "github.com/nicholaskh/jsconf"
)

type ConfigWorkerPnb struct {
	MaxProcs int
	Backlog  int

	PublishKey   string
	SubscribeKey string
	SecretKey    string
	CipherKey    string
	UseSsl       bool
}

func (this *ConfigWorkerPnb) loadConfig(cf *conf.Conf) {
	this.MaxProcs = cf.Int("max_procs", 50)
	this.Backlog = cf.Int("backlog", 200)
	this.PublishKey = cf.String("publish_key", "")
	this.SubscribeKey = cf.String("subscribe_key", "")
	this.SecretKey = cf.String("secret_key", "")
	this.CipherKey = cf.String("cipher_key", "")
	this.UseSsl = cf.Bool("use_ssl", false)
}
