package main

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/google/go-github/v39/github"
	"golang.org/x/oauth2"
)

func hasGitConfig() bool {
	_, err := os.Stat(".git/config")
	return !os.IsNotExist(err)
}

func getRepoURLFromGitConfig() string {
	gitConfigExists := hasGitConfig()
	if !gitConfigExists {
		fmt.Println(".git/config doesn't exist")
		os.Exit(1)
	}
	file, err := os.Open(".git/config")
	if err != nil {
		panic(err)
	}
	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)
	var line string
	for scanner.Scan() {
		if strings.Contains(scanner.Text(), "url = ") {
			urlLine := strings.Split(scanner.Text(), "url = ")
			// SSH style
			line = strings.ReplaceAll(urlLine[1], "git@github.com:", "")
			line = strings.ReplaceAll(line, ".git", "")
			// HTTPS style
			line = strings.ReplaceAll(line, "https://github.com/", "")
		}
	}
	return line
}

func templateFile(filename string, repo *github.Repository, gitName, gitEmail string) {
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

	err = ioutil.WriteFile(filename, []byte(newContents), 0664)
	if err != nil {
		panic(err)
	}
}

// https://golang.cafe/blog/how-to-list-files-in-a-directory-in-go.html
func getAllFiles() (files []string) {
	err := filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			fmt.Println(err)
			return err
		}

		if !info.IsDir() {
			// Don't edit .git/ files
			if !strings.HasPrefix(path, ".git/") {
				files = append(files, path)
			}
		}
		return nil
	})
	if err != nil {
		fmt.Println(err)
	}
	return files
}

func getGitProfile() (gitName, gitEmail string) {
	cmd := exec.Command("git", "config", "user.Name")
	var stdoutName bytes.Buffer
	var stdoutEmail bytes.Buffer
	cmd.Stdout = &stdoutName
	err := cmd.Run()
	if err != nil {
		log.Fatalf("cmd.Run() failed with %s\n", err)
	}
	gitName = stdoutName.String()
	gitName = strings.TrimSuffix(gitName, "\n")

	cmd = exec.Command("git", "config", "user.Email")
	cmd.Stdout = &stdoutEmail
	err = cmd.Run()
	if err != nil {
		log.Fatalf("cmd.Run() failed with %s\n", err)
	}

	gitEmail = stdoutEmail.String()
	gitEmail = strings.TrimSuffix(gitEmail, "\n")
	return gitName, gitEmail

}
func main() {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: os.Getenv("ghToken")},
	)
	tc := oauth2.NewClient(ctx, ts)

	client := github.NewClient(tc)

	gitName, gitEmail := getGitProfile()
	repoURL := getRepoURLFromGitConfig()
	ownerRepo := strings.Split(repoURL, "/")
	owner := strings.ToLower(ownerRepo[0])
	repoName := strings.ToLower(ownerRepo[1])
	repo, _, err := client.Repositories.Get(ctx, owner, repoName)
	if err != nil {
		panic(err)
	}

	files := getAllFiles()
	for _, file := range files {
		templateFile(file, repo, gitName, gitEmail)
	}
}
