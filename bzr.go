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

// BzrRevRead retrieves the given or current wkspc rev.  A Revision struct
// pointer is returned (how filled out depends upon if the read is just the
// basic core/raw VCS revision or full data for the given VCS which will
// include tags, branches, timestamp info, author/committer, date, comment).
// Note: this reads one version but that could be expanded to take <N>
// revisions or a range, eg BzrRevRead(reader, <scope>, rev1, "..", rev2),
// without changing this methods params or return signature (but code
// changes  would be needed)
func BzrRevRead(r RevReader, scope ReadScope, vcsRev ...Rev) ([]Revisioner, string, error) {
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
		if specificRev != "" {
			output, err = exec.Command("bzr", "revno", "-r", specificRev).CombinedOutput()
		} else {
			output, err = exec.Command("bzr", "revno", "--tree").CombinedOutput()
		}
		if err != nil {
			return nil, string(output), err
		}
		rev.SetCore(Rev(strings.TrimSpace(string(output))))
		revs = append(revs, rev)
	} else {
		//FIXME: erik: get additional data about the version if possible (fix this)
		if specificRev != "" {
			output, err = exec.Command("bzr", "revno", "-r", specificRev).CombinedOutput()
		} else {
			output, err = exec.Command("bzr", "revno", "--tree").CombinedOutput()
		}
		if err != nil {
			return nil, string(output), err
		}
		rev.SetCore(Rev(strings.TrimSpace(string(output))))
		revs = append(revs, rev)
	}
	return revs, string(output), err
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

// BzrCheckRemote attempts to take a remote string (URL) and validate
// it (although with Bzr that doesn't work well) and set it if it is not
// currently set (this happens if a local clone exists only).  Returns:
// - string: this is the new remote (current remote returned if no new remote)
// - string: output of the Bzr command to try and determine the remote
// - error: non-nil if an error occurred
func BzrCheckRemote(e Existence, remote string) (string, string, error) {
	// With the other VCS we can check if the endpoint locally is different
	// from the one configured internally. But, with Bzr you can't. For example,
	// if you do `bzr branch https://launchpad.net/govcstestbzrrepo` and then
	// use `bzr info` to get the parent branch you'll find it set to
	// http://bazaar.launchpad.net/~mattfarina/govcstestbzrrepo/trunk/. Notice
	// the change from https to http and the path chance.
	// Here we set the remote to be the local one if none is passed in.
	var outStr string
	if exists, err := e.Exists(Wkspc); err == nil && exists && remote == "" {
		oldDir, err := os.Getwd()
		if err != nil {
			return remote, "", err
		}
		err = os.Chdir(e.WkspcPath())
		if err != nil {
			return remote, "", err
		}
		defer os.Chdir(oldDir)
		output, err := exec.Command("bzr", "info").CombinedOutput()
		if err != nil {
			return remote, string(output), err
		}
		outStr = string(output)
		m := bzrDetectURL.FindStringSubmatch(outStr)

		// If no remote was passed in but one is configured for the locally
		// checked out Bzr VCS pkg (repo) use that one.
		if m[1] != "" {
			return m[1], outStr, nil
		}
	}
	return remote, outStr, nil
}

