package vcs

import (
	"fmt"
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

// SvnRevRead retrieves the given or current wkspc rev.  A Revision struct
// pointer is returned (how filled out depends upon if the read is just the
// basic core/raw VCS revision or full data for the given VCS which will
// include tags, branches, timestamp info, author/committer, date, comment).
// Note: this reads one version but that could be expanded to take <N>
// revisions or a range, eg SvnRevRead(reader, <scope>, rev1, "..", rev2),
// without changing this methods params or return signature (but code
// changes  would be needed)
func SvnRevRead(r RevReader, scope ReadScope, vcsRev ...Rev) ([]Revisioner, string, error) {
	specificRev := ""
	if vcsRev != nil && vcsRev[0] != "" {
		specificRev = string(vcsRev[0])
	}
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
	var revs []Revisioner
	if scope == CoreRev {
		// client just wants the core/base VCS revision only..
		//FIXME: erik: based on SVN docs this doesn't seem like it
		//       can handle the various output formats correctly with
		//       modifiers like "<rev#>M" or "<rev#>S" or "<rev#>:<rev#>",
		//       see svnversion -h, perhaps update
		//       Also: specificRev isn't implemented for SVN, is it possible?
		if specificRev != "" {
			return nil, "", fmt.Errorf("Reading specified revision, %s, not supported by SVN", specificRev)
		} else {
			output, err = exec.Command("svnversion", ".").CombinedOutput()
		}
		if err != nil {
			return nil, string(output), err
		}
		rev.SetCore(Rev(strings.TrimSpace(string(output))))
		revs = append(revs, rev)
	} else {
		//FIXME: erik: this needs to add more data if possible for SVN
		//       Also: specificRev isn't implemented for SVN, is it possible?
		if specificRev != "" {
			return nil, "", fmt.Errorf("Reading specified revision, %s, not supported by SVN", specificRev)
		} else {
			output, err = exec.Command("svnversion", ".").CombinedOutput()
		}
		if err != nil {
			return nil, string(output), err
		}
		rev.SetCore(Rev(strings.TrimSpace(string(output))))
		revs = append(revs, rev)
	}
	return revs, string(output), err
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

// SvnCheckRemote  attempts to take a remote string (URL) and validate
// it against any local checkout and try and set it when it is empty.  Returns:
// - string: this is the new remote (current remote returned if no new remote)
// - string: output of the Bzr command to try and determine the remote
// - error: non-nil if an error occurred
func SvnCheckRemote (e Existence, remote string) (string, string, error) {
	// Make sure the wkspc Svn repo is configured the same as the remote when
	// A remote value was passed in.
	var outStr string
	if exists, err := e.Exists(Wkspc); err == nil && exists {
		// An SVN repo was found so test that the URL there matches
		// the repo passed in here.
		output, err := exec.Command("svn", "info", e.WkspcPath()).CombinedOutput()
		outStr = string(output)
		if err != nil {
			return remote, outStr, err
		}

		m := svnDetectURL.FindStringSubmatch(outStr)
		if remote != "" && m[1] != "" && m[1] != remote {
			return remote, outStr, ErrWrongRemote
		}

		// If no remote was passed in but one is configured for the locally
		// checked out Svn repo use that one.
		if remote == "" && m[1] != "" {
			return m[1], outStr, nil
		}
	}
	return remote, outStr, nil
}

