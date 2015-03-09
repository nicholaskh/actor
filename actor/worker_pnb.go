package actor

import (
	"github.com/fate2013/pubnub-go/messaging"
	"github.com/nicholaskh/actor/config"
	log "github.com/nicholaskh/log4go"
	"github.com/nicholaskh/metrics"
	"sync"

	"time"
)

type WorkerPnb struct {
	config  *config.ConfigWorkerPnb
	backlog chan *Push
	pnb     *messaging.Pubnub
	latency metrics.Histogram
}

func NewPnbWorker(config *config.ConfigWorkerPnb) *WorkerPnb {
	this := new(WorkerPnb)
	this.config = config
	this.backlog = make(chan *Push, config.Backlog)
	this.pnb = messaging.NewPubnub(this.config.PublishKey,
		this.config.SubscribeKey, this.config.SecretKey,
		this.config.CipherKey, this.config.UseSsl, "")
	this.latency = metrics.NewHistogram(
		metrics.NewExpDecaySample(1028, 0.015))
	if this.Enabled() {
		metrics.Register("latency.pnb", this.latency)
	}
	return this
}

func (this *WorkerPnb) Enabled() bool {
	return this.config.PublishKey != "" && this.config.MaxProcs > 0
}

func (this *WorkerPnb) Start() {
	if !this.Enabled() {
		return
	}

	var wg sync.WaitGroup
	for i := 0; i < this.config.MaxProcs; i++ {
		wg.Add(1)

		go func() {
			wg.Done()

			for {
				select {
				case push := <-this.backlog:
					this.doPublish(push)
				}
			}
		}()
	}

	wg.Wait()
	log.Info("worker[pnb] ready")
}

func (this *WorkerPnb) Wake(w Wakeable) {
	push := w.(*Push)
	select {
	case this.backlog <- push:
	default:
		log.Warn("rtm backlog full, discarded: %s", string(push.Body))
	}
}

func (this *WorkerPnb) doPublish(push *Push) {
	msg, _, toIds := push.Unmarshal()
	for _, channel := range toIds {
		go func(channel string) {
			successChannel := make(chan []byte)
			errorChannel := make(chan []byte)
			t1 := time.Now()
			go this.pnb.Publish(channel, msg, successChannel, errorChannel)
			select {
			case <-successChannel:
				this.latency.Update(time.Since(t1).Nanoseconds() / 1e6)
				log.Trace("pnb %s: {to^%+v msg^%s", time.Since(t1), channel, msg)

			case err := <-errorChannel:
				log.Error("pnb %s: %s", time.Since(t1), string(err))
			}
		}(channel)
	}

}
