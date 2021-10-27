package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/google/go-github/v39/github"
	"golang.org/x/oauth2"
)

func getRepoNameFromGitConfig(currentRepo *git.Repository) string {
	currentRepoConfig, err := currentRepo.Config()
	var originURL string
	if err != nil {
		panic(err)
	}
	for _, remote := range currentRepoConfig.Remotes {
		if remote.Name == "origin" {
			for _, URL := range remote.URLs {
				originURL = URL
			}
		}
	}

	// SSH style
	repoName := strings.ReplaceAll(originURL, "git@github.com:", "")
	repoName = strings.ReplaceAll(repoName, ".git", "")
	// HTTPS style
	repoName = strings.ReplaceAll(repoName, "https://github.com/", "")
	return repoName
}

func getUsage() string {
	usage, err := ioutil.ReadFile(".usage")
	if err != nil {
		panic(err)
	}
	return string(usage)
}

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

// https://golang.cafe/blog/how-to-list-files-in-a-directory-in-go.html
func getAllFiles(dirPath string) (files []string) {
	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			fmt.Println(err)
			return err
		}

		if !info.IsDir() {
			// Don't list .git/ files
			if !strings.Contains(path, ".git/") && !strings.Contains(path, ".template/") {
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
	gitConfig, err := config.LoadConfig(config.GlobalScope)
	if err != nil {
		panic(err)
	}
	gitName = gitConfig.User.Name
	gitEmail = gitConfig.User.Email
	return gitName, gitEmail
}

func copyFiles(src, dst string) {
	srcFiles := getAllFiles(src)
	for _, file := range srcFiles {
		destFile := strings.TrimPrefix(file, src)
		destFile = dst + "/" + destFile
		bytesRead, err := ioutil.ReadFile(file)
		if err != nil {
			log.Fatal(err)
		}
		err = ioutil.WriteFile(destFile, bytesRead, 0644)

		if err != nil {
			log.Fatal(err)
		}

	}
}

func main() {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: os.Getenv("ghToken")},
	)
	tc := oauth2.NewClient(ctx, ts)

	client := github.NewClient(tc)

	gitName, gitEmail := getGitProfile()
	currentRepo, err := git.PlainOpen(".")
	if err != nil {
		panic(err)
	}
	repoURL := getRepoNameFromGitConfig(currentRepo)
	ownerRepo := strings.Split(repoURL, "/")
	owner := strings.ToLower(ownerRepo[0])
	repoName := strings.ToLower(ownerRepo[1])
	repo, _, err := client.Repositories.Get(ctx, owner, repoName)
	if err != nil {
		panic(err)
	}
	repoDir := "/tmp/foo/"
	templateRepo := repo.GetTemplateRepository()
	if templateRepo != nil {
		fmt.Println(*templateRepo.Name)
		fmt.Println(*templateRepo.CloneURL)

		if _, err := os.Stat(repoDir); !os.IsNotExist(err) {
			// Delete it, and clone it fresh
			err = os.RemoveAll(repoDir)
			if err != nil {
				panic(err)
			}
		}
		_, err := git.PlainClone(repoDir, false, &git.CloneOptions{
			URL:      *templateRepo.CloneURL,
			Progress: nil,
		})
		if err != nil {
			panic(err)
		}

	}
	copyFiles(repoDir, ".")

	files := getAllFiles(".")
	for _, file := range files {
		templateFile(file, repo, gitName, gitEmail)
	}
	// What branch we on
	branches, _, err := client.Repositories.ListBranches(ctx, owner, repoName, &github.BranchListOptions{})
	if err != nil {
		panic(err)
	}

	var pullRequestIncrement int
	for _, branch := range branches {
		if strings.HasPrefix(*branch.Name, "ggth-") {
			incrementStr := strings.TrimPrefix(*branch.Name, "ggth-")
			increment, err := strconv.Atoi(incrementStr)
			if err != nil {
				panic(err)
			}

			if increment >= pullRequestIncrement {
				pullRequestIncrement = increment + 1
			}

		}
	}
	// Check worktree status
	wt, err := currentRepo.Worktree()
	if err != nil {
		panic(err)
	}
	wtStatus, err := wt.Status()
	if err != nil {
		panic(err)
	}

	// If not clean, checkout, add, commit, push
	if !wtStatus.IsClean() {
		newBranchName := fmt.Sprintf("ggth-%d", pullRequestIncrement)
		err = wt.Checkout(&git.CheckoutOptions{
			Create: true,
			Keep:   true,
			Force:  false,
			Branch: plumbing.NewBranchReferenceName(newBranchName),
		})
		if err != nil {
			panic(err)
		}

		_, err = wt.Add(".")
		if err != nil {
			panic(err)
		}

		_, err = wt.Commit("GGTH template updating", &git.CommitOptions{})
		if err != nil {
			panic(err)
		}

		err = currentRepo.PushContext(ctx, &git.PushOptions{
			RemoteName: "origin",
		})
		if err != nil {
			panic(err)
		}

	}
}
