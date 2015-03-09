package actor

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"
)

type March struct {
	Uid     int64          `json:"uid"`
	MarchId int64          `json:"march_id"`
	Type    sql.NullString `json:"type"`
	OppUid  sql.NullInt64  `json:"-"`
	State   string         `json:"-"`
	K       int16          `json:"-"` // kingdom id
	X1      int16          `json:"-"`
	Y1      int16          `json:"-"`
	EndTime time.Time      `json:"-"`
}

func (this *March) Marshal() []byte {
	b, _ := json.Marshal(*this)
	return b
}

func (this *March) Ignored() bool {
	return this.State == MARCH_DONE || this.State == MARCH_ENCAMP
}

// rally里有自己向自己行军的情况，这时像自己打自己，只锁该uid一次，否则死锁
func (this *March) IsOpponentSelf() bool {
	return this.OppUid.Valid &&
		this.Uid == this.OppUid.Int64
}

func (this *March) DueTime() time.Time {
	return this.EndTime
}

func (this *March) String() string {
	return fmt.Sprintf("March{uid:%d, opp:%d, mid:%d, type:%s, state:%s, (%d,%d,%d)}",
		this.Uid, this.OppUid.Int64,
		this.MarchId, this.Type.String, this.State, this.K, this.X1, this.Y1)
}
