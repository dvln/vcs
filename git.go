package vcs

import (
	"os"
	"os/exec"
	"strings"

	"github.com/dvln/util/dir"
	"github.com/dvln/util/file"
	"github.com/dvln/util/url"
)

var defaultGitSchemes []string

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

// GitUpdate performs a git fetch and merge to an existing checkout (ie:
// a git pull).  The return is the output (string) and any error that may
// have occurred.
func GitUpdate(u *GitUpdater, rev ...Rev) (string, error) {
	// Perform a fetch to make sure everything is up to date, note that
	// we fetch all versions from the remote tracking branch (depending
	// on the revision of git)
	// FIXME: erik: check: may need to add string(rev[0]) as the last option
	//        if rev[0] is given so we fetch the right ref to merge with (?)
	var output string
	var err error
	if u.mirror {
		output, err = runFromWkspcDir(u.WkspcPath(), "git", "remote", "update", u.RemoteRepoName())
	} else {
		output, err = runFromWkspcDir(u.WkspcPath(), "git", "fetch", u.RemoteRepoName())
	}
	if err != nil {
		return output, err
	}

	if !u.mirror {
		// if user asks for a specific version on pull, use that
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
			pullOut, err = runFromWkspcDir(u.WkspcPath(), "git", "pull", rebaseStr)
		} else {
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
	oldDir, err := os.Getwd()
	if err != nil {
		return nil, "", err
	}
	err = os.Chdir(r.WkspcPath())
	if err != nil {
		return nil, "", err
	}
	specificRev := ""
	if vcsRev != nil && vcsRev[0] != "" {
		specificRev = string(vcsRev[0])
	}
	var output []byte
	defer os.Chdir(oldDir)
	rev := &Revision{}
	var revs []Revisioner
	if scope == CoreRev {
		// client just wants the core/base VCS revision only..
		if specificRev != "" {
			output, err = exec.Command("git", "log", "-1", "--format=%H", specificRev).CombinedOutput()
		} else {
			output, err = exec.Command("git", "log", "-1", "--format=%H").CombinedOutput()
		}
		if err != nil {
			return nil, string(output), err
		}
		rev.SetCore(Rev(strings.TrimSpace(string(output))))
		revs = append(revs, rev)
	} else {
		//FIXME: correct the full data one to run something like this:
		//% git log -1 --format='%H [%cD]%d'
		//a862506d017d643091368d53128447d032a03f54 [Thu, 11 Sep 2014 17:45:32 -0700] (HEAD -> topic, tag: main/7353, tag: acme__main__new__1410482753, origin/main, origin/HEAD)
		//should also add author+authorid+committer+committerid and then add in the
		//revision comment on the line following that data
		if specificRev != "" {
			output, err = exec.Command("git", "log", "-1", "--format=%H", specificRev).CombinedOutput()
		} else {
			output, err = exec.Command("git", "log", "-1", "--format=%H").CombinedOutput()
		}
		if err != nil {
			return nil, string(output), err
		}
		rev.SetCore(Rev(strings.TrimSpace(string(output))))
		revs = append(revs, rev)
	}
	return revs, string(output), nil
}

// GitExists verifies the wkspc or remote location is a Git repo,
// returns where it was found (or "" if not found) and any error
func GitExists(e Existence, l Location) (string, error) {
	var err error
	path := ""
	if l == Wkspc {
		if exists, err := dir.Exists(e.WkspcPath() + "/.git"); exists && err == nil {
			return e.WkspcPath(), nil // if non-bare, non-mirror we should find it here
		}
		if exists, err := dir.Exists(e.WkspcPath() + "/refs"); exists && err == nil {
			if exists, err = file.Exists(e.WkspcPath() + "/config"); exists && err == nil {
				return e.WkspcPath(), nil // if bare/mirror, we do a rough check here
			}
		}
		//FIXME: erik: if err != nil should use something like:
		//       out.WrapErrf(ErrNoExists, #, "%v git location, \"%s\", does not exist, err: %s", l, e.WkspcPath(), err)
	} else { // checking remote "URL" as well as possible for current VCS..
		remote := e.Remote()
		scheme := url.GetScheme(remote)
		// if we have a scheme then just see if the repo exists...
		if scheme != "" {
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
		//FIXME: erik: better erroring on failure to detect would be good here as well, such
		//             as a combined error on the various remote URL's checked and the error
		//             returned from each one (along with out.WrapErr's and such for tracing),
		//             also need to dump commands in exec.Command at Trace level and output
		//             here of those commands at Trace level also (along with other routines)

		if err == nil {
			return path, nil
		}
	}
	return path, err
}

// GitCheckRemote  attempts to take a remote string (URL) and validate
// it against any local repo and try and set it when it is empty.  Returns:
// - string: this is the new remote (current remote returned if no new remote)
// - string: output of the Bzr command to try and determine the remote
// - error: non-nil if an error occurred
func GitCheckRemote(e Existence, remote string) (string, string, error) {
	// Make sure the wkspc Git repo is configured the same as the remote when
	// a remote value was passed in, if no remote try and determine it here
	var outStr string
	if loc, err := e.Exists(Wkspc); err == nil && loc != "" {
		oldDir, err := os.Getwd()
		if err != nil {
			return remote, "", err
		}
		err = os.Chdir(e.WkspcPath())
		if err != nil {
			return remote, "", err
		}
		defer os.Chdir(oldDir)
		output, err := exec.Command("git", "config", "--get", "remote.origin.url").CombinedOutput()
		if err != nil {
			return remote, string(output), err
		}

		outStr = string(output)
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
