package weblimiter

import (
	"encoding/json"
	"reflect"
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"
)
type WebLimiter struct {
	mux *sync.RWMutex
	//rateConfMap stores the value of limiting rules
	rateConfs []RateConf
	//configCenter stores the value of limiting rules,
	//it could be etcd,zk...
	configCenter ConfigCenter
	//key defines the key in configCentor,
	// and the return value is limiting rules
	key      string
	interval time.Duration
	limiters []*TokenRateLimiter
	confsChan chan []RateConf
	notifyLimiterMsg   func(msg string) error
}

func NewWebLimiter(config ConfigCenter, key string,interval time.Duration) *WebLimiter {
	mux := new(sync.RWMutex)
	return &WebLimiter{
		mux:          mux,
		key:          key,
		configCenter: config,
		interval:	interval,
	}
}

func (limiter *WebLimiter) Init() {
	confs,err:=limiter.getConfsFromConfigCenter()
	if err!=nil{
		log.Errorf("get confs from config center error:%s",err)
	}
	pconfs,err:=limiter.parseConfs(confs)
	if err!=nil{
		log.Errorf("parse confs error:%s")
	}
	limiters:=limiter.makeLimiters(pconfs)
	limiter.limiters = limiters
	limiter.rateConfs = pconfs

	go limiter.Daemon()
}

func (limiter *WebLimiter) Daemon(){
	ticker := time.NewTicker(limiter.interval)
	defer ticker.Stop()
	for {
		select {
		case <- ticker.C:
			newConfs,err:= limiter.getAndParseConfs()
			if err!=nil{
				continue
			}
			if !limiter.compareConfs(newConfs){
				limiter.rateConfs=newConfs
				limiters := limiter.makeLimiters(newConfs)
				limiter.mux.Lock()
				limiter.limiters = limiters
				limiter.mux.Unlock()
			}
		}
	}
}

func (limiter *WebLimiter) getAndParseConfs() ([]RateConf,error){
	confs,err:=limiter.getConfsFromConfigCenter()
	if err!=nil{
		return nil,err
	}
	return limiter.parseConfs(confs)
}

func (limiter *WebLimiter) compareConfs(newConfs []RateConf) bool {
	if !compareRateConf(limiter.rateConfs,newConfs) {
		return false
	}

	return true
}

func (limiter *WebLimiter) getConfsFromConfigCenter() (map[string]string, error) {
	config := limiter.configCenter
	key := limiter.key
	confs, err := config.GetConfig(key)
	if err!=nil{
		return nil,err
	}
	dump:=make(map[string]string)
	if err:= json.Unmarshal([]byte(confs),&dump);err!=nil{
		return nil,err
	}
	return dump,nil
}

func (limiter *WebLimiter) parseConfs(confs map[string]string) ([]RateConf,error){
	rateConfs := []RateConf{}
	return rateConfs,nil
}

func (limiter *WebLimiter) makeLimiters(rateConfs []RateConf) []*TokenRateLimiter{
	rateLimiter :=NewTokenRateLimiter(rateConfs)
	return []*TokenRateLimiter{rateLimiter}
}


func compareRateConf(old,new []RateConf) bool{
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
