package vcs

import (
	"os"
	"os/exec"
	"strings"
	"regexp"

    "github.com/dvln/util/dir"
)

var svnDetectURL = regexp.MustCompile("URL: (?P<foo>.+)\n")

// SvnGet is used to perform an initial checkout of a repository.
// Note, because SVN isn't distributed this is a checkout without
// a clone.  One can checkout an optionally passed in revision.
func SvnGet(g Getter, rev ...Rev) (string, error) {
	var output string
	var err error
	if rev == nil || ( rev != nil && rev[0] == "" ) {
		output, err = run("svn", "checkout", g.Remote(), g.WkspcPath())
	} else {
		output, err = run("svn", "checkout", "-r", string(rev[0]), g.Remote(), g.WkspcPath())
	}
	return output, err
}

// SvnUpdate performs an SVN update to an existing checkout (ie: a merge).
func SvnUpdate(u Updater, rev ...Rev) (string, error) {
	var output string
	var err error
	if rev == nil || ( rev != nil && rev[0] == "" ) {
		output, err = runFromWkspcDir(u.WkspcPath(), "svn", "update")
	} else {
		output, err = runFromWkspcDir(u.WkspcPath(), "svn", "update", "-r", string(rev[0]))
	}
	return output, err
}

// SvnRevSet sets the wkspc revision of a pkg currently checked out via
// Svn.  Note that a single specific revision must be given (vs a generic
// Revision structure as such a struct may have <N> different valid rev's
// that reference the revision).  The output (if any) and any error
// is returned from the svn update run.
func SvnRevSet(r RevSetter, rev Rev) (string, error) {
	return runFromWkspcDir(r.WkspcPath(), "svn", "update", "-r", string(rev))
}

// SvnRevRead retrieves the current version.
func SvnRevRead(r RevReader, scope ...ReadScope) (*Revision, string, error) {
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
		//FIXME: erik: based on SVN docs this doesn't seem like it
		//       can handle the various output formats correctly with
		//       modifiers like "<rev#>M" or "<rev#>S" or "<rev#>:<rev#>",
		//       see svnversion -h, perhaps update
		output, err = exec.Command("svnversion", ".").CombinedOutput()
		if err != nil {
			return nil, string(output), err
		}
		rev.SetCore(Rev(strings.TrimSpace(string(output))))
	} else {
		//FIXME: erik: this needs to add more data if possible for SVN
		output, err = exec.Command("svnversion", ".").CombinedOutput()
		if err != nil {
			return nil, string(output), err
		}
		rev.SetCore(Rev(strings.TrimSpace(string(output))))
	}
	return rev, string(output), err
}

// SvnExists verifies the wkspc or remote location is of the SVN type
func SvnExists(e Existence, l Location) (bool, error) {
	var err error
	if l == Wkspc {
		if there, err := dir.Exists(e.WkspcPath() + "/.svn"); there && err == nil {
			return true, nil
		}
		//FIXME: erik: if err != nil should use something like:
		//       out.WrapErrf(ErrNoExists, #, "%v SVN location, \"%s\", does not exist, err: %s", l, e.WkspcPath(), err)
	} else {
		//FIXME: erik: need to actually check if remote repo exists ;)
		// should use this "ErrNoExist" from repo.go if doesn't exist
		return true, nil
	}
	return false, err
}

