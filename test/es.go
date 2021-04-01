package main

import (
	"fmt"
	"rule_engine_by_go/app"
)

type Employee struct {
	FirstName string   `json:"first_name"`
	LastName  string   `json:"last_name"`
	Age       int      `json:"age"`
	About     string   `json:"about"`
	Interests []string `json:"interests"`
}

func main() {

	e1 := Employee{"Jane", "Smith", 32, "I like to collect rock albums", []string{"music"}}
	dataSlice := []Employee{e1}
	interfaceSlice := make([]interface{}, len(dataSlice))
	for i, d := range dataSlice {
		interfaceSlice[i] = d
	}
	fmt.Println(interfaceSlice)
	for {
		app.SendEs("nids", "2", interfaceSlice)
	}

}
