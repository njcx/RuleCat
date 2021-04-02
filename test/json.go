package main

import "rulecat/utils/json"

const json_ = `{"name":{"first":"Janet","name":"Prichard"},"age":47}`

func main() {
	value := json.Get(json_, "name.last\\.name")
	println(value.String())
}
