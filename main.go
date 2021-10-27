package main

import (
	"context"
	"fmt"
	"os"

	"flag"

	git "github.com/go-git/go-git/v5"
	"github.com/google/go-github/v39/github"
	"golang.org/x/oauth2"
)

func main() {
	githubOwnerPtr := flag.String("githubOwner", "", "Github owner")
	githubRepoPtr := flag.String("githubRepo", "", "Github repo to source, owned by githubOwner")
	//create
	flag.Parse()

	githubUsername := os.Getenv("ghUsername")
	githubPassword := os.Getenv("ghToken")
	githubOwner := *githubOwnerPtr
	githubRepo := *githubRepoPtr

	// Create github client
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: githubPassword},
	)
	tc := oauth2.NewClient(ctx, ts)

	client := github.NewClient(tc)

	workDir := "."
	var sourceRepo *github.Repository
	if *githubOwnerPtr != "" {
		if *githubRepoPtr != "" {
			workDir = "/tmp/source/"
			// Clone template repo to a temporary directory
			remoteSourceRepo, _, err := client.Repositories.Get(ctx, githubOwner, githubRepo)
			if err != nil {
				panic(err)
			}

			cloneRepo(remoteSourceRepo, workDir, githubUsername, githubPassword)
			sourceRepo = remoteSourceRepo
		}
	}

	// Open repo on fs, to determine upstream gitHub information, and make changes

	currentRepo, err := git.PlainOpen(workDir)
	if err != nil {
		panic(err)
	}
	blankGithubRepo := &github.Repository{}
	if sourceRepo == blankGithubRepo {
		// Working on local repo, did not clone it.
		// Open Github Repository
		*githubOwnerPtr, *githubRepoPtr = getRepoNameFromGitConfig(currentRepo)
		sourceRepo, _, err = client.Repositories.Get(ctx, githubOwner, githubRepo)
		if err != nil {
			panic(err)
		}
	}
	// Retrieve templateRepo from currentRepo so we can clone it
	templateRepo := sourceRepo.GetTemplateRepository()
	// Clone template repo to a temporary directory
	tempTemplateDir := "/tmp/foo/"
	cloneRepo(templateRepo, tempTemplateDir, githubUsername, githubPassword)
	// Copy files from template to current Repo
	copyFiles(tempTemplateDir, workDir)

	// Template all files in current Repo
	files := getAllFiles(workDir)
	// Get git username and email, used for template variables
	gitName, gitEmail := getGitProfile()
	for _, file := range files {
		templateFile(file, sourceRepo, gitName, gitEmail)
	}

	// Create new branch for PR
	branches, _, err := client.Repositories.ListBranches(ctx, githubOwner, githubRepo, &github.BranchListOptions{})
	if err != nil {
		panic(err)
	}
	newBranchName := getNewBranchName(branches)

	// If not clean, checkout, add, commit, push
	pushBranch(ctx, currentRepo, newBranchName, githubUsername, githubPassword)

	// Create PR from branch
	pull := &github.NewPullRequest{
		Title:               github.String("Automated PR from ggth"),
		Head:                github.String(newBranchName),
		Base:                sourceRepo.DefaultBranch,
		Body:                github.String("This is a automated PR by the https://github.com/Jmainguy/ggth tool"),
		MaintainerCanModify: github.Bool(true),
	}
	pr, _, err := client.PullRequests.Create(ctx, githubOwner, githubRepo, pull)
	if err != nil {
		panic(err)
	}
	fmt.Println(*pr.HTMLURL)
}
