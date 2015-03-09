package actor

import (
	"database/sql"
	"github.com/nicholaskh/actor/config"
	"github.com/nicholaskh/golib/breaker"
	log "github.com/nicholaskh/log4go"
	"github.com/nicholaskh/metrics"
	_ "github.com/go-sql-driver/mysql"
	"time"
)

type MysqlPoller struct {
	interval             time.Duration
	slowQueryThreshold   time.Duration
	manyWakeupsThreshold int

	stopChan chan bool

	mysql   *sql.DB
	breaker *breaker.Consecutive

	jobQueryStmt   *sql.Stmt
	marchQueryStmt *sql.Stmt
	pveQueryStmt   *sql.Stmt

	latency metrics.Histogram
}

func NewMysqlPoller(my *config.ConfigMysqlInstance,
	cf *config.ConfigMysql) (*MysqlPoller, error) {
	this := new(MysqlPoller)
	this.interval = cf.Interval
	this.slowQueryThreshold = cf.SlowThreshold
	this.manyWakeupsThreshold = cf.ManyWakeupsThreshold
	this.breaker = &breaker.Consecutive{
		FailureAllowance: cf.Breaker.FailureAllowance,
		RetryTimeout:     cf.Breaker.RetryTimeout}

	this.stopChan = make(chan bool)

	this.latency = metrics.NewHistogram(
		metrics.NewExpDecaySample(1028, 0.015))
	metrics.Register("latency.mysql", this.latency)

	var err error
	this.mysql, err = sql.Open("mysql", my.DSN())
	if err != nil {
		return nil, err
	}

	err = this.mysql.Ping()
	if err != nil {
		this.mysql = nil
		return nil, err
	}

	this.mysql.SetMaxIdleConns(1)
	this.mysql.SetMaxOpenConns(1)

	log.Debug("mysql connected: %s", my.DSN())

	if cf.Query.Job != "" {
		this.jobQueryStmt, err = this.mysql.Prepare(cf.Query.Job)
		if err != nil {
			log.Critical("db prepare err: %s", err.Error())
			return nil, err
		}
	}

	if cf.Query.March != "" {
		this.marchQueryStmt, err = this.mysql.Prepare(cf.Query.March)
		if err != nil {
			return nil, err
		}
	}

	if cf.Query.Pve != "" {
		this.pveQueryStmt, err = this.mysql.Prepare(cf.Query.Pve)
		if err != nil {
			return nil, err
		}
	}

	return this, nil
}

func (this *MysqlPoller) Poll(ch chan<- Wakeable) {
	if this.jobQueryStmt != nil {
		go this.doPoll("job", ch)
		defer this.jobQueryStmt.Close()
	}

	if this.marchQueryStmt != nil {
		go this.doPoll("march", ch)
		defer this.marchQueryStmt.Close()
	}

	if this.pveQueryStmt != nil {
		go this.doPoll("pve", ch)
		defer this.pveQueryStmt.Close()
	}

	<-this.stopChan
}

func (this *MysqlPoller) doPoll(typ string, ch chan<- Wakeable) {
	ticker := time.NewTicker(this.interval)
	defer ticker.Stop()

	var ws []Wakeable
	for now := range ticker.C {
		ws = this.fetchWakeables(typ, now)
		if len(ws) == 0 {
			continue
		}

		if len(ws) > this.manyWakeupsThreshold {
			log.Warn("too many wakes[%s] #%d: %+v", typ, len(ws), ws)
		} else {
			log.Debug("wakes[%s] #%d: %+v", typ, len(ws), ws)
		}

		for _, w := range ws {
			ch <- w
		}
	}
}

func (this *MysqlPoller) fetchWakeables(typ string, dueTime time.Time) (ws []Wakeable) {
	ws = make([]Wakeable, 0, 100)
	var stmt *sql.Stmt
	switch typ {
	case "job":
		stmt = this.jobQueryStmt

	case "march":
		stmt = this.marchQueryStmt

	case "pve":
		stmt = this.pveQueryStmt
	}

	rows, err := this.Query(stmt, dueTime.Unix())
	if err != nil {
		log.Error("db query: %s", err.Error())

		return
	}

	this.latency.Update(time.Since(dueTime).Nanoseconds() / 1e6)

	switch typ {
	case "job":
		for rows.Next() {
			var w Job
			err = rows.Scan(&w.Uid)
			if err != nil {
				log.Error("db scan: %s", err.Error())
				continue
			}

			ws = append(ws, &w)
		}

	case "march":
		for rows.Next() {
			var w March
			err = rows.Scan(&w.Uid, &w.MarchId, &w.OppUid,
				&w.K, &w.X1, &w.Y1, &w.Type,
				&w.State, &w.EndTime)
			if err != nil {
				log.Error("db scan: %s", err.Error())
				continue
			}

			ws = append(ws, &w)
		}

	case "pve":
		for rows.Next() {
			var w Pve
			err = rows.Scan(&w.Uid, &w.MarchId, &w.State, &w.EndTime)
			if err != nil {
				log.Error("db scan: %s", err.Error())
				continue
			}

			ws = append(ws, &w)
		}
	}

	rows.Close()
	return
}

func (this *MysqlPoller) Query(stmt *sql.Stmt,
	args ...interface{}) (rows *sql.Rows, err error) {
	//log.Debug("%+v, args=%+v", *stmt, args)

	if this.breaker.Open() {
		return nil, ErrCircuitOpen
	}

	t0 := time.Now()
	rows, err = stmt.Query(args...)
	if err != nil {
		this.breaker.Fail()
		return
	} else {
		this.breaker.Succeed()
	}

	elapsed := time.Since(t0)
	if elapsed > this.slowQueryThreshold {
		log.Warn("slow query:%s, %+v", elapsed, *stmt)
	}

	return
}

func (this *MysqlPoller) Stop() {
	if this.mysql != nil {
		this.mysql.Close()
	}
	close(this.stopChan)
}
