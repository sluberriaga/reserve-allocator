package allocator

import (
	"errors"
	"math/rand"
	"reserve/reserve"
	"sync"
	"time"
)

type DB struct {
	reserves     map[int64]reserve.Reserve
	mu           sync.Mutex
	splitDelay   time.Duration
	reserveDelay time.Duration
}

func (db *DB) List(userID uint64) []reserve.Reserve {
	db.mu.Lock()
	defer db.mu.Unlock()

	var userReserves []reserve.Reserve

	for _, reserveEntry := range db.reserves {
		if reserveEntry.UserID == userID {
			userReserves = append(userReserves, reserveEntry)
		}
	}

	return userReserves
}

func (db *DB) Release(reserveID int64) {
	db.mu.Lock()
	defer db.mu.Unlock()

	for i, reserveEntry := range db.reserves {
		if reserveEntry.ID == reserveID {
			reserveToRelease, _ := db.reserves[i]
			reserveToRelease.Status = "released"
			db.reserves[i] = reserveToRelease
			break
		}
	}

	return
}

func (db *DB) Insert(request reserve.ReserveRequest) (reserve.Reserve, error) {
	db.mu.Lock()
	defer db.mu.Unlock()
	time.Sleep(db.reserveDelay)

	ID := rand.Int63n(1000000)

	var version = "initial_tbs"
	db.reserves[ID] = reserve.Reserve{
		ID:                ID,
		Version:           &version,
		TTL:               nil,
		ExternalReference: request.Body.ExternalReference,
		IdempotencyKey:    request.IdempotencyKey,
		Reason:            request.Body.Reason,
		Mode:              request.Body.Mode,
		Amount:            request.Body.Amount,
		ClientID:          request.ClientID,
		UserID:            request.UserID,
		Status:            "reserved",
		DateCreated:       time.Now().String(),
		LastModified:      time.Now().String(),
	}

	return db.reserves[ID], nil
}

func (db *DB) Split(request reserve.ReserveRequest, toSplitReserveID int64) (reserve.Reserve, reserve.Reserve, error) {
	db.mu.Lock()
	defer db.mu.Unlock()
	time.Sleep(db.splitDelay)

	// request reserve.ReserveRequest, toSplitReserveID int64,
	originalReserve, ok := db.reserves[toSplitReserveID]
	if !ok {
		return reserve.Reserve{}, reserve.Reserve{}, errors.New("could not find reserve to split")
	}

	if originalReserve.Amount <= request.Body.Amount {
		return reserve.Reserve{}, reserve.Reserve{}, errors.New("could not split reserve")
	}

	originalReserve.Status = "released"
	db.reserves[toSplitReserveID] = originalReserve

	newParentReserveID := rand.Int63n(1000000)
	newSplittedReserveID := rand.Int63n(1000000)

	var version = "splitted_rest"
	newParentReserve := reserve.Reserve{
		ID:                newParentReserveID,
		Version:           &version,
		TTL:               nil,
		ExternalReference: originalReserve.ExternalReference,
		IdempotencyKey:    originalReserve.IdempotencyKey,
		Reason:            originalReserve.Reason,
		Mode:              originalReserve.Mode,
		Amount:            originalReserve.Amount - request.Body.Amount,
		ClientID:          request.ClientID,
		UserID:            request.UserID,
		Status:            "reserved",
		DateCreated:       time.Now().String(),
		LastModified:      time.Now().String(),
	}
	db.reserves[newParentReserveID] = newParentReserve

	var versionS = "splitted"
	newSplittedReserve := reserve.Reserve{
		ID:                newSplittedReserveID,
		Version:           &versionS,
		TTL:               nil,
		ExternalReference: request.Body.ExternalReference,
		IdempotencyKey:    request.IdempotencyKey,
		Reason:            request.Body.Reason,
		Mode:              request.Body.Mode,
		Amount:            request.Body.Amount,
		ClientID:          request.ClientID,
		UserID:            request.UserID,
		Status:            "reserved",
		DateCreated:       time.Now().String(),
		LastModified:      time.Now().String(),
	}
	db.reserves[newSplittedReserveID] = newSplittedReserve

	return newParentReserve, newSplittedReserve, nil
}
