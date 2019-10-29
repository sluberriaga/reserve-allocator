package allocator

import (
	"fmt"
	"github.com/emirpasic/gods/maps/treebidimap"
	"github.com/gin-gonic/gin"
	"reserve/reserve"
	"strconv"
	"time"
)

type Service struct {
	registry           registry
	client             client
	overshootFactor    int
	maxRetryAllocation int
	reserveLifetime    time.Duration
}

func NewService(config reserve.AllocatorConfig) Service {
	return Service{
		newRegistry(),
		newClient(),
		config.OvershootFactor,
		config.MaxRetryAllocation,
		config.ReserveLifetime,
	}
}

func (s *Service) AllocateReserve(
	request reserve.ReserveRequest, isConcurrent bool,
) (
	reserve.Reserve, error,
) {
	var allocatedReserve reserve.Reserve

	if isConcurrent {
		allocErr := s.registry.LoadAndStore(request.UserID, func(reserves treebidimap.Map) treebidimap.Map {

			shouldTryToReserveNew := reserves.Size() == 0

			for i := 0; i < s.maxRetryAllocation; i++ {
				if shouldTryToReserveNew {
					newReserve, err := s.client.PostReserve(request, 10)
					if err != nil {
						fmt.Println("Error posting reserve")

						return reserves
					}

					reserves.Put(time.Now(), newReserve)
				}

				var toRemove []time.Time
				var toAllocate []reserve.Reserve
				for _, reservesR := range reserves.Values() {
					parentReserve, ok := reservesR.(reserve.Reserve)
					if !ok {
						fmt.Println("Error parsing reserve")
						return reserves
					}

					if parentReserve.Amount > request.Body.Amount {
						var newParentReserve reserve.Reserve
						newParentReserve, allocatedReserve, _ = s.client.SplitReserve(request, parentReserve.ID)
						timeKeyR, found := reserves.GetKey(parentReserve)
						if !found {
							fmt.Println("Could not find key parent reserve")
							return reserves
						}

						timeKey, ok := timeKeyR.(time.Time)
						if !ok {
							fmt.Println("Could not find key parent reserve")
							return reserves
						}

						toRemove = append(toRemove, timeKey)
						toAllocate = append(toAllocate, newParentReserve)
					} else {
						shouldTryToReserveNew = true
						break
					}
				}

				for _, deletedEntry := range toRemove {
					reserves.Remove(deletedEntry)
				}

				for _, allocatedParentReserve := range toAllocate {
					reserves.Put(time.Now(), allocatedParentReserve)

					return reserves
				}

				shouldTryToReserveNew = true
			}

			return reserves
		})

		return allocatedReserve, allocErr
	}

	notConcurrentReserve, err := s.client.PostReserve(request, 1)
	if err != nil {
		return reserve.Reserve{}, err
	}
	version := "standalone"
	notConcurrentReserve.Version = &version

	return notConcurrentReserve, nil
}

func (s *Service) RegisterBucketExpirationMiddleware(c *gin.Context) {
	timeout := time.After(s.reserveLifetime)
	shouldExit := false

	userIDParam := c.Param("user_id")
	userID, _ := strconv.ParseUint(userIDParam, 10, 64)
	go func(userID uint64) {
		defer func (){
			if r := recover(); r != nil {
				fmt.Println("Recovered in f", r)
			}
		}()

		for {
			_, found, err := s.registry.Load(userID)
			if err != nil {
				fmt.Println("Error loading user buckets")
				return
			}

			if found {
				return
			}

			<-timeout

			allocErr := s.registry.LoadAndStore(userID, func(reserves treebidimap.Map) treebidimap.Map {
				currentTime := time.Now()

				var toRemove *time.Time
				for _, reserveR := range reserves.Keys() {
					reserveTime, ok := reserveR.(time.Time)
					if !ok {
						fmt.Println("Error parsing time bucket")
						return reserves
					}

					if currentTime.After(reserveTime.Add(s.reserveLifetime)) {
						reserveValue, _ := reserves.Get(reserveTime)
						reserveToRelease, ok := reserveValue.(reserve.Reserve)
						if !ok {
							fmt.Println("Error parsing reserve")
							return reserves
						}

						s.client.ReleaseReserve(reserveToRelease.ID)

						toRemove = &reserveTime
					}
				}

				if toRemove != nil {
					reserves.Remove(*toRemove)
				}

				if reserves.Size() == 0 {
					shouldExit = true
				}

				return reserves
			})
			if allocErr != nil || shouldExit {
				return
			}

			timeout = time.After(s.reserveLifetime)
		}
	}(userID)
}

func (s *Service) ListFromRegistry(userID uint64) []reserve.Reserve {
	var toReturn []reserve.Reserve
	reserves, _, _ := s.registry.Load(userID)

	for _, value := range reserves.Values() {
		if registryReserve, ok := value.(reserve.Reserve); ok {
			toReturn = append(toReturn, registryReserve)
		}
	}

	return toReturn
}

func (s *Service) ListFromDB(userID uint64) []reserve.Reserve {
	return s.client.ListReservesForUser(userID)
}
