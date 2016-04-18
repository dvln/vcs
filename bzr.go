package vcs

import (
	"os"
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
func BzrGet(g *BzrGetter, rev ...Rev) (Resulter, error) {
	results := newResults()
	var err error
	var result *Result
	if rev == nil || (rev != nil && rev[0] == "") {
		result, err = run("bzr", "branch", g.Remote(), g.LocalRepoPath())
	} else {
		result, err = run("bzr", "branch", "-r", string(rev[0]), g.Remote(), g.LocalRepoPath())
	}
	results.add(result)
	return results, err
}

// BzrUpdate performs a Bzr pull and update to an existing checkout.
func BzrUpdate(u *BzrUpdater, rev ...Rev) (Resulter, error) {
	results := newResults()
	result, err := runFromLocalRepoDir(u.LocalRepoPath(), "bzr", "pull")
	results.add(result)
	if err != nil {
		return results, err
	}
	var updResult *Result
	if rev == nil || (rev != nil && rev[0] == "") {
		updResult, err = runFromLocalRepoDir(u.LocalRepoPath(), "bzr", "update")
	} else {
		updResult, err = runFromLocalRepoDir(u.LocalRepoPath(), "bzr", "update", "-r", string(rev[0]))
	}
	results.add(updResult)
	return results, err
}

// BzrRevSet sets the local repo rev of a pkg currently checked out via Bzr.
// Note that a single specific revision must be given (vs a generic
// Revision structure as such a struct may have <N> different valid rev's
// that reference the revision).  The raw cmd results (if any) and any
// error is returned from the bzr update run.
func BzrRevSet(r RevSetter, rev Rev) (Resulter, error) {
	results := newResults()
	result, err := runFromLocalRepoDir(r.LocalRepoPath(), "bzr", "update", "-r", string(rev))
	results.add(result)
	return results, err
}

// BzrRevRead retrieves the given or current local repo rev.  A Revision struct
// pointer is returned (how filled out depends upon if the read is just the
// basic core/raw VCS revision or full data for the given VCS which will
// include tags, branches, timestamp info, author/committer, date, comment).
// Note: this reads one version but that could be expanded to take <N>
// revisions or a range, eg BzrRevRead(reader, <scope>, rev1, "..", rev2),
// without changing this methods params or return signature (but code
// changes  would be needed)
func BzrRevRead(r RevReader, scope ReadScope, vcsRev ...Rev) ([]Revisioner, Resulter, error) {
	results := newResults()
	specificRev := ""
	if vcsRev != nil && vcsRev[0] != "" {
		specificRev = string(vcsRev[0])
	}
	oldDir, err := os.Getwd()
	if err != nil {
		return nil, nil, err
	}
	err = os.Chdir(r.LocalRepoPath())
	if err != nil {
		return nil, nil, err
	}
	defer os.Chdir(oldDir)

	rev := &Revision{}
	var revs []Revisioner
	var result *Result
	if scope == CoreRev {
		// client just wants the core/base VCS revision only..
		if specificRev != "" {
			result, err = run("bzr", "revno", "-r", specificRev)
		} else {
			result, err = run("bzr", "revno", "--tree")
		}
		results.add(result)
		if err != nil {
			return nil, results, err
		}
		rev.SetCore(Rev(strings.TrimSpace(string(result.Output))))
		revs = append(revs, rev)
	} else {
		//FIXME: get additional data about the version if possible (fix this)
		if specificRev != "" {
			result, err = run("bzr", "revno", "-r", specificRev)
		} else {
			result, err = run("bzr", "revno", "--tree")
		}
		results.add(result)
		if err != nil {
			return nil, results, err
		}
		rev.SetCore(Rev(strings.TrimSpace(result.Output)))
		revs = append(revs, rev)
	}
	return revs, results, err
}

// BzrExists verifies the local repo or remote location is of the Bzr repo type,
// returns where it was found ("" if not found) and any error
func BzrExists(e Existence, l Location) (string, Resulter, error) {
	results := newResults()
	var err error
	path := ""
	if l == LocalPath {
		if exists, err := dir.Exists(e.LocalRepoPath() + "/.bzr"); exists && err == nil {
			return e.LocalRepoPath(), nil, nil
		}
		//FIXME: if err != nil should use something like:
		//       out.WrapErrf(ErrNoExists, #, "%v bzr location, \"%s\", does not exist, err: %s", l, e.LocalRepoPath(), err)
	} else { // checking remote "URL" as well as possible for current VCS..
		remote := e.Remote()
		scheme := url.GetScheme(remote)
		// if we have a scheme then just see if the repo exists...
		if scheme != "" {
			var result *Result
			result, err = run("bzr", "info", remote)
			results.add(result)
			if err == nil {
				path = remote
			}
		} else {
			vcsSchemes := e.Schemes()
			for _, scheme = range vcsSchemes {
				var result *Result
				result, err = run("bzr", "info", scheme+"://"+remote)
				results.add(result)
				if err == nil {
					path = scheme + "://" + remote
					break
				}
			}
		}
		if err == nil {
			return path, results, nil
		}
	}
	return path, results, err
}

// BzrCheckRemote attempts to take a remote string (URL) and validate
// it (although with Bzr that doesn't work well) and set it if it is not
// currently set (this happens if a local clone exists only).  Returns:
// - string: this is the new remote (current remote returned if no new remote)
// - string: output of the Bzr command to try and determine the remote
// - error: non-nil if an error occurred
func BzrCheckRemote(e Existence, remote string) (string, Resulter, error) {
	// With the other VCS we can check if the endpoint locally is different
	// from the one configured internally. But, with Bzr you can't. For example,
	// if you do `bzr branch https://launchpad.net/govcstestbzrrepo` and then
	// use `bzr info` to get the parent branch you'll find it set to
	// http://bazaar.launchpad.net/~mattfarina/govcstestbzrrepo/trunk/. Notice
	// the change from https to http and the path chance.
	// Here we set the remote to be the local one if none is passed in.
	results := newResults()
	var outStr string
	if loc, existResults, err := e.Exists(LocalPath); err == nil && loc != "" && remote == "" {
		if existResults != nil {
			for _, existResult := range existResults.All() {
				results.add(existResult)
			}
		}
		oldDir, err := os.Getwd()
		if err != nil {
			return remote, nil, err
		}
		err = os.Chdir(e.LocalRepoPath())
		if err != nil {
			return remote, nil, err
		}
		defer os.Chdir(oldDir)
		result, err := run("bzr", "info")
		results.add(result)
		if err != nil {
			return remote, results, err
		}
		outStr = string(result.Output)
		m := bzrDetectURL.FindStringSubmatch(outStr)

		// If no remote was passed in but one is configured for the locally
		// checked out Bzr VCS pkg (repo) use that one.
		if m[1] != "" {
			return m[1], results, nil
		}
	} else if err != nil {
		if existResults != nil {
			for _, existResult := range existResults.All() {
				results.add(existResult)
			}
		}
	}
	return remote, results, nil
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
