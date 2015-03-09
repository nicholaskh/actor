package config

import (
	"fmt"
	conf "github.com/funkygao/jsconf"
)

type ConfigBeanstalk struct {
	Breaker ConfigBreaker
	Servers map[string]*ConfigBeanstalkInstance // key is tube
}

func (this *ConfigBeanstalk) loadConfig(cf *conf.Conf) {
	section, err := cf.Section("breaker")
	if err == nil {
		this.Breaker.loadConfig(section)
	}

	this.Servers = make(map[string]*ConfigBeanstalkInstance)
	for i := 0; i < len(cf.List("servers", nil)); i++ {
		section, err := cf.Section(fmt.Sprintf("servers[%d]", i))
		if err != nil {
			panic(err)
		}

		server := new(ConfigBeanstalkInstance)
		server.loadConfig(section)
		this.Servers[server.Tube] = server
	}
}

type ConfigBeanstalkInstance struct {
	Tube       string
	ServerAddr string
}

func (this *ConfigBeanstalkInstance) loadConfig(cf *conf.Conf) {
	this.Tube = cf.String("tube", "")
	this.ServerAddr = cf.String("server", "")
}
