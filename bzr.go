package vcs

import (
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/dvln/util/dir"
	"github.com/dvln/util/url"
)

var bzrDetectURL = regexp.MustCompile("parent branch: (?P<foo>.+)\n")
var defaultBzrSchemes []string

func init() {
	SetDefaultBzrSchemes(nil)
}

// BzrGet is used to perform an initial clone of a repository.
func BzrGet(g *BzrGetter, rev ...Rev) (string, error) {
	var output string
	var err error
	if rev == nil || (rev != nil && rev[0] == "") {
		output, err = run("bzr", "branch", g.Remote(), g.WkspcPath())
	} else {
		output, err = run("bzr", "branch", "-r", string(rev[0]), g.Remote(), g.WkspcPath())
	}
	return output, err
}

// BzrUpdate performs a Bzr pull and update to an existing checkout.
func BzrUpdate(u *BzrUpdater, rev ...Rev) (string, error) {
	output, err := runFromWkspcDir(u.WkspcPath(), "bzr", "pull")
	if err != nil {
		return output, err
	}
	var updOut string
	if rev == nil || (rev != nil && rev[0] == "") {
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

// BzrExists verifies the wkspc or remote location is of the Bzr repo type,
// returns where it was found ("" if not found) and any error
func BzrExists(e Existence, l Location) (string, error) {
	var err error
	path := ""
	if l == Wkspc {
		if exists, err := dir.Exists(e.WkspcPath() + "/.bzr"); exists && err == nil {
			return e.WkspcPath(), nil
		}
		//FIXME: erik: if err != nil should use something like:
		//       out.WrapErrf(ErrNoExists, #, "%v bzr location, \"%s\", does not exist, err: %s", l, e.WkspcPath(), err)
	} else { // checking remote "URL" as well as possible for current VCS..
		remote := e.Remote()
		scheme := url.GetScheme(remote)
		// if we have a scheme then just see if the repo exists...
		if scheme != "" {
			_, err = exec.Command("bzr", "info", remote).CombinedOutput()
			if err == nil {
				path = remote
			}
		} else {
			vcsSchemes := e.Schemes()
			for _, scheme = range vcsSchemes {
				_, err = exec.Command("bzr", "info", scheme+"://"+remote).CombinedOutput()
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
	if loc, err := e.Exists(Wkspc); err == nil && loc != "" && remote == "" {
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

// SetDefaultBzrSchemes allows one to override the default ordering
// and set of bzr remote URL schemes to try for any remote that has
// no scheme provided, defaults to Go core list for now.
func SetDefaultBzrSchemes(schemes []string) {
	if schemes == nil {
		defaultBzrSchemes = []string{"https", "http", "bzr", "bzr+ssh"}
	} else {
		defaultBzrSchemes = schemes
	}
}
