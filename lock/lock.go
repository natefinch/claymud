package lock

import (
	"sort"
	"sync"

	"github.com/natefinch/claymud/util"
)

var _ sync.Locker = IdLocker(nil)

// IdLocker represents lockable object that also has an Id.
type IdLocker interface {
	Lock()
	Unlock()
	Id() util.Id
}

// All locks all the locks in the list, in Id order, to implement lock ordering
// and avoid deadlocks.
func All(locks []IdLocker) {
	sort.Sort(byId(locks))
	for _, l := range locks {
		l.Lock()
	}
}

// UnlockAll unlocks all the locks in the list, in reverse Id order.
func UnlockAll(locks []IdLocker) {
	sort.Sort(sort.Reverse(byId(locks)))
	for _, l := range locks {
		l.Unlock()
	}
}

// byId is implements Sort.Interface, sorting by the lock's Id
type byId []IdLocker

func (b byId) Len() int {
	return len(b)
}

func (b byId) Swap(i, j int) {
	b[i], b[j] = b[j], b[i]
}

func (b byId) Less(i, j int) bool {
	return b[i].Id() < b[j].Id()
}
