package main

import (
	"context"
	"os"

	git "github.com/go-git/go-git/v5"
	"github.com/google/go-github/v39/github"
	"golang.org/x/oauth2"
)

func main() {
	// Create github client
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: os.Getenv("ghToken")},
	)
	tc := oauth2.NewClient(ctx, ts)

	client := github.NewClient(tc)

	// Open repo on fs, to determine upstream gitHub information, and make changes
	currentRepo, err := git.PlainOpen(".")
	if err != nil {
		panic(err)
	}
	// Open Github Repository
	owner, repoName := getRepoNameFromGitConfig(currentRepo)
	repo, _, err := client.Repositories.Get(ctx, owner, repoName)
	if err != nil {
		panic(err)
	}
	// Retrieve templateRepo from currentRepo so we can clone it
	templateRepo := repo.GetTemplateRepository()
	// Clone template repo to a temporary directory
	tempTemplateDir := "/tmp/foo/"
	cloneRepo(tempTemplateDir, templateRepo)
	// Copy files from template to current Repo
	copyFiles(tempTemplateDir, ".")

	// Template all files in current Repo
	files := getAllFiles(".")
	// Get git username and email, used for template variables
	gitName, gitEmail := getGitProfile()
	for _, file := range files {
		templateFile(file, repo, gitName, gitEmail)
	}

	// Create new branch for PR
	branches, _, err := client.Repositories.ListBranches(ctx, owner, repoName, &github.BranchListOptions{})
	if err != nil {
		panic(err)
	}
	newBranchName := getNewBranchName(branches)

	// If not clean, checkout, add, commit, push
	pushBranch(ctx, newBranchName, currentRepo)
}
