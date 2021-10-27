package main

import (
	"io/ioutil"
	"strings"

	"github.com/google/go-github/v39/github"
)

func templateFile(filename string, repo *github.Repository, gitName, gitEmail string) {
	usage := getUsage()
	read, err := ioutil.ReadFile(filename)
	if err != nil {
		panic(err)
	}
	newContents := strings.Replace(string(read), "--REPOLINK--", *repo.HTMLURL, -1)
	newContents = strings.Replace(newContents, "--DESCRIPTION--", *repo.Description, -1)
	newContents = strings.Replace(newContents, "--BINARY--", *repo.Name, -1)
	newContents = strings.Replace(newContents, "--REPONAME--", *repo.Name, -1)
	newContents = strings.Replace(newContents, "--REPOOWNER--", *repo.Owner.Login, -1)

	newContents = strings.Replace(newContents, "--MAINTAINEREMAIL--", gitEmail, -1)
	newContents = strings.Replace(newContents, "--MAINTAINERNAME--", gitName, -1)
	newContents = strings.Replace(newContents, "--USAGE--", usage, -1)

	err = ioutil.WriteFile(filename, []byte(newContents), 0664)
	if err != nil {
		panic(err)
	}
}
