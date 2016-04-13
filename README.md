# VCS Repository Management for Go

Manage VCS pkgs (repos) in varying version control systems with ease through
common interfaces.

## Quick Usage

Quick usage:

	remote := "https://github.com/dvln/vcs"
    local, _ := ioutil.TempDir("", "dlvn-vcs")
    vcsReader, err := NewReader(remote, local)

In this case `NewReader` will detect the VCS is Git and return a `GitReader` to
read the VCS package/repo. All of VCS's implement the Reader interface with a
common set of features between them.  Beyond that there are more specific interfaces
such as the Getter and Updater interfaces (subset of the overall Reader interface,
only providing those interfaces needed to complete the get/clone or update/merge
classes of functionality.

## Supported VCS

Git, SVN, Bazaar (Bzr), and Mercurial (Hg) are currently supported. They each
have their own type (e.g., `GitReader`) that follow a simple naming pattern. Each
type implements the `Reader` interface and has a constructor (e.g., `NewGitReader`).
The constructors have the same signature as `NewReader`.

## Motivation

The package `golang.org/x/tools/go/vcs` provides some valuable functionality
for working with packages in repositories in varying source control management
systems. Beyond that other packages such as the 'nut' VCS tool vcs package and
the github.com/Masterminds/vcs ('glide' tool) packages all provided further
insights and capabilities.  This is a fork of the Masterminds vcs pkg (thanks
much to all folks above and especially the Masterminds folks) but it is
heavily changed and moves in somewhat of a different direction at this point.

## Concurrency: Goroutine Friendly?

Currently only the git backend is Goroutine friendly (get/clone, upd/fetch/pull,
revision reading, etc) and only the git backend implements the more extensive
capabilities around mirror clones (or not) and mirror updates (or not) as well
as specific fetch/delete ref targets working in both regular and mirror/bare
clones.  This is an early version with this target... after refining the git
solution the next SCM targeted for "concurrent friendly" is Hg.

## Status

This is very early pre-release work that is not in use or ready for regular
use (currently: v0.2.0).  Feel free to fork and give suggestions of course
but know it will be undergoing dramatic change throughout 2016 (at least).

