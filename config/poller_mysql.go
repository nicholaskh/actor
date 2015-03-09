package config

import (
	"fmt"
	conf "github.com/nicholaskh/jsconf"
	"time"
)

type ConfigMysql struct {
	Interval             time.Duration
	ConnectTimeout       time.Duration // part of DSN
	SlowThreshold        time.Duration
	ManyWakeupsThreshold int // will log.warn if exceeded

	Query   ConfigMysqlQuery
	Breaker ConfigBreaker

	Servers map[string]*ConfigMysqlInstance // key is pool
}

func (this *ConfigMysql) loadConfig(cf *conf.Conf) {
	this.Interval = cf.Duration("interval", time.Second)
	this.ConnectTimeout = cf.Duration("connect_timeout", 4*time.Second)
	this.SlowThreshold = cf.Duration("slow_threshold", 1*time.Second)
	this.ManyWakeupsThreshold = cf.Int("many_wakeups_threshold", 200)

	section, err := cf.Section("query")
	if err != nil {
		panic(err)
	}
	this.Query.loadConfig(section)

	section, err = cf.Section("breaker")
	if err == nil {
		this.Breaker.loadConfig(section)
	}
	this.Servers = make(map[string]*ConfigMysqlInstance)
	for i := 0; i < len(cf.List("servers", nil)); i++ {
		section, err := cf.Section(fmt.Sprintf("servers[%d]", i))
		if err != nil {
			panic(err)
		}

		server := new(ConfigMysqlInstance)
		server.ConnectTimeout = this.ConnectTimeout
		server.loadConfig(section)
		this.Servers[server.Pool] = server
	}

}

type ConfigMysqlQuery struct {
	Job   string
	March string
	Pve   string
}

func (this *ConfigMysqlQuery) loadConfig(cf *conf.Conf) {
	this.Job = cf.String("job", "")
	this.March = cf.String("march", "")
	this.Pve = cf.String("pve", "")
	if this.Job == "" &&
		this.March == "" &&
		this.Pve == "" {
		panic("empty mysql query")
	}
}

type ConfigMysqlInstance struct {
	ConnectTimeout time.Duration

	Pool    string
	Host    string
	Port    string
	User    string
	Pass    string
	DbName  string
	Charset string

	dsn string
}

func (this *ConfigMysqlInstance) loadConfig(section *conf.Conf) {
	this.Pool = section.String("pool", "")
	this.Host = section.String("host", "")
	this.Port = section.String("port", "3306")
	this.DbName = section.String("db", "")
	this.User = section.String("username", "")
	this.Pass = section.String("password", "")
	this.Charset = section.String("charset", "utf8")
	if this.Host == "" ||
		this.Port == "" ||
		this.Pool == "" ||
		this.DbName == "" {
		panic("required field missing")
	}

	this.dsn = ""
	if this.User != "" {
		this.dsn = this.User + ":"
		if this.Pass != "" {
			this.dsn += this.Pass
		}
	}
	this.dsn += fmt.Sprintf("@tcp(%s:%s)/%s?", this.Host, this.Port, this.DbName)
	this.dsn += "autocommit=true" // we are not using transaction
	this.dsn += fmt.Sprintf("&timeout=%s", this.ConnectTimeout)
	if this.Charset != "utf8" { // driver default utf-8
		this.dsn += "&charset=" + this.Charset
	}
	this.dsn += "&parseTime=true" // parse db timestamp automatically
}

func (this *ConfigMysqlInstance) String() string {
	return this.DSN()
}

func (this *ConfigMysqlInstance) DSN() string {
	return this.dsn
}
