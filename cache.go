package UReference

import (
	"errors"
	"reflect"
	"time"

	"github.com/coreservice-io/UReference/sortedset"
)

const (
	MaxRecords            = 1000000
	MinRecords            = 10000
	MaxTTLSecs            = 7200
	RecycleIntervalSecs   = 5
	RecycleOverLimitRatio = 0.5
)

type Cache struct {
	s     *sortedset.SortedSet
	limit int64
}

func New() *Cache {
	cache := &Cache{
		s:     sortedset.Make(),
		limit: MaxRecords,
	}

	cache.Recycle()
	return cache
}

//RecycleOverLimitRatio of records will be recycled if the number of total keys exceeds this limit
func (lc *Cache) SetMaxRecords(limit int64) {
	if limit < MinRecords {
		limit = MinRecords
	}
	lc.limit = limit
}

func (lc *Cache) Get(key string) (value interface{}, ttl int64, exist bool) {
	//check expire
	e, exist := lc.s.Get(key)
	if !exist {
		return nil, 0, false
	}
	nowTime := time.Now().Unix()
	if e.Score <= nowTime {
		return nil, 0, false
	}
	return e.Value, e.Score - nowTime, true
}

//if ttl < 0 just return and nothing changes
//ttl is set to MaxTTLSecs if ttl > MaxTTLSecs
//if record exist , "0" ttl changes nothing
//if record not exist, "0" ttl is equal to "30" seconds
func (lc *Cache) Set(key string, value interface{}, ttlSecond int64) error {
	if value == nil {
		return errors.New("value can not be nil")
	}
	if ttlSecond < 0 {
		return errors.New("ttl error")
	}

	t := reflect.TypeOf(value).Kind()
	if t != reflect.Ptr && t != reflect.Slice && t != reflect.Map {
		return errors.New("value only support Pointer Slice and Map")
	}

	if ttlSecond > MaxTTLSecs {
		ttlSecond = MaxTTLSecs
	}
	var expireTime int64

	if ttlSecond == 0 {
		//keep
		ttlLeft, exist := lc.ttl(key)
		if !exist {
			ttlLeft = 30
		}
		expireTime = time.Now().Unix() + ttlLeft
	} else {
		//new expire
		expireTime = time.Now().Unix() + ttlSecond
	}
	lc.s.Add(key, expireTime, value)
	return nil
}

func (lc *Cache) Delete(key string) {
	lc.s.Remove(key)
}

// get ttl of a key in seconds
func (lc *Cache) ttl(key string) (int64, bool) {
	e, exist := lc.s.Get(key)
	if !exist {
		return 0, false
	}
	ttl := e.Score - time.Now().Unix()
	if ttl <= 0 {
		return 0, false
	}
	return ttl, true
}

func (lc *Cache) Recycle() {
	time.Sleep(500 * time.Millisecond)
	safeInfiLoop(func() {
		//remove expired keys
		lc.s.RemoveByScore(time.Now().Unix())
		//check overlimit
		if lc.s.Len() >= lc.limit {
			deleteCount := (lc.s.Len() - lc.limit) + int64(float64(lc.limit)*RecycleOverLimitRatio)
			lc.s.RemoveByRank(0, int64(deleteCount))
		}
	}, nil, RecycleIntervalSecs, 60)
}

func (lc *Cache) GetLen() int64 {
	return lc.s.Len()
}

func (lc *Cache) SetRand(key string, ttlSecond int64) string {
	rs := GenRandStr(20)
	lc.Set(key, rs, ttlSecond)
	return rs
}

func (lc *Cache) GetRand(key string) string {
	v, _, exist := lc.Get(key)
	if !exist {
		return ""
	}
	return v.(string)
}

func safeInfiLoop(todo func(), onPanic func(err interface{}), interval int, redoDelaySec int) {
	runChannel := make(chan struct{})
	go func() {
		for {
			<-runChannel
			go func() {
				defer func() {
					if err := recover(); err != nil {
						if onPanic != nil {
							onPanic(err)
						}
						time.Sleep(time.Duration(redoDelaySec) * time.Second)
						runChannel <- struct{}{}
					}
				}()
				for {
					todo()
					time.Sleep(time.Duration(interval) * time.Second)
				}
			}()
		}
	}()
	runChannel <- struct{}{}
}
