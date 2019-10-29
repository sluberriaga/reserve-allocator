package allocator

import (
	"errors"
	"math/rand"
	"reserve/reserve"
	"sync"
	"time"
)

var db DB

func init() {
	db = DB{
		reserves:     map[int64]reserve.Reserve{},
		mu:           sync.Mutex{},
		splitDelay:   35 * time.Millisecond,
		reserveDelay: 70 * time.Millisecond,
	}
}

type client struct {
	percentageSplitFailure      int
	percentageAllocationFailure int
}

func newClient() client {
	return client{0, 0}
}

func (c *client) ListReservesForUser(userID uint64) []reserve.Reserve {
	return db.List(userID)
}

func (c *client) ReleaseReserve(reserveID int64) {
	db.Release(reserveID)

	return
}

func (c *client) PostReserve(request reserve.ReserveRequest, factor int) (reserve.Reserve, error) {
	if rand.Intn(101) > c.percentageAllocationFailure {
		request.Body.Amount = request.Body.Amount / 100 * int64(factor)
		newReserve, err := db.Insert(request)
		newReserve.Amount = newReserve.Amount * 100

		return newReserve, err
	}

	return reserve.Reserve{}, errors.New("generic error")
}

func (c *client) SplitReserve(
	request reserve.ReserveRequest, toSplitReserveID int64,
) (
	newParentReserve reserve.Reserve, newSplittedReserve reserve.Reserve, err error,
) {
	request.Body.Amount = request.Body.Amount / 100
	if rand.Intn(101) > c.percentageSplitFailure {
		newOriginal, newSplitted, err := db.Split(request, toSplitReserveID)
		newOriginal.Amount = newOriginal.Amount * 100
		newSplitted.Amount = newSplitted.Amount * 100

		return newOriginal, newSplitted, err
	}

	return reserve.Reserve{}, reserve.Reserve{}, errors.New("generic error")

}
