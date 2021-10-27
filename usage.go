package main

import (
	"io/ioutil"
)

func getUsage() string {
	usage, err := ioutil.ReadFile(".usage")
	if err != nil {
		panic(err)
	}
	return string(usage)
}
