package actor

import (
	"encoding/json"
	"fmt"
	"time"
)

type Job struct {
	Uid int64 `json:"uid"`
}

func (this *Job) Marshal() []byte {
	b, _ := json.Marshal(*this)
	return b
}

func (this *Job) DueTime() time.Time {
	return time.Now() // FIXME
}

func (this *Job) Ignored() bool {
	return false
}

func (this *Job) String() string {
	return fmt.Sprintf("Job{uid:%d}", this.Uid)
}
