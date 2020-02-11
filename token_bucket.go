package weblimiter

import (
	"time"

	"sync"

	"golang.org/x/time/rate"
)

/*
	a RateLimiter based on token bucket
*/

const (
	defaultBucketSize = 10
	defaultToken      = 1
)

type RateConf struct {
	//Request defines the input request
	Request string
	//Rate defines the rate speed in one minutes, for example,if Rate = 60, it means only 60 request can be passed in one minutes
	Rate float64
}

type httpRequestConf struct {
	//request defines the input request
	request string
	//rateSpeed defines the rate speed in one seconds
	rateSpeed float64
	//bucketSize defines the buckt size for token bucket
	bucketSize int
	//token defines how much token put into the bucket in one second
	token int
}

type RequestLimiter struct {
	Limiter *rate.Limiter
}

type TokenRateLimiter struct {
	//controller stores the relation between request path and the corresponding limiter
	controller map[string]*RequestLimiter
	mux        *sync.RWMutex
	//rateConfMap stores the relation between request path the the corresponding limiter conf
	rateConfMap map[string]*httpRequestConf
}

// NewTokenRateLimiter creates a ratelimiter controller
func NewTokenRateLimiter(confs []RateConf) *TokenRateLimiter {
	mux := &sync.RWMutex{}
	limiter := &TokenRateLimiter{
		controller:  make(map[string]*RequestLimiter),
		mux:         mux,
		rateConfMap: make(map[string]*httpRequestConf),
	}
	limiter.setRateConf(confs)
	return limiter
}

// setRateConf is used to parse rate conf and store the value in rateConfMap
func (rl *TokenRateLimiter) setRateConf(confs []RateConf) {
	for _, v := range confs {
		conf := httpRequestConf{
			request:    v.Request,
			rateSpeed:  float64(v.Rate / 60.0),
			bucketSize: defaultBucketSize,
			token:      defaultToken,
		}
		rl.rateConfMap[v.Request] = &conf
	}

}

//addLimiter is used to add the corresponding request limiter.
func (rl *TokenRateLimiter) addLimiter(request string) *rate.Limiter {
	rateSpeed := rl.rateConfMap[request].rateSpeed
	bucketSize := rl.rateConfMap[request].bucketSize

	limiter := rate.NewLimiter(rate.Limit(rateSpeed), bucketSize)
	rl.mux.Lock()
	defer rl.mux.Unlock()
	rl.controller[request] = &RequestLimiter{
		Limiter: limiter,
	}
	return limiter
}

//getLimiter is used to get the limiter for the request,
//if the corresponding limiter is not exists,it will call the func addLimiter to generate a limiter
func (rl *TokenRateLimiter) getLimiter(request string) *rate.Limiter {
	rl.mux.RLock()
	v, exists := rl.controller[request]

	if !exists {
		rl.mux.RUnlock()
		return rl.addLimiter(request)
	}
	rl.mux.RUnlock()
	return v.Limiter
}

//Limited is used to determine whether to limit the request
func (rl *TokenRateLimiter) Limited(request string, options map[string]string) bool {
	if _, ok := rl.rateConfMap[request]; !ok {
		return false
	}

	limiter := rl.getLimiter(request)
	token := rl.rateConfMap[request].token

	if limiter.AllowN(time.Now(), token) == false {
		return true
	}
	return false
}
