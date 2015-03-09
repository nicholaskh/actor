package actor

import (
	"strconv"
	"strings"
	"time"
)

type Push struct {
	Id   uint64 // beanstalk job id
	Body []byte
}

func (this *Push) String() string {
	return ""
}

func (this *Push) DueTime() time.Time {
	return time.Now()
}

func (this *Push) Marshal() []byte {
	return nil
}

func (this *Push) Ignored() bool {
	return false
}

func (this *Push) Type() string {
	return string(this.Body[:1])
}

// TODO toIds is []int64
func (this *Push) Unmarshal() (msg string, fromId int64, toIds []string) {
	const SEP = "|"
	body := string(this.Body[2:])
	parts := strings.SplitN(body, SEP, 3)
	fromId, _ = strconv.ParseInt(parts[1], 0, 0)
	toIds = strings.Split(parts[0], ",")
	msg = parts[2]
	return
}
