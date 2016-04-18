# VCS Repository Management for Go

Manage VCS pkgs (repos) in varying version control systems with ease through
common interfaces.

## Supported VCS operations

Currently this has limited support for "get" (clone equivalent), "update" (various
forms of fetching/merging/updating from a remote repo to the local instance) as
well as some basic "reading" capability which can read version info from a
local workspace or a given version (or set the current version to something
else for reading... so perhaps a bit of a misnomer at the moment).

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
much to all folks above and especially the Masterminds folks) but it has
diverged heavily and moves in different direction at this point.

The thought here is that most SCM/VCS tools seem to need specific operations
such as update or clone (get) or diff or possibly reading data that already
exists in the repo.  Most o these also need to do things like check basic
existence and describe where the repo is and what kind of op you want to
perform (a description).  The idea here is to set up interfaces for a
Describer and to make an Updater be a describer but, beyond that, be
primarily focused on being an updater (with some basic ability to
check for existence and such).  Aside: would use an Existence but
how Go deals with multiple interfaces gets a little tricky when
there is overlap (ie: if an Existence is a Describer and an Updater
is a Describer then it's not possible to add Existence to Updater
even if the signatures of the describe methods are exactly 
matching (which, of course, they are).  Due to this you will
see things like the routine Exists() defined in the Updater
interface (vs sucking it in via the Existence interface).

With that in mind you'll see that one would instantiate the
needed op by only using the minimal data needed for that
operation.  I find this to be useful when interacting with
the SCM... and it keeps the interface to a bare minimum for
whatever SCM op one might need to perform.

## Concurrency: Goroutine Friendly?

Currently only the git backend is Goroutine friendly (get/clone, upd/fetch/pull,
revision reading, etc) and only the git backend implements the more extensive
capabilities around mirror clones (or not) and mirror updates (or not) as well
as specific fetch/delete ref targets working in both regular and mirror/bare
clones.  This is an early version with this target... after refining the git
solution the next SCM targeted for "concurrent friendly" is Hg.

## Usage

Haven't fleshed this README out as the API has been in flux.  The test
files should give a general idea for use, eg: git_test.go for example.
Here's a quick example:

```go
	import "github.com/dvln/vcs"

    ...
	mirror := true
	// Perform an update operation, not mirroring (just a regular clone), don't use
	// rebase (one can indcate don't use rebase, use rebase or use the users default)
	// ... note: we're updating a test repo here so that's what the tempDir points to ...
	gitUpdater, err := NewUpdater("https://github.com/dvln/git-test-repo", tempDir+sep+"VCSTestRepo", !mirror, RebaseUser, nil)
	if err != nil {
		t.Fatal(err)
	}
	results, err := gitUpdater.Update()
	if err != nil {
		t.Fatalf("Failed to run git update, error: %s, results:\n%s", err, results)
	}
```

That's about it.  You created an updater (gitUpdater) and then you ran update on it.
The various options allow one to do a mirror update ("git remote update --prune") as
well as add specific git refs (eg: "refs/heads/master") via the last option.  How those
refs are fetched depends upon mirror mode or not (if mirroring fetch them to the same
name, if not mirroring fetch heads to refs/remotes/origin/master for example).  As
always the code is the master source for info.

## Status

This is very early pre-release work that is not in use or ready for regular
use (currently: v0.3.0).  Feel free to fork and give suggestions of course
but know it will be undergoing dramatic change throughout 2016 (at least).
If you use it make sure you vendor it locally as the API may change.

