package config

import (
	conf "github.com/nicholaskh/jsconf"
)

type ConfigWorker struct {
	Php ConfigWorkerPhp
	Pnb ConfigWorkerPnb
	Rtm ConfigWorkerRtm
}

func (this *ConfigWorker) loadConfig(cf *conf.Conf) {
	section, err := cf.Section("php")
	if err == nil {
		this.Php.loadConfig(section)
	} else {
		panic(err)
	}

	pnbExists := false
	section, err = cf.Section("pnb")
	if err == nil {
		this.Pnb.loadConfig(section)
		pnbExists = true
	}

	rtmExists := false
	section, err = cf.Section("rtm")
	if err == nil {
		this.Rtm.loadConfig(section)
		rtmExists = true
	}

	if !rtmExists && !pnbExists {
		panic("None of rtm and pnb")
	}

}
