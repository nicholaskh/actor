package config

import (
	conf "github.com/nicholaskh/jsconf"
)

type ConfigWorkerRtm struct {
	MaxProcs     int
	Backlog      int
	MaxRetries   int
	PrimaryHosts []string
	BackupHosts  []string
	ProjectId    int32
	SecretKey    string
}

func (this *ConfigWorkerRtm) loadConfig(cf *conf.Conf) {
	this.MaxProcs = cf.Int("max_procs", 50)
	this.Backlog = cf.Int("backlog", 200)
	this.MaxRetries = cf.Int("max_retries", 5)
	this.PrimaryHosts = make([]string, 0)
	for _, host := range cf.StringList("primary_hosts", nil) {
		this.PrimaryHosts = append(this.PrimaryHosts, host)
	}
	this.BackupHosts = make([]string, 0)
	for _, host := range cf.StringList("backup_hosts", nil) {
		this.BackupHosts = append(this.BackupHosts, host)
	}
	this.ProjectId = int32(cf.Int("project_id", 0))
	this.SecretKey = cf.String("secret_key", "")
}

func (this *ConfigWorkerRtm) Enabled() bool {
	return len(this.PrimaryHosts) > 0

}
