# UReference

```high-speed```
```thread-safe```
```key-value```
```all data in memory```
```not-persistent```
```auto recycling ```

## usage

```go
//import
import (
    "github.com/coreservice-io/UReference"
)
```

### example

```go
package main

import (
	"github.com/coreservice-io/UReference"
	"log"
)

type Person struct {
	Name     string
	Age      int
	Location string
}

func main() {
	 
	lc := UReference.New()  //new a UReference instance with default config
	lc.SetMaxRecords(10000)

	//set
	lc.Set("foo", "bar", 300)
	lc.Set("a", 1, 300)
	lc.Set("b", Person{"Jack", 18, "London"}, 300)
	lc.Set("b*", &Person{"Jack", 18, "London"}, 300)
	lc.Set("c", true, 100)

	//get
	value, ttlLeft, exist := lc.Get("foo")
	if exist {
		//value type is interface{}, please convert to the right type before usage
		valueStr, ok := value.(string)
		if ok {
			log.Println("key:foo, value:", valueStr)
		}
		log.Println("key:foo, ttl:", ttlLeft)
	}

	//get
	log.Println("---get---")
	log.Println(lc.Get("foo"))
	log.Println(lc.Get("a"))
	log.Println(lc.Get("b"))
	log.Println(lc.Get("b*"))
	log.Println(lc.Get("c"))

	//overwrite
	log.Println("---set overwrite---")
	log.Println(lc.Get("c"))
	lc.Set("c", false, 60)
	log.Println(lc.Get("c"))
}
```

### default config

```
MaxRecords(*)         = 1000000
MinRecords            = 10000
MaxTTLSecs            = 7200
RecycleIntervalSecs   = 5
RecycleOverLimitRatio = 0.15
(* : configurable)
```

### auto recycling

RecycleOverLimitRatio of records will be recycled automatically
if MaxRecords is reached.

### custom config

```go
//new instance
lc,err := UReference.New()
if err != nil {
    panic(err.Error())
}
lc.SetMaxRecords(10000) //custom the max key-value pairs that can be kept in memory
```

## Benchmark

### set

```
cpu: Intel(R) Core(TM) i7-7700HQ CPU @ 2.80GHz
BenchmarkLocalCache_SetPointer
BenchmarkLocalCache_SetPointer-8   	 1000000	      1618 ns/op	     379 B/op	      10 allocs/op
PASS
```

### get

```
cpu: Intel(R) Core(TM) i7-7700HQ CPU @ 2.80GHz
BenchmarkLocalCache_GetPointer
BenchmarkLocalCache_GetPointer-8   	 9931429	       129.7 ns/op	       0 B/op	       0 allocs/op
PASS
```
