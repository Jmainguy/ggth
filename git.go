package main

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"

	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/google/go-github/v39/github"
)

func pushBranch(ctx context.Context, currentRepo *git.Repository, newBranchName, githubUsername, githubPassword string) {
	// Create Worktree
	wt, err := currentRepo.Worktree()
	if err != nil {
		panic(err)
	}
	// Check worktree status
	wtStatus, err := wt.Status()
	if err != nil {
		panic(err)
	}
	// If not clean, new branch, add, commit, push
	if !wtStatus.IsClean() {
		// Checkout new branch
		err := wt.Checkout(&git.CheckoutOptions{
			Create: true,
			Keep:   true,
			Force:  false,
			Branch: plumbing.NewBranchReferenceName(newBranchName),
		})
		if err != nil {
			panic(err)
		}
		// git add .
		_, err = wt.Add(".")
		if err != nil {
			panic(err)
		}
		// git commit -m
		_, err = wt.Commit("GGTH template updating", &git.CommitOptions{})
		if err != nil {
			panic(err)
		}
		// git push origin
		err = currentRepo.PushContext(ctx, &git.PushOptions{
			RemoteName: "origin",
			Auth: &http.BasicAuth{
				Username: githubUsername,
				Password: githubPassword,
			},
		})
		if err != nil {
			panic(err)
		}

	}

}

func getNewBranchName(branches []*github.Branch) string {
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
	newBranchName := fmt.Sprintf("ggth-%d", pullRequestIncrement)
	return newBranchName
}

func cloneRepo(templateRepo *github.Repository, repoDir, githubUsername, githubPassword string) {
	if templateRepo != nil {
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
			Auth: &http.BasicAuth{
				Username: githubUsername,
				Password: githubPassword,
			},
		})
		if err != nil {
			panic(err)
		}

	} else {
		fmt.Println("Templated repo could not be found")
		os.Exit(1)
	}
}

func getRepoNameFromGitConfig(currentRepo *git.Repository) (owner, name string) {
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
	ownerRepo := strings.Split(repoName, "/")
	owner = strings.ToLower(ownerRepo[0])
	name = strings.ToLower(ownerRepo[1])

	return owner, name
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
