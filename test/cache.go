package main

import (
	"log"
	"rule_engine_by_go/utils/cache"
	"time"
)

func main() {

	tc := cache.New(60*time.Second, 5*time.Second)

	tc.Set("a", 1, cache.DefaultExpiration)
	tc.Set("b", "b", cache.DefaultExpiration)
	tc.Set("c", 3.5, cache.DefaultExpiration)

	tc.Set("c", 3, 10*time.Second)
	tc.Set("d", 4, 10*time.Second)

	tc.Increment("d", 2)

	value, found := tc.Get("b")
	if found {
		log.Println("found:", value)
	} else {
		log.Println("not found")
	}

	value, found = tc.Get("c")
	if found {
		log.Println("found:", value)
	} else {
		log.Println("not found")
	}

	time.Sleep(60 * time.Second)
	log.Println("sleep 60s...")
	value, found = tc.Get("d")
	if found {
		log.Println("found:", value)
	} else {
		log.Println("not found")
	}

}
