package vcs

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"regexp"

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

// SvnExists verifies the wkspc or remote location is of the SVN type,
// returns where it was found ("" if not found) and any error
func SvnExists(e Existence, l Location) (string, error) {
	var err error
	path := ""
	if l == Wkspc {
		if exists, err := dir.Exists(e.WkspcPath() + "/.svn"); exists && err == nil {
			return e.WkspcPath(), nil
		}
		//FIXME: erik: if err != nil should use something like:
		//       out.WrapErrf(ErrNoExists, #, "%v SVN location, \"%s\", does not exist, err: %s", l, e.WkspcPath(), err)
	} else { // checking remote "URL" as well as possible for current VCS..
		remote := e.Remote()
		scheme := url.GetScheme(remote)
		// if we have a scheme then just see if the repo exists...
		if scheme != "" {
			_, err = exec.Command("svn", "info", remote).CombinedOutput()
			if err == nil {
				path = remote
			}
		} else {
			vcsSchemes := e.Schemes()
			for _, scheme = range vcsSchemes {
				_, err = exec.Command("svn", "info", scheme + "://" + remote).CombinedOutput()
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

// SvnCheckRemote attempts to take a remote string (URL) and validate
// it against any local wkspc checkout, tries to set it when empty.  Returns:
// - string: this is the new remote (current remote returned if no new remote)
// - string: output of the Bzr command to try and determine the remote
// - error: non-nil if an error occurred
func SvnCheckRemote (e Existence, remote string) (string, string, error) {
	// Make sure the wkspc Svn repo is configured the same as the remote when
	// A remote value was passed in.
	var outStr string
	if loc, err := e.Exists(Wkspc); err == nil && loc != "" {
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
