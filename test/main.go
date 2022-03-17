package main

import (
	"log"
	"time"

	"github.com/coreservice-io/UReference"
)

type Person struct {
	Name     string
	Age      int
	Location string
}

func main() {

	lc := UReference.New()
	lc.SetMaxRecords(10000)

	//set ""
	lc.Set("", "nothing value", 300)
	//get ""
	valuen, _, okn := lc.Get("")
	if okn {
		log.Println("key:nothing value:", valuen.(string))
	}

	//set
	lc.Set("foo", "bar", 300)
	lc.Set("a", 1, 300)
	lc.Set("b", Person{"Jack", 18, "London"}, 300)
	lc.Set("b*", &Person{"Jack", 18, "London"}, 300)
	lc.Set("c", true, 100)

	//get
	value, ttlLeft, exist := lc.Get("foo")
	if exist {
		valueStr, ok := value.(string) //value type is interface{},convert to the right type before usage
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
	lc.Set("c", false, 10)
	log.Println(lc.Get("c"))

	go func() {
		for {
			time.Sleep(2 * time.Second)
			log.Println(lc.Get("c"))
		}
	}()

	time.Sleep(30 * time.Second)
}
