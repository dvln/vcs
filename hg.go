package vcs

import (
	"os"
	"regexp"
	"strings"

	"github.com/dvln/util/dir"
	"github.com/dvln/util/url"
)

var hgDetectURL = regexp.MustCompile("default = (?P<foo>.+)\n")

var defaultHgSchemes []string

// set up default hg remote URL schemes and a search order (for any remote
// that doesn't have a full URL), eg: https, http, ssh
func init() {
	SetDefaultHgSchemes(nil)
}

// HgGet is used to perform an initial clone of a repository.
func HgGet(g *HgGetter, rev ...Rev) (Resulter, error) {
	results := newResults()
	var result *Result
	var err error
	if rev == nil || (rev != nil && rev[0] == "") {
		result, err = run("hg", "clone", "-U", g.Remote(), g.LocalRepoPath())
	} else {
		result, err = run("hg", "clone", "-u", string(rev[0]), "-U", g.Remote(), g.LocalRepoPath())
	}
	results.add(result)
	return results, err
}

// HgUpdate performs a Mercurial pull (like git fetch) + mercurial update (like git pull)
// into the workspace checkout.  Note that one can optionally identify a specific rev
// to update/merge to.  It will return any output of the cmd and an error that occurs.
// Note that there will be a pull and a merge class of functionality in dvln but
// pull is likely Mercurial pull (ie: git fetch) and merge is similar to git/hg,
// whereas update is like a fetch/merge in git or pull/upd(/merge) in hg.
func HgUpdate(u *HgUpdater, rev ...Rev) (Resulter, error) {
	//FIXME: should support a "date:<datestr>" class of rev,
	//       if that is passed in use "-d <date>" for update, so should
	//       all other VCS's that can support it... other option is to
	//       switch to a *Revision as the param but a special revision
	//       that only has *1* revision set (raw, tags, branches, semver,
	//       time)... and a silly routine to get that rev (then no need
	//       to mark up the 'Rev' type (which is a string), but a strong
	//       need to pass in the right thing of course if that is done. ;)
	results := newResults()
	result, err := runFromLocalRepoDir(u.LocalRepoPath(), "hg", "pull")
	results.add(result)
	if err != nil {
		return results, err
	}
	var updResult *Result
	if rev == nil || (rev != nil && rev[0] == "") {
		updResult, err = runFromLocalRepoDir(u.LocalRepoPath(), "hg", "update")
	} else {
		updResult, err = runFromLocalRepoDir(u.LocalRepoPath(), "hg", "update", "-r", string(rev[0]))
	}
	results.add(updResult)
	return results, err
}

// HgRevSet sets the local repo rev of a pkg currently checked out via Hg.
// Note that a single specific revision must be given vs a generic
// Revision structure (since it may have <N> different valid rev's
// that reference the revision, this one decides exactly the one
// the client wishes to "set" or checkout in the local repo).
func HgRevSet(r RevSetter, rev Rev) (Resulter, error) {
	results := newResults()
	if rev == "" {
		result, err := runFromLocalRepoDir(r.LocalRepoPath(), "hg", "update")
		results.add(result)
		return results, err
	}
	result, err := runFromLocalRepoDir(r.LocalRepoPath(), "hg", "update", "-r", string(rev))
	results.add(result)
	return results, err
}

// HgRevRead retrieves the given or current local repo rev.  A Revision struct
// pointer is returned (how filled out depends upon if the read is just the
// basic core/raw VCS revision or full data for the given VCS which will
// include tags, branches, timestamp info, author/committer, date, comment).
// Note: this reads one version but that could be expanded to take <N>
// revisions or a range, eg HgRevRead(reader, <scope>, rev1, "..", rev2),
// without changing this methods params or return signature (but code
// changes  would be needed)
func HgRevRead(r RevReader, scope ReadScope, vcsRev ...Rev) ([]Revisioner, Resulter, error) {
	results := newResults()
	oldDir, err := os.Getwd()
	if err != nil {
		return nil, nil, err
	}
	specificRev := ""
	if vcsRev != nil && vcsRev[0] != "" {
		specificRev = string(vcsRev[0])
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
		var result *Result
		if specificRev != "" {
			result, err = run("hg", "identify", "-r", specificRev)
		} else {
			result, err = run("hg", "identify")
		}
		results.add(result)
		if err != nil {
			return nil, results, err
		}
		parts := strings.SplitN(result.Output, " ", 2)
		sha := strings.TrimSpace(parts[0])
		rev.SetCore(Rev(sha))
		revs = append(revs, rev)
	} else {
		//FIXME: implement more extensive hg data grab
		//       if the client has asked for extra data (vs speed)
		/* Here is how to get the details, if no "branch:" it's default
		   141 [brady-air]/Users/brady/vcs/mercurial-repo: hg log -l 1
		   changeset:   26211:ea489d94e1dc
		   bookmark:    @
		   tag:         tip
		   user:        Gregory Szorc <gregory.szorc@gmail.com>
		   date:        Sat Aug 22 17:08:37 2015 -0700
		   summary:     hgweb: assign ctype to requestcontext

		   142 [brady-air]/Users/brady/vcs/mercurial-repo: hg log -l 1 -r 3.5.1
		   changeset:   26120:1a45e49a6bed
		   branch:      stable
		   tag:         3.5.1
		   user:        Matt Mackall <mpm@selenic.com>
		   date:        Tue Sep 01 16:08:07 2015 -0500
		   summary:     hgweb: fix trust of templates path (BC)

		   143 [brady-air]/Users/brady/vcs/mercurial-repo: hg identify
		   ea489d94e1dc tip @
		   144 [brady-air]/Users/brady/vcs/mercurial-repo: hg identify -r 3.5.1
		   1a45e49a6bed (stable) 3.5.1
		   145 [brady-air]/Users/brady/vcs/mercurial-repo: hg identify -i -b -t -r 3.5.1
		   1a45e49a6bed stable 3.5.1
		   146 [brady-air]/Users/brady/vcs/mercurial-repo: hg identify -i -b -t
		   ea489d94e1dc default tip
		   147 [brady-air]/Users/brady/vcs/mercurial-repo:
		*/
		var result *Result
		if specificRev != "" {
			result, err = run("hg", "identify", "-r", specificRev)
		} else {
			result, err = run("hg", "identify")
		}
		results.add(result)
		if err != nil {
			return nil, results, err
		}
		parts := strings.SplitN(result.Output, " ", 2)
		sha := strings.TrimSpace(parts[0])
		rev.SetCore(Rev(sha))
		revs = append(revs, rev)
	}
	return revs, results, err
}

// HgExists verifies the local repo or remote location is a Hg repo,
// returns where it was found ("" if not found), a resulter (cmds
// run and their output to accomplish task) and and any error
func HgExists(e Existence, l Location) (string, Resulter, error) {
	results := newResults()
	var err error
	path := ""
	if l == LocalPath {
		if exists, err := dir.Exists(e.LocalRepoPath() + "/.hg"); exists && err == nil {
			return e.LocalRepoPath(), nil, nil
		}
	} else { // checking remote "URL" as well as possible for current VCS..
		remote := e.Remote()
		scheme := url.GetScheme(remote)
		// if we have a scheme then just see if the repo exists...
		if scheme != "" {
			var result *Result
			result, err = run("hg", "identify", remote)
			results.add(result)
			if err == nil {
				path = remote
			}
		} else {
			vcsSchemes := e.Schemes()
			for _, scheme = range vcsSchemes {
				var result *Result
				result, err = run("hg", "identify", scheme+"://"+remote)
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

// HgCheckRemote  attempts to take a remote string (URL) and validate
// it against any local repo and try and set it when it is empty.  Returns:
// - string: this is the new remote (current remote returned if no new remote)
// - Resulter: cmd(s) run and output of the Hg commands
// - error: non-nil if an error occurred
func HgCheckRemote(e Existence, remote string) (string, Resulter, error) {
	results := newResults()
	// Make sure the local Hg repo is configured the same as the remote when
	// A remote value was passed in.
	var outStr string
	if loc, existResults, err := e.Exists(LocalPath); err == nil && loc != "" {
		if existResults != nil {
			for _, existResult := range existResults.All() {
				results.add(existResult)
			}
		}
		// An Hg repo was found so test that the URL there matches
		// the repo passed in here.
		oldDir, err := os.Getwd()
		if err != nil {
			return remote, nil, err
		}
		err = os.Chdir(e.LocalRepoPath())
		if err != nil {
			return remote, nil, err
		}
		defer os.Chdir(oldDir)
		result, err := run("hg", "paths")
		results.add(result)
		if err != nil {
			return remote, results, err
		}

		outStr = result.Output
		m := hgDetectURL.FindStringSubmatch(outStr)
		//FIXME: added that remote != "", think it's needed, check
		if remote != "" && m[1] != "" && m[1] != remote {
			return remote, results, ErrWrongRemote
		}

		// If no remote was passed in but one is configured for the locally
		// checked out Hg repo use that one.
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

// SetDefaultHgSchemes allows one to override the default ordering
// and set of hg remote URL schemes to try for any remote that has
// no scheme provided, defaults to Go core list for now.
func SetDefaultHgSchemes(schemes []string) {
	if schemes == nil {
		defaultHgSchemes = []string{"https", "http", "ssh"}
	} else {
		defaultHgSchemes = schemes
	}
}
