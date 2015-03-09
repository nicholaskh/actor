package actor

import (
	log "github.com/funkygao/log4go"
	"github.com/kr/beanstalk"
	"time"
)

type BeanstalkdPoller struct {
	conn *beanstalk.Conn
}

func NewBeanstalkdPoller(addr string, watchTubes ...string) (this *BeanstalkdPoller,
	err error) {
	this = new(BeanstalkdPoller)
	this.conn, err = beanstalk.Dial("tcp", addr)
	if err != nil {
		return
	}
	this.conn.TubeSet = *beanstalk.NewTubeSet(this.conn, watchTubes...)
	return
}

func (this *BeanstalkdPoller) Poll(ch chan<- Wakeable) {
	var (
		id   uint64
		body []byte
		err  error
		push *Push
	)
	for {
		id, body, err = this.conn.Reserve(time.Hour * 100) // TODO
		if err != nil {
			log.Error("beanstalk.reserve: %v", err)
			continue
		}

		this.conn.Delete(id) // FIXME ackFail or ackSuccess

		push = new(Push) // TODO mem pool
		push.Body = body
		push.Id = id

		ch <- push
	}
}

func (this *BeanstalkdPoller) Stop() {
	if this.conn != nil {
		this.conn.Close()
	}
}
