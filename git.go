package vcs

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/dvln/out"
	"github.com/dvln/util/dir"
	"github.com/dvln/util/file"
	"github.com/dvln/util/url"
)

var defaultGitSchemes []string
var refsRegex = regexp.MustCompile(`^refs/heads/(.*)$`)

func init() {
	SetDefaultGitSchemes(nil)
}

// GitGet is used to perform an initial clone of a repository, returns
// output and error
func GitGet(g *GitGetter, rev ...Rev) (string, error) {
	mirrorStr := ""
	if g.mirror { // if in mirror clone mode add in --mirror
		mirrorStr = "--mirror"
	}
	output, err := run("git", "clone", mirrorStr, g.Remote(), g.WkspcPath())
	if rev != nil {
		checkoutOut, err := g.RevSet(rev[0])
		if err != nil {
			return checkoutOut, err
		}
		output = output + checkoutOut
	}
	return output, err
}

// gitUpdateRefs is fired if GitUpdate() gets specific refs to operate
// on... meaning fetch or delete ops (at this point).
func gitUpdateRefs(u *GitUpdater) (string, error) {
	var output string
	var err error
	runOpt := "-C"
	runDir := u.WkspcPath()
	for ref, refOp := range u.refs {
		switch refOp {
		case RefDelete:
			output, err = run("git", runOpt, runDir, "update-ref", "-d", ref)
		case RefFetch:
			if u.mirror { // request is to mirror refs exactly, do so
				refSpec := fmt.Sprintf("+%s:%s", ref, ref)
				output, err = run("git", runOpt, runDir, "fetch", u.RemoteRepoName(), refSpec)
			} else { // normal fetch requested, heads remapped, all else comes in "as-is"
				m := refsRegex.FindStringSubmatch(ref) // look for refs/heads/<name> refs
				if m[1] != "" {                        // if it was a refs/heads then map it:
					remoteRef := fmt.Sprintf("refs/remotes/%s/%s", u.RemoteRepoName(), m[1])
					refSpec := fmt.Sprintf("+%s:%s", ref, remoteRef)
					output, err = run("git", runOpt, runDir, "fetch", u.RemoteRepoName(), refSpec)
				} else { // bring in tags/etc under the same namespace
					refSpec := fmt.Sprintf("+%s:%s", ref, ref)
					output, err = run("git", runOpt, runDir, "fetch", u.RemoteRepoName(), refSpec)
				}
			}
		default:
			err = out.NewErrf(4502, "Update refs: invalid ref operation given \"%v\", clone: %s", refOp, u.WkspcPath())
		}
	}
	return output, err
}

// GitUpdate performs a git fetch and merge to an existing checkout (ie:
// a git pull).  The return is the output (string) and any error that may
// have occurred.
func GitUpdate(u *GitUpdater, rev ...Rev) (string, error) {
	// Perform required fetches optionally with pulls as well as handling
	// more specific fetches on single refs (or deletion of refs)... has
	// some handling of mirror/bare clones vs local clones and for std
	// clones can do rebase type pulls (if that section of the routine is
	// reached).
	var output string
	var err error
	runOpt := "-C"
	runDir := u.WkspcPath()
	if u.refs != nil {
		return gitUpdateRefs(u)
	}
	if u.mirror {
		output, err = run("git", runOpt, runDir, "remote", "update", "--prune", u.RemoteRepoName())
	} else {
		output, err = run("git", runOpt, runDir, "fetch", u.RemoteRepoName())
	}
	if err != nil {
		return output, err
	}

	bareRepo := false
	gitDir, workTree, err := findGitDirs(runDir)
	if err != nil {
		return "", err
	}
	if gitDir == runDir && workTree == "" {
		bareRepo = true
	}
	if !u.mirror && !bareRepo { // if not a mirror and a regular clone
		// Try and run a git pull to do the merge|rebase op
		rebaseStr := "--rebase=false"
		switch u.rebase {
		case RebaseTrue:
			rebaseStr = "--rebase=true"
		case RebasePreserve:
			rebaseStr = "--rebase=preserve"
		default:
			rebaseStr = ""
		}
		var pullOut string
		if rev == nil || (rev != nil && rev[0] == "") {
			pullOut, err = runFromWkspcDir(u.WkspcPath(), "git", "pull", rebaseStr, u.RemoteRepoName())
		} else { // if user asks for a specific version on pull, use that
			pullOut, err = runFromWkspcDir(u.WkspcPath(), "git", "pull", rebaseStr, u.RemoteRepoName(), string(rev[0]))
		}
		output = output + pullOut
	}
	return output, err
}

// GitRevSet sets the wkspc revision of a pkg currently checked out via Git.
// Note that a single specific revision must be given vs a generic
// Revision structure (since it may have <N> different valid rev's
// that reference the revision, this one decides exactly the one
// the client wishes to "set" or checkout in the wkspc).
func GitRevSet(r RevSetter, rev Rev) (string, error) {
	return runFromWkspcDir(r.WkspcPath(), "git", "checkout", string(rev))
}

// GitRevRead retrieves the given or current wkspc rev.  A Revision struct
// pointer is returned (how filled out depends upon if the read is just the
// basic core/raw VCS revision or full data for the given VCS which will
// include tags, branches, timestamp info, author/committer, date, comment).
// Note: this reads one version but that could be expanded to take <N>
// revisions or a range, eg GitRevRead(reader, <scope>, rev1, "..", rev2),
// without changing this methods params or return signature (but code
// changes  would be needed)
func GitRevRead(r RevReader, scope ReadScope, vcsRev ...Rev) ([]Revisioner, string, error) {
	runOpt := "-C"
	runDir := r.WkspcPath()
	specificRev := ""
	if vcsRev != nil && vcsRev[0] != "" {
		specificRev = string(vcsRev[0])
	}
	var output string
	rev := &Revision{}
	var revs []Revisioner
	var err error
	if scope == CoreRev {
		// client just wants the core/base VCS revision only..
		if specificRev != "" {
			output, err = run("git", runOpt, runDir, "log", "-1", "--format=%H", specificRev)
		} else {
			output, err = run("git", runOpt, runDir, "log", "-1", "--format=%H")
		}
		if err != nil {
			return nil, output, err
		}
		rev.SetCore(Rev(strings.TrimSpace(output)))
		revs = append(revs, rev)
	} else {
		//FIXME: correct the full data one to run something like this:
		//% git log -1 --format='%H [%cD]%d'
		//a862506d017d643091368d53128447d032a03f54 [Thu, 11 Sep 2014 17:45:32 -0700] (HEAD -> topic, tag: main/7353, tag: acme__main__new__1410482753, origin/main, origin/HEAD)
		//should also add author+authorid+committer+committerid and then add in the
		//revision comment on the line following that data
		if specificRev != "" {
			output, err = run("git", runOpt, runDir, "log", "-1", "--format=%H", specificRev)
		} else {
			output, err = run("git", runOpt, runDir, "log", "-1", "--format=%H")
		}
		if err != nil {
			return nil, output, err
		}
		rev.SetCore(Rev(strings.TrimSpace(output)))
		revs = append(revs, rev)
	}
	return revs, output, nil
}

// GitExists verifies the wkspc or remote location is a Git repo,
// returns where it was found (or "" if not found) and any error
func GitExists(e Existence, l Location) (string, error) {
	var err error
	path := ""
	if l == Wkspc {
		_, _, err := findGitDirs(e.WkspcPath()) // see if git clone there
		if err == nil {
			return e.WkspcPath(), nil // it's a local git clone, success
		}
	} else { // checking remote "URL" as well as possible for current VCS..
		remote := e.Remote()
		scheme := url.GetScheme(remote)
		if scheme != "" { // if we have a scheme then see if the repo exists...
			_, err = exec.Command("git", "ls-remote", remote).CombinedOutput()
			if err == nil {
				path = remote
			}
		} else {
			vcsSchemes := e.Schemes()
			for _, scheme = range vcsSchemes {
				_, err = exec.Command("git", "ls-remote", scheme+"://"+remote).CombinedOutput()
				if err == nil {
					path = scheme + "://" + remote
					break
				}
			}
		}
		if err == nil {
			return path, nil
		}
		err = out.WrapErrf(ErrNoExist, 4501, "Remote git location, \"%s\", does not exist, err: %s", e.WkspcPath(), err)
	}
	return path, err
}

// GitCheckRemote  attempts to take a remote string (URL) and validate
// it against any local repo and try and set it when it is empty.  Returns:
// - string: this is the new remote (current remote returned if no new remote)
// - string: output of the Git command to try and determine the remote
// - error: non-nil if an error occurred
func GitCheckRemote(e Existence, remote string) (string, string, error) {
	// Make sure the wkspc Git repo is configured the same as the remote when
	// a remote value was passed in, if no remote try and determine it here
	var outStr string
	if loc, err := e.Exists(Wkspc); err == nil && loc != "" {
		runOpt := "-C"
		runDir := loc
		output, err := run("git", runOpt, runDir, "config", "--get", "remote.origin.url")
		if err != nil {
			return remote, output, err
		}
		outStr = output
		localRemote := strings.TrimSpace(outStr)
		if remote != "" && localRemote != remote {
			return remote, outStr, ErrWrongRemote
		}

		// If no remote was passed in but one is configured for the locally
		// checked out Git repo use that one.
		if remote == "" && localRemote != "" {
			return localRemote, outStr, nil
		}
	}
	return remote, outStr, nil
}

// SetDefaultGitSchemes allows one to override the default ordering
// and set of git remote URL schemes to try for any remote that has
// no scheme provided, defaults to Go core list for now.
func SetDefaultGitSchemes(schemes []string) {
	if schemes == nil {
		defaultGitSchemes = []string{"git", "https", "http", "git+ssh"}
	} else {
		defaultGitSchemes = schemes
	}
}

// findGitDirs expects to be pointed at a git workspace, either
// bare or standard.  It'll find the gitdir and worktree dirs
// and return them, if it fails it'll return non-nil err.  Params:
//	path (string): path to the git workspace
// Returns:
//	gitDir (string): path to git metadata location
//	workTreeDir (string): working tree, "" if bare clone
//	err (error): a valid error if unable to find a git repo
func findGitDirs(path string) (string, string, error) {
	gitDir := filepath.Join(path, ".git") // see if std git clone
	var err error
	var exists bool
	if exists, err = dir.Exists(gitDir); exists && err == nil {
		return gitDir, path, nil
	}
	gitRefsDir := filepath.Join(path, "refs")
	if exists, err = dir.Exists(gitRefsDir); exists && err == nil {
		gitConfigFile := filepath.Join(path, "config")
		if exists, err = file.Exists(gitConfigFile); exists && err == nil {
			return path, "", nil
		}
	}
	return "", "", out.WrapErrf(ErrNoExist, 4500, "Unable to find valid git dir under path: %s, err: %s", path, err)
}
