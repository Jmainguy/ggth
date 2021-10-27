package main

import (
	"io/ioutil"
)

func getUsage() string {
	usage, err := ioutil.ReadFile(".usage")
	if err != nil {
		return "--USAGE--"
	}
	return string(usage)
}
