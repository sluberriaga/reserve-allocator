package concurrency

import (
	"github.com/gin-gonic/gin"
	"reserve/reserve"
	"strconv"
	"time"
)

type Service struct {
	heatMap    heatMap
	decayDelay time.Duration
	decay      uint
	heat       int
	concurrencyThresshold uint64
}

func NewService(config reserve.ConcurrencyConfig) Service {
	return Service{
		heatMap:    newHeatMap(),
		decayDelay: config.DecayDelay,
		decay:      config.Decay,
		heat:       config.Heat,
		concurrencyThresshold: config.ConcurrrentThresshold,
	}
}

func (s *Service) CheckConcurrency(entryID uint64) bool {
	value, err := s.heatMap.Load(entryID)
	if err != nil {
		return false
	}

	return value > s.concurrencyThresshold
}

func (s *Service) RegisterEntryMiddleware(c *gin.Context) {
	c.Next()

	userIDParam := c.Param("user_id")
	userID, _ := strconv.ParseUint(userIDParam, 10, 64)

	shouldRegisterExpirer := false

	err := s.heatMap.LoadAndStore(
		userID,
		0,
		func(curr uint64) uint64 {
			nextVal := curr + uint64(s.heat)

			if curr == uint64(s.heat) {
				shouldRegisterExpirer = true
			}

			return nextVal
		},
	)
	if err != nil {
		return
	}

	if shouldRegisterExpirer {
		go s.expirer(userID)
	}

	return
}

func (s *Service) expirer(userID uint64) {
	timeout := time.After(s.decayDelay)
	for {
		<-timeout

		shouldExit := false
		err := s.heatMap.LoadAndStore(
			userID,
			0,
			func(curr uint64) uint64 {
				nextVal := curr - uint64(s.decay)
				if nextVal <= 0 {
					shouldExit = true
					return 0
				}
				return nextVal
			},
		)
		if err != nil || shouldExit {
			return
		}

		timeout = time.After(s.decayDelay)
	}
}
