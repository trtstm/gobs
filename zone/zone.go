package zone

import (
	"sync"
)

type Zone struct {
	name string

	pidToBillerIdLock sync.RWMutex
	pidToBillerId     map[uint]uint
}

func NewZone(name string) *Zone {
	zone := Zone{name: name, pidToBillerId: map[uint]uint{}}

	return &zone
}

func (z *Zone) ToBillerId(pid uint) (uint, bool) {
	z.pidToBillerIdLock.RLock()
	defer z.pidToBillerIdLock.RUnlock()

	billerId, ok := z.pidToBillerId[pid]
	return billerId, ok
}

func (z *Zone) AddPlayer(pid uint, billerId uint) bool {
	z.pidToBillerIdLock.Lock()
	defer z.pidToBillerIdLock.Unlock()

	_, ok := z.pidToBillerId[pid]
	if ok {
		return false
	}

	z.pidToBillerId[pid] = billerId
	return true
}

func (z *Zone) RemovePlayer(pid uint) {
	z.pidToBillerIdLock.Lock()
	defer z.pidToBillerIdLock.Unlock()

	delete(z.pidToBillerId, pid)
}
