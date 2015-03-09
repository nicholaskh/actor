package actor

const (
	GEOHASH_SHIFT = 12
)

const (
	API_OP_LOCK   = "lock"
	API_OP_UNLOCK = "unlock"
	API_TYPE_USER = "user"
	API_TYPE_TILE = "tile"
)

const (
	MARCH_RALLY  = "rally"
	MARCH_ENCAMP = "encamping"
	MARCH_DONE   = "done"
)

const (
	LOCKER_REASON = "actor.lock"
	LOCKER_LOCK   = "lock"
	LOCKER_UNLOCK = "unlock"
)

const (
	ticket_user int64 = iota + 1
	ticket_alliance
	ticket_chat_room
)
