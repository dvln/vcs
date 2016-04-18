package vcs

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/dvln/util/dir"
	"github.com/dvln/util/url"
)

var svnDetectURL = regexp.MustCompile("URL: (?P<foo>.+)\n")
var defaultSvnSchemes []string

func init() {
	SetDefaultSvnSchemes(nil)
}

// SvnGet is used to perform an initial checkout of a repository.
// Note, because SVN isn't distributed this is a checkout without
// a clone.  One can checkout an optionally passed in revision.
func SvnGet(g *SvnGetter, rev ...Rev) (Resulter, error) {
	results := newResults()
	var result *Result
	var err error
	if rev == nil || (rev != nil && rev[0] == "") {
		result, err = run("svn", "checkout", g.Remote(), g.LocalRepoPath())
	} else {
		result, err = run("svn", "checkout", "-r", string(rev[0]), g.Remote(), g.LocalRepoPath())
	}
	results.add(result)
	return results, err
}

// SvnUpdate performs an SVN update to an existing checkout (ie: a merge).
func SvnUpdate(u *SvnUpdater, rev ...Rev) (Resulter, error) {
	results := newResults()
	var result *Result
	var err error
	if rev == nil || (rev != nil && rev[0] == "") {
		result, err = runFromLocalRepoDir(u.LocalRepoPath(), "svn", "update")
	} else {
		result, err = runFromLocalRepoDir(u.LocalRepoPath(), "svn", "update", "-r", string(rev[0]))
	}
	results.add(result)
	return results, err
}

// SvnRevSet sets the local repo rev of a pkg currently checked out via
// Svn.  Note that a single specific revision must be given (vs a generic
// Revision structure as such a struct may have <N> different valid rev's
// that reference the revision).  The results (if any) and any error
// is returned from the svn update run.
func SvnRevSet(r RevSetter, rev Rev) (Resulter, error) {
	results := newResults()
	result, err := runFromLocalRepoDir(r.LocalRepoPath(), "svn", "update", "-r", string(rev))
	results.add(result)
	return results, err
}

// SvnRevRead retrieves the given or current local repo rev.  A Revision struct
// pointer is returned (how filled out depends upon if the read is just the
// basic core/raw VCS revision or full data for the given VCS which will
// include tags, branches, timestamp info, author/committer, date, comment).
// Note: this reads one version but that could be expanded to take <N>
// revisions or a range, eg SvnRevRead(reader, <scope>, rev1, "..", rev2),
// without changing this methods params or return signature (but code
// changes  would be needed)
func SvnRevRead(r RevReader, scope ReadScope, vcsRev ...Rev) ([]Revisioner, Resulter, error) {
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
	if scope == CoreRev {
		// client just wants the core/base VCS revision only..
		//FIXME: based on SVN docs this doesn't seem like it
		//       can handle the various output formats correctly with
		//       modifiers like "<rev#>M" or "<rev#>S" or "<rev#>:<rev#>",
		//       see svnversion -h, perhaps update
		//       Also: specificRev isn't implemented for SVN, is it possible?
		if specificRev != "" {
			return nil, nil, fmt.Errorf("Reading specified revision, %s, not supported by SVN", specificRev)
		}
		var result *Result
		result, err = run("svnversion", ".")
		results.add(result)
		if err != nil {
			return nil, results, err
		}
		rev.SetCore(Rev(strings.TrimSpace(result.output)))
		revs = append(revs, rev)
	} else {
		//FIXME: this needs to add more data if possible for SVN
		//       Also: specificRev isn't implemented for SVN, is it possible?
		if specificRev != "" {
			return nil, nil, fmt.Errorf("Reading specified revision, %s, not supported by SVN", specificRev)
		}
		var result *Result
		result, err = run("svnversion", ".")
		results.add(result)
		if err != nil {
			return nil, results, err
		}
		rev.SetCore(Rev(strings.TrimSpace(result.output)))
		revs = append(revs, rev)
	}
	return revs, results, err
}

// SvnExists verifies the local repo or remote location is of the SVN type,
// returns where it was found ("" if not found) and any error
func SvnExists(e Existence, l Location) (string, Resulter, error) {
	results := newResults()
	var err error
	path := ""
	if l == LocalPath {
		if exists, err := dir.Exists(e.LocalRepoPath() + "/.svn"); exists && err == nil {
			return e.LocalRepoPath(), nil, nil
		}
	} else { // checking remote "URL" as well as possible for current VCS..
		remote := e.Remote()
		scheme := url.GetScheme(remote)
		// if we have a scheme then just see if the repo exists...
		if scheme != "" {
			var result *Result
			result, err = run("svn", "info", remote)
			results.add(result)
			if err == nil {
				path = remote
			}
		} else {
			vcsSchemes := e.Schemes()
			for _, scheme = range vcsSchemes {
				var result *Result
				result, err = run("svn", "info", scheme+"://"+remote)
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

// SvnCheckRemote attempts to take a remote string (URL) and validate
// it against any local repo checkout, tries to set it when empty.  Returns:
// - string: this is the new remote (current remote returned if no new remote)
// - string: output of the Bzr command to try and determine the remote
// - error: non-nil if an error occurred
func SvnCheckRemote(e Existence, remote string) (string, Resulter, error) {
	// Make sure the local Svn repo is configured the same as the remote when
	// A remote value was passed in.
	results := newResults()
	var outStr string
	if loc, existResults, err := e.Exists(LocalPath); err == nil && loc != "" {
		if existResults != nil {
			for _, existResult := range existResults.All() {
				results.add(existResult)
			}
		}
		// An SVN repo was found so test that the URL there matches
		// the repo passed in here.
		var result *Result
		result, err := run("svn", "info", e.LocalRepoPath())
		results.add(result)
		outStr = result.output
		if err != nil {
			return remote, results, err
		}

		m := svnDetectURL.FindStringSubmatch(outStr)
		if remote != "" && m[1] != "" && m[1] != remote {
			return remote, results, ErrWrongRemote
		}

		// If no remote was passed in but one is configured for the locally
		// checked out Svn repo use that one.
		if remote == "" && m[1] != "" {
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

// SetDefaultSvnSchemes allows one to override the default ordering
// and set of svn remote URL schemes to try for any remote that has
// no scheme provided, defaults to Go core list for now.
func SetDefaultSvnSchemes(schemes []string) {
	if schemes == nil {
		defaultSvnSchemes = []string{"https", "http", "svn", "svn+ssh"}
	} else {
		defaultSvnSchemes = schemes
	}
}
