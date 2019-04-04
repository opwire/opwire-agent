package services

import (
	"context"
	"runtime"
	"golang.org/x/sync/semaphore"
	"golang.org/x/sync/singleflight"
)

type ReqRestrictor struct {
	context context.Context
	semaphore *semaphore.Weighted
	flightGroup *singleflight.Group
	options *ReqRestrictorOptions
}

type ReqRestrictorOptions interface {
	ConcurrentLimitEnabled() bool
	ConcurrentLimitTotal() int
	SingleFlightEnabled() bool
}

func NewReqRestrictor(opts ReqRestrictorOptions) (*ReqRestrictor, error) {
	rr := new(ReqRestrictor)

	rr.context = context.TODO()

	// create the semaphore
	limitEnabled := false
	limitTotal := 0
	if opts != nil {
		limitEnabled = opts.ConcurrentLimitEnabled()
		limitTotal = opts.ConcurrentLimitTotal()
	}
	if limitTotal == 0 {
		limitTotal = runtime.GOMAXPROCS(0)
	}
	if limitEnabled && limitTotal > 0 {
		rr.semaphore = semaphore.NewWeighted(int64(limitTotal))
	}

	// create the singleflight
	sfEnabled := false
	if opts != nil {
		sfEnabled = opts.SingleFlightEnabled()
	}
	if sfEnabled {
		rr.flightGroup = new(singleflight.Group)
	}

	return rr, nil
}

func (rr *ReqRestrictor) HasSemaphore() bool {
	return rr.semaphore != nil
}

func (rr *ReqRestrictor) Acquire(weight int) error {
	if rr.semaphore == nil {
		return nil
	}
	return rr.semaphore.Acquire(rr.context, int64(weight))
}

func (rr *ReqRestrictor) Release(weight int) {
	if rr.semaphore == nil {
		return
	}
	rr.semaphore.Release(int64(weight))
}

func (rr *ReqRestrictor) HasSingleFlight() bool {
	return rr.semaphore != nil
}

func (rr *ReqRestrictor) Filter(groupKey string, action func() (interface{}, error)) (interface{}, error, bool) {
	if rr.flightGroup == nil {
		out, err := action()
		return out, err, false
	}
	return rr.flightGroup.Do(groupKey, action)
}
