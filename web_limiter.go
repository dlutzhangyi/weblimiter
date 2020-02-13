package weblimiter

import (
	"reflect"
	"sync"

	log "github.com/sirupsen/logrus"
)

type WebLimiter struct {
	mux *sync.RWMutex
	//rateConfMap stores the value of limiting rules
	rateConfig []RateConf
	//client is used to get limiting rules from config center,like zk,etcd,redis..
	client ConfigClient
	//key defines the key in configCentor,
	// and the return value is limiting rules
	key string
	//limiters for rate limit
	limiters []*TokenRateLimiter
	//rateConfChan stores the value get from client, once the client the new rateconfig,
	//it put the config into rateConfChan
	rateConfChan chan []RateConf
	//notifyMsg is used to notify the msg for the change of rateConf
	notifyMsg func(msg string) error
}

func NewWebLimiter(client ConfigClient, key string) *WebLimiter {
	mux := new(sync.RWMutex)
	return &WebLimiter{
		mux:    mux,
		key:    key,
		client: client,
	}
}

func (limiter *WebLimiter) Init() {
	rateConfig, err := limiter.getAndParseConfig()
	if err != nil {
		log.Fatalf("get and parse config err:%s", err)
	}

	limiters := limiter.makeLimiters(rateConfig)
	limiter.limiters = limiters
	limiter.rateConfig = rateConfig
	limiter.client.RegisterConfigChannel(limiter.rateConfChan)

	go limiter.Daemon()
}

func (limiter *WebLimiter) getAndParseConfig() ([]RateConf, error) {
	config, err := limiter.client.GetConfig(limiter.key)
	if err != nil {
		return nil, err
	}
	return limiter.client.ParseConfig(config)
}

func (limiter *WebLimiter) Daemon() {
	for {
		select {
		case config := <-limiter.rateConfChan:
			if !limiter.compareConfig(config) {
				limiters := limiter.makeLimiters(config)
				limiter.mux.Lock()
				limiter.rateConfig = config
				limiter.limiters = limiters
				limiter.mux.Unlock()
			}
		}
	}
}

// compare the new config with the last config,if same,then return
func (limiter *WebLimiter) compareConfig(newConfs []RateConf) bool {
	return compareRateConf(limiter.rateConfs, newConfs)
}

func (limiter *WebLimiter) makeLimiters(rateConfs []RateConf) []*TokenRateLimiter {
	rateLimiter := NewTokenRateLimiter(rateConfs)
	return []*TokenRateLimiter{rateLimiter}
}

func compareRateConf(old, new []RateConf) bool {
	oldMap := make(map[string]float64)
	newMap := make(map[string]float64)
	for _, conf := range old {
		oldMap[conf.Request] = conf.Rate
	}
	for _, conf := range new {
		newMap[conf.Request] = conf.Rate
	}
	return reflect.DeepEqual(oldMap, newMap)
}
