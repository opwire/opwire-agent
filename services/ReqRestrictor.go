package services

import (
	"context"
	"crypto/sha256"
	"fmt"
	"log"
	"net"
	"net/http"
	"runtime"
	"strings"
	"golang.org/x/sync/semaphore"
	"golang.org/x/sync/singleflight"
)

type ReqRestrictor struct {
	context context.Context
	semaphore *semaphore.Weighted
	flightGroup *singleflight.Group
	flightPattern *SingleFlightPattern
	options *ReqRestrictorOptions
}

type ReqRestrictorOptions interface {
	ConcurrentLimitEnabled() bool
	ConcurrentLimitTotal() int
	SingleFlightEnabled() bool
	SingleFlightReqIdName() string
	SingleFlightByMethod() bool
	SingleFlightByPath() bool
	SingleFlightByBody() bool
	SingleFlightByHeaders() []string
	SingleFlightByQueries() []string
	SingleFlightByUserIP() bool
}

type SingleFlightPattern struct {
	ReqIdName string
	HasMethod bool
	HasPath bool
	HasHeaders []string
	HasQueries []string
	HasBody bool
	HasUserIP bool
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

	// create the singleflight pattern
	sfp := &SingleFlightPattern{}
	if opts != nil {
		sfp.ReqIdName = opts.SingleFlightReqIdName()
		sfp.HasMethod = opts.SingleFlightByMethod()
		sfp.HasPath = opts.SingleFlightByPath()
		sfp.HasHeaders = opts.SingleFlightByHeaders()
		sfp.HasQueries = opts.SingleFlightByQueries()
		sfp.HasBody = opts.SingleFlightByBody()
		sfp.HasUserIP = opts.SingleFlightByUserIP()
	}
	if sfEnabled {
		rr.flightPattern = sfp
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
	return rr.flightGroup != nil && rr.flightPattern != nil
}

func (rr *ReqRestrictor) Filter(groupKey string, action func() (interface{}, error)) (interface{}, error, bool) {
	if !rr.HasSingleFlight() {
		out, err := action()
		return out, err, false
	}
	return rr.flightGroup.Do(groupKey, action)
}

func (rr *ReqRestrictor) FilterByDigest(r *http.Request, action func() (interface{}, error)) (interface{}, error, bool) {
	var (
		out interface{}
		err error
		shared bool
	)
	if !rr.HasSingleFlight() {
		out, err = action()
		return out, err, false
	}
	p := rr.flightPattern
	if len(p.ReqIdName) > 0 {
		reqId := r.Header.Get(p.ReqIdName)
		if len(reqId) > 0 {
			out, err, shared = rr.flightGroup.Do(reqId, action)
			return rr.LogResult(reqId, out, err, shared)
		}
	}
	groupKey := rr.Digest(r)
	out, err, shared = rr.flightGroup.Do(groupKey, action)
	return rr.LogResult(groupKey, out, err, shared)
}

func (rr *ReqRestrictor) LogResult(groupKey string, state interface{}, err error, shared bool) (interface{}, error, bool) {
	log.Printf("request [%s] has duplicated by key: %t", groupKey, shared)
	return state, err, shared
}

func (rr *ReqRestrictor) Digest(r *http.Request) string {
	o := []string{}
	p := rr.flightPattern
	if p.HasMethod {
		o = append(o, r.Method)
	}
	if p.HasPath {
		o = append(o, r.URL.Path)
	}

	h := sha256.New()
	if p.HasHeaders != nil {
		headers := r.Header
		for _, key := range p.HasHeaders {
			val := headers.Get(key)
			if len(val) > 0 {
				h.Write([]byte(val))
			}
		}
	}
	if p.HasQueries != nil {
		queries := r.URL.Query()
		for _, key := range p.HasQueries {
			val := queries.Get(key)
			if len(val) > 0 {
				h.Write([]byte(val))
			}
		}
	}
	s := h.Sum(nil)
	o = append(o, fmt.Sprintf("%x", s))

	if p.HasUserIP {
		userip, err := extractUserIP(r)
		if err == nil && userip != nil {
			o = append(o, userip.String())
		}
	}

	return strings.Join(o, "|")
}

// extractUserIP() extracts the user IP address from req, if present.
func extractUserIP(req *http.Request) (net.IP, error) {
	ip, _, err := net.SplitHostPort(req.RemoteAddr)
	if err != nil {
		return nil, fmt.Errorf("userip: %q is not IP:port", req.RemoteAddr)
	}
	userIP := net.ParseIP(ip)
	if userIP == nil {
		return nil, fmt.Errorf("userip: %q is not IP:port", req.RemoteAddr)
	}
	return userIP, nil
}
