package reference

import (
	"errors"
	"reflect"
	"time"

	"github.com/coreservice-io/reference/sortedset"
)

const (
	MaxRecords            = 1000000
	MinRecords            = 10000
	MaxTTLSecs            = 7200
	RecycleIntervalSecs   = 5
	RecycleOverLimitRatio = 0.5
)

type Reference struct {
	s     *sortedset.SortedSet
	limit int64
}

func New() *Reference {
	ref := &Reference{
		s:     sortedset.Make(),
		limit: MaxRecords,
	}

	ref.Recycle()
	return ref
}

//RecycleOverLimitRatio of records will be recycled if the number of total keys exceeds this limit
func (lf *Reference) SetMaxRecords(limit int64) {
	if limit < MinRecords {
		limit = MinRecords
	}
	lf.limit = limit
}

//if not found or timeout => return nil,0
//if found and not timeout =>return not_nil_pointer,left_secs
func (lf *Reference) Get(key string) (value interface{}, ttl int64) {
	//check expire
	e, exist := lf.s.Get(key)
	if !exist {
		return nil, 0
	}
	nowTime := time.Now().Unix()
	if e.Score <= nowTime {
		return nil, 0
	}
	return e.Value, e.Score - nowTime
}

//if ttl < 0 just return and nothing changes
//ttl is set to MaxTTLSecs if ttl > MaxTTLSecs
//if record exist , "0" ttl changes nothing
//if record not exist, "0" ttl is equal to "30" seconds
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
		expireTime = time.Now().Unix() + ttlLeft
	} else {
		//new expire
		expireTime = time.Now().Unix() + ttlSecond
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
	ttl := e.Score - time.Now().Unix()
	if ttl <= 0 {
		return 0, false
	}
	return ttl, true
}

func (lf *Reference) Recycle() {
	time.Sleep(500 * time.Millisecond)
	safeInfiLoop(func() {
		//remove expired keys
		lf.s.RemoveByScore(time.Now().Unix())
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

func (lf *Reference) SetRand(key string, ttlSecond int64) string {
	rs := GenRandStr(20)
	lf.Set(key, rs, ttlSecond)
	return rs
}

func (lf *Reference) GetRand(key string) string {
	v, _ := lf.Get(key)
	if v == nil {
		return ""
	}
	return v.(string)
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
