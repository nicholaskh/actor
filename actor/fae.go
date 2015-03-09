package actor

import (
	"github.com/funkygao/fae/config"
	"github.com/funkygao/fae/servant/gen-go/fun/rpc"
	"github.com/funkygao/fae/servant/proxy"
	"github.com/funkygao/golib/ip"
	log "github.com/funkygao/log4go"
	"strconv"
	"sync/atomic"
)

type FaeExecutor struct {
	proxy *proxy.Proxy

	myIp string
	txn  int64
}

func NewFaeExecutor(poolSize int) *FaeExecutor {
	this := new(FaeExecutor)
	cf := config.NewDefaultProxy()
	cf.PoolCapacity = poolSize
	this.proxy = proxy.New(cf)
	this.myIp = ip.LocalIpv4Addrs()[0]
	return this
}

func (this *FaeExecutor) StartCluster() {
	go this.proxy.StartMonitorCluster()
	this.proxy.AwaitClusterTopologyReady()

	log.Info("fae cluster ready")
}

func (this *FaeExecutor) Warmup() {
	this.proxy.Warmup()
}

func (this *FaeExecutor) NewContext(reason string) *rpc.Context {
	ctx := rpc.NewContext()
	ctx.Reason = reason
	rid := atomic.AddInt64(&this.txn, 1)
	ctx.Rid = strconv.FormatInt(rid, 10)
	ctx.Host = this.myIp
	return ctx
}
