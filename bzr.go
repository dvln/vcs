package vcs

import (
	"os"
	"os/exec"
	"strings"
	"regexp"

    "github.com/dvln/util/dir"
)

var bzrDetectURL = regexp.MustCompile("parent branch: (?P<foo>.+)\n")

// BzrGet is used to perform an initial clone of a repository.
func BzrGet(g Getter, rev ...Rev) (string, error) {
	var output string
	var err error
	if rev == nil || ( rev != nil && rev[0] == "" ) {
		output, err = run("bzr", "branch", g.Remote(), g.WkspcPath())
	} else {
		output, err = run("bzr", "branch", "-r", string(rev[0]), g.Remote(), g.WkspcPath())
	}
	return output, err
}

// BzrUpdate performs a Bzr pull and update to an existing checkout.
func BzrUpdate(u Updater, rev ...Rev) (string, error) {
	output, err := runFromWkspcDir(u.WkspcPath(), "bzr", "pull")
	if err != nil {
		return output, err
	}
	var updOut string
	if rev == nil || ( rev != nil && rev[0] == "" ) {
		updOut, err = runFromWkspcDir(u.WkspcPath(), "bzr", "update")
	} else {
		updOut, err = runFromWkspcDir(u.WkspcPath(), "bzr", "update", "-r", string(rev[0]))
	}
	output = output + updOut
	return output, err
}

// BzrRevSet sets the wkspc revision of a pkg currently checked out via Bzr.
// Note that a single specific revision must be given (vs a generic
// Revision structure as such a struct may have <N> different valid rev's
// that reference the revision).  The output (if any) and any error
// is returned from the svn update run.
func BzrRevSet(r RevSetter, rev Rev) (string, error) {
	return runFromWkspcDir(r.WkspcPath(), "bzr", "update", "-r", string(rev))
}

// BzrRevRead retrieves the current version (and any cmd out for what was run)
func BzrRevRead(r RevReader, scope ...ReadScope) (*Revision, string, error) {
	oldDir, err := os.Getwd()
	if err != nil {
		return nil, "", err
	}
	err = os.Chdir(r.WkspcPath())
	if err != nil {
		return nil, "", err
	}
	defer os.Chdir(oldDir)
	var output []byte

	rev := &Revision{}
	if scope == nil || ( scope != nil && scope[0] == CoreRev ) {
		// client just wants the core/base VCS revision only..
		output, err = exec.Command("bzr", "revno", "--tree").CombinedOutput()
		if err != nil {
			return nil, string(output), err
		}
		rev.SetCore(Rev(strings.TrimSpace(string(output))))
	} else {
		//FIXME: erik: get additional data about the version if possible (fix this)
		output, err = exec.Command("bzr", "revno", "--tree").CombinedOutput()
		if err != nil {
			return nil, string(output), err
		}
		rev.SetCore(Rev(strings.TrimSpace(string(output))))
	}
	return rev, string(output), err
}

// BzrExists verifies the wkspc or remote location is of the Bzr repo type
func BzrExists(e Existence, l Location) (bool, error) {
	var err error
	if l == Wkspc {
		if there, err := dir.Exists(e.WkspcPath() + "/.bzr"); there && err == nil {
			return true, nil
		}
		//FIXME: erik: if err != nil should use something like:
		//       out.WrapErrf(ErrNoExists, #, "%v bzr location, \"%s\", does not exist, err: %s", l, e.WkspcPath(), err)
	} else {
		//FIXME: erik: need to actually check if remote repo exists ;)
		// should use this "ErrNoExist" from repo.go if doesn't exist
		return true, nil
	}
	return false, err
}

