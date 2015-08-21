# VCS Repository Management for Go

Manage repos in varying version control systems with ease through a common
interface.

## Quick Usage

Quick usage:

	remote := "https://github.com/Masterminds/vcs"
    local, _ := ioutil.TempDir("", "go-vcs")
    repo, err := NewRepo(remote, local)

In this case `NewRepo` will detect the VCS is Git and return a `GitRepo`. All of
the repos implement the `Repo` interface with a common set of features between
them.

## Supported VCS

Git, SVN, Bazaar (Bzr), and Mercurial (Hg) are currently supported. They each
have their own type (e.g., `GitRepo`) that follow a simple naming pattern. Each
type implements the `Repo` interface and has a constructor (e.g., `NewGitRepo`).
The constructors have the same signature as `NewRepo`.

## Motivation

The package `golang.org/x/tools/go/vcs` provides some valuable functionality
for working with packages in repositories in varying source control management
systems. That package, while useful and well tested, is designed with a specific
purpose in mind. This version of the package is from github.com/Masterminds/vcs
and implements a Go interface based approach to VCS "repo" abstraction which is
something I was planning on trying but now don't need to (thanks much folks!).
