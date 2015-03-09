package actor

import (
	"github.com/funkygao/lockkey"
	log "github.com/funkygao/log4go"
)

type Locker []string

func NewLocker() Locker {
	this := make([]string, 0)
	return this
}

func (this *Locker) LockUser(uid int64) bool {
	return this.acquireLock(lockkey.User(uid))
}

func (this *Locker) LockAttackee(k, x, y int16) bool {
	return this.acquireLock(lockkey.Attackee(k, x, y))
}

func (this *Locker) acquireLock(key string) (success bool) {
	svt, err := fae.proxy.ServantByKey(key)
	if err != nil {
		log.Error("fae.lock[%s]: %s", key, err.Error())
		return false
	}

	if success, err = svt.GmLock(fae.NewContext(LOCKER_REASON), LOCKER_LOCK, key); success {
		*this = append(*this, key)
	}
	if err != nil {
		log.Error("fae.lock[%s]: %s {ok^%+v err^%s}", key, svt.Addr(), success, err)
		svt.Close()
	} else {
		log.Debug("fae.lock[%s]: %s {ok^%+v}", key, svt.Addr(), success)
	}

	svt.Recycle()

	return
}

func (this *Locker) ReleaseAll() {
	for _, key := range *this {
		svt, err := fae.proxy.ServantByKey(key)
		if err != nil {
			log.Error("fae.unlock[%s]: %s", key, err.Error())
			continue
		}

		if err = svt.GmUnlock(fae.NewContext(LOCKER_REASON),
			LOCKER_UNLOCK, key); err != nil {
			log.Error("fae.unlock[%s]: %s", key, err.Error())
		} else {
			log.Debug("fae.unlock[%s]: %s", key, svt.Addr())
		}

		svt.Recycle()
	}
}
