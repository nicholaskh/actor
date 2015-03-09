package actor

import (
	"github.com/fate2013/go-rtm/proxy"
	"github.com/fate2013/go-rtm/servergated"
	"github.com/nicholaskh/actor/config"
	"github.com/nicholaskh/golib/pool"
	log "github.com/nicholaskh/log4go"
	"sync/atomic"
	"time"
)

type rtmConn struct {
	id uint64
	*servergated.ServerGatedServiceClient
	pool *rtmPool
}

func (this *rtmConn) Close() {
	if this.ServerGatedServiceClient != nil {
		this.ServerGatedServiceClient.Transport.Close()
	}
}

func (this *rtmConn) Id() uint64 {
	return this.id
}

func (this *rtmConn) IsOpen() bool {
	return this.Transport.IsOpen()
}

func (this *rtmConn) Recycle() {
	if this.Transport.IsOpen() {
		this.pool.pool.Put(this)
	} else {
		this.pool.pool.Kill(this)
		this.pool.pool.Put(nil)
	}
}

type rtmPool struct {
	cf         *config.ConfigWorkerRtm
	pool       *pool.ResourcePool
	nextConnId uint64
	txn        int64
}

func newRtmPool(cf *config.ConfigWorkerRtm) (this *rtmPool) {
	this = &rtmPool{cf: cf}
	if cf.MaxProcs == 0 || len(cf.PrimaryHosts) == 0 {
		log.Warn("rtm disabled")
		return
	}

	factory := func() (pool.Resource, error) {
		conn := &rtmConn{
			pool: this,
			id:   atomic.AddUint64(&this.nextConnId, 1),
		}

		var err error
		t1 := time.Now()
		conn.ServerGatedServiceClient, err = proxy.NewRtmClient(
			this.cf.PrimaryHosts[0])
		if err == nil {
			log.Debug("rtm connected[%d]: %s", conn.id, time.Since(t1))
		}

		return conn, err
	}

	borrowMaxSeconds := 9
	this.pool = pool.NewResourcePool("rtm", factory,
		this.cf.MaxProcs, this.cf.MaxProcs, 0,
		time.Second*10, borrowMaxSeconds)

	return
}

func (this *rtmPool) Close() {
	this.pool.Close()
}

func (this *rtmPool) Get() (*rtmConn, error) {
	rtm, err := this.pool.Get()
	if err != nil {
		return nil, err
	}

	return rtm.(*rtmConn), nil
}

func (this *rtmPool) NextTxn() int64 {
	return atomic.AddInt64(&this.txn, 1)
}
