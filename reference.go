package reference

import (
	"errors"
	"reflect"
	"time"

	"github.com/coreservice-io/reference/sortedset"
)

const (
	MaxRecords            = 5000000
	MinRecords            = 10000
	MaxTTLSecs            = 7200
	RecycleIntervalSecs   = 5
	RecycleOverLimitRatio = 0.5
)

type Reference struct {
	s           *sortedset.SortedSet
	limit       int64
	now_unixtime int64
}

func New() *Reference {
	ref := &Reference{
		s:           sortedset.Make(),
		limit:       MaxRecords,
		now_unixtime: time.Now().Unix(),
	}

	ref.Recycle()
	go func() {
		for {
			time.Sleep(1 * time.Second)
			ref.now_unixtime = time.Now().Unix()
		}
	}()

	return ref
}

// RecycleOverLimitRatio of records will be recycled if the number of total keys exceeds this limit
func (lf *Reference) SetMaxRecords(limit int64) {
	if limit < MinRecords {
		limit = MinRecords
	}
	lf.limit = limit
}

// get current unix time in reference
func (lf *Reference) GetUnixTime() int64 {
	return lf.now_unixtime
}

// if not found or timeout => return nil,0
// if found and not timeout =>return not_nil_pointer,left_secs
func (lf *Reference) Get(key string) (value interface{}, ttl int64) {
	//check expire
	e, exist := lf.s.Get(key)
	if !exist {
		return nil, 0
	}
	if e.Score <= lf.now_unixtime {
		return nil, 0
	}
	return e.Value, 1
}

// if ttl < 0 just return and nothing changes
// ttl is set to MaxTTLSecs if ttl > MaxTTLSecs
// if record exist , "0" ttl changes nothing
// if record not exist, "0" ttl is equal to "30" seconds
func (lf *Reference) Set(key string, value interface{}, ttlSecond int64) error {
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
		ttlLeft, exist := lf.ttl(key)
		if !exist {
			ttlLeft = 30
		}
		expireTime = lf.now_unixtime + ttlLeft
	} else {
		//new expire
		expireTime = lf.now_unixtime + ttlSecond
	}
	lf.s.Add(key, expireTime, value)
	return nil
}

func (lf *Reference) Delete(key string) {
	lf.s.Remove(key)
}

// get ttl of a key in seconds
func (lf *Reference) ttl(key string) (int64, bool) {
	e, exist := lf.s.Get(key)
	if !exist {
		return 0, false
	}
	ttl := e.Score - lf.now_unixtime
	if ttl <= 0 {
		return 0, false
	}
	return ttl, true
}

func (lf *Reference) Recycle() {
	time.Sleep(500 * time.Millisecond)
	safeInfiLoop(func() {
		//remove expired keys
		lf.s.RemoveByScore(lf.now_unixtime)
		//check overlimit
		if lf.s.Len() >= lf.limit {
			deleteCount := (lf.s.Len() - lf.limit) + int64(float64(lf.limit)*RecycleOverLimitRatio)
			lf.s.RemoveByRank(0, int64(deleteCount))
		}
	}, nil, RecycleIntervalSecs, 60)
}

func (lf *Reference) GetLen() int64 {
	return int64(lf.s.Len())
}

func safeInfiLoop(todo func(), onPanic func(err interface{}), interval int64, redoDelaySec int64) {
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
