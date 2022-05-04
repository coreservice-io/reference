package main

import (
	"log"
	"time"

	"github.com/coreservice-io/reference"
)

type Person struct {
	Name     string
	Age      int32
	Location string
}

func main() {

	lf := reference.New()
	lf.SetMaxRecords(10000)

	//set ""
	v := "nothing value"
	err := lf.Set("", &v, 300) //only support Pointer Slice and Map
	if err != nil {
		log.Fatalln("reference set error:", err)
	}
	//get ""
	valuen, ttl := lf.Get("")
	if valuen != nil {
		log.Println("key:nothing value:", valuen.(*string), "ttl:", ttl)
	}

	//set slice
	err = lf.Set("slice", []int{1, 2, 3}, 300)
	if err != nil {
		log.Fatalln("reference set error:", err)
	}

	//set struct pointer
	err = lf.Set("struct*", &Person{"Jack", 18, "London"}, 300)
	if err != nil {
		log.Fatalln("reference set error:", err)
	}

	//set map
	err = lf.Set("map", map[string]int{"a": 1, "b": 2}, 100)
	if err != nil {
		log.Fatalln("reference set error:", err)
	}

	//get
	log.Println("---get---")
	log.Println(lf.Get("slice"))
	log.Println(lf.Get("struct*"))
	log.Println(lf.Get("map"))

	//overwrite
	log.Println("---set overwrite---")
	log.Println(lf.Get("struct*"))
	err = lf.Set("struct*", &Person{"Tom", 38, "London"}, 10)
	if err != nil {
		log.Fatalln("reference set error:", err)
	}
	log.Println(lf.Get("struct*"))

	//test ttl
	go func() {
		for {
			time.Sleep(2 * time.Second)
			log.Println(lf.Get("struct*"))
		}
	}()

	time.Sleep(20 * time.Second)

	//if not a pointer cause error
	err = lf.Set("int", 10, 10)
	if err != nil {
		log.Fatalln("reference set error:", err)
	}
}
