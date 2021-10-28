# ggth
[![Go Report Card](https://goreportcard.com/badge/github.com/Jmainguy/ggth)](https://goreportcard.com/badge/github.com/Jmainguy/ggth)
[![Release](https://img.shields.io/github/release/Jmainguy/ggth.svg?style=flat-square)](https://github.com/Jmainguy/ggth/releases/latest)
[![Coverage Status](https://coveralls.io/repos/github/Jmainguy/ggth/badge.svg?branch=main)](https://coveralls.io/github/Jmainguy/ggth?branch=main)

Golang Github Template Helper

## Usage
```/bin/bash
Usage of ./ggth:
  -description string
    	Description of repo you wish to create
  -githubOwner string
    	Github owner
  -githubRepo string
    	Github repo to source, owned by githubOwner
  -templateName string
    	Name of template, if you wish to create a new repo
  -templateOwner string
    	Owner of template, if you wish to create a new repo
```

If -templateName, and -templateOwner are defined, it will attempt to create the repo on Github from this template. 

-githubOwner and -githubRepo must be defined for it to know where to put it. 

-description should be defined if you want the repo to have a description

## Env Variables
```/bin/bash
export ghToken=ghp_FAKETOKEN
export ghUsername=Jmainguy
```

You should export these variables for it to be able to auth to github.

## PreBuilt Binaries
Grab Binaries from [The Releases Page](https://github.com/Jmainguy/ggth/releases)

## Install

### Homebrew

```/bin/bash
brew install Jmainguy/tap/ggth
```

## Build
```/bin/bash
export GO111MODULE=on
go build
```
