package vcs

import (
	"os"
	"os/exec"
	"strings"

	"github.com/dvln/util/dir"
)

// GitGet is used to perform an initial clone of a repository, returns
// output and error
// FIXME: erik: need to support verbosity levels for output
func GitGet(g Getter, rev ...Rev) (string, error) {
	output, err := run("git", "clone", g.Remote(), g.WkspcPath())
	if rev != nil {
		checkoutOut, err := g.RevSet(rev[0])
		if err != nil {
			return checkoutOut, err
		}
		output = output + checkoutOut
	}
	return output, err
}

// GitUpdate performs an git fetch and merge to an existing checkout (or
// a git pull).  The return is the output (string) and any error that may
// have occurred.
func GitUpdate(u Updater, rev ...Rev) (string, error) {
	// Perform a fetch to make sure everything is up to date, note that
	// we fetch all versions from the remote tracking branch (depending
	// on the revision of git)
	// FIXME: erik: check: may need to add string(rev[0]) as the last option
	//        if rev[0] is given so we fetch the right ref to merge with (?)
	output, err := runFromWkspcDir(u.WkspcPath(), "git", "fetch", u.RemoteRepoName())
	if err != nil {
		return output, err
	}
	var pullOut string
	// if user asks for a specific version on pull, use that
	if rev == nil || ( rev != nil && rev[0] == "" ) {
		pullOut, err = runFromWkspcDir(u.WkspcPath(), "git", "pull")
	} else {
		pullOut, err = runFromWkspcDir(u.WkspcPath(), "git", "pull", u.RemoteRepoName(), string(rev[0]))
	}
	output = output + pullOut
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

// GitRevRead retrieves the current workspace revision.  The returned item is a
// revision structure (how filled out depends upon if the read is optimized
// for speed in which case just the raw VCS revision is read, or if full
// data is requested than tags/branch refs, etc and "timestamp" of the
// revision are added.
//FIXME: erik: consider returning a Revisioner intfc (?)
func GitRevRead(r RevReader, scope ...ReadScope) (*Revision, string, error) {
	oldDir, err := os.Getwd()
	if err != nil {
		return nil, "", err
	}
	err = os.Chdir(r.WkspcPath())
	if err != nil {
		return nil, "", err
	}
	var output []byte
	defer os.Chdir(oldDir)
	rev := &Revision{}
	if scope == nil || ( scope != nil && scope[0] == CoreRev ) {
		// client just wants the core/base VCS revision only..
		output, err = exec.Command("git", "log", "-1", "--format=%H").CombinedOutput()
		if err != nil {
			return nil, string(output), err
		}
		rev.SetCore(Rev(strings.TrimSpace(string(output))))
	} else {
		//FIXME: erik: correct the full data one to run something like this:
		//% git log -1 --format='%H [%cD]%d'
        //a862506d017d643091368d53128447d032a03f54 [Thu, 11 Sep 2014 17:45:32 -0700] (HEAD -> topic, tag: main/7353, tag: acme__main__new__1410482753, origin/main, origin/HEAD)
		//should also add author+authorid+committer+committerid and then add in the
		//revision comment on the line following that data
		output, err := exec.Command("git", "log", "-1", "--format=%H").CombinedOutput()
		if err != nil {
			return nil, string(output), err
		}
		rev.SetCore(Rev(strings.TrimSpace(string(output))))
	}
	return rev, string(output), nil
}

// GitExists verifies the wkspc or remote location is a Git repo.
func GitExists(e Existence, l Location) (bool, error) {
	var err error
	if l == Wkspc {
		if there, err := dir.Exists(e.WkspcPath() + "/.git"); there && err == nil {
			return true, nil
		}
		//FIXME: erik: if err != nil should use something like:
		//       out.WrapErrf(ErrNoExists, #, "%v git location, \"%s\", does not exist, err: %s", l, e.WkspcPath(), err)
	} else {
		//FIXME: erik: need to actually check if remote repo exists ;)
		// should use this "ErrNoExist" from repo.go if doesn't exist
		return true, nil
	}
	return false, err
}

// GitCheckRemote  attempts to take a remote string (URL) and validate
// it against any local repo and try and set it when it is empty.  Returns:
// - string: this is the new remote (current remote returned if no new remote)
// - string: output of the Bzr command to try and determine the remote
// - error: non-nil if an error occurred
func GitCheckRemote (e Existence, remote string) (string, string, error) {
	// Make sure the wkspc Git repo is configured the same as the remote when
	// a remote value was passed in, if no remote try and determine it here
	var outStr string
	if exists, err := e.Exists(Wkspc); err == nil && exists {
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
