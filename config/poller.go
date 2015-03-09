package config

import (
	conf "github.com/funkygao/jsconf"
)

type ConfigPoller struct {
	Mysql     ConfigMysql
	Beanstalk ConfigBeanstalk
}

func (this *ConfigPoller) loadConfig(cf *conf.Conf) {
	section, err := cf.Section("mysql")
	if err == nil {
		this.Mysql.loadConfig(section)
	}

	section, err = cf.Section("beanstalk")
	if err == nil {
		this.Beanstalk.loadConfig(section)
	}

	if len(this.Mysql.Servers)+len(this.Beanstalk.Servers) == 0 {
		panic("Zero poller in config")
	}

}
