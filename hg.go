package vcs

import (
	"os"
	"os/exec"
	"strings"
	"regexp"

    "github.com/dvln/util/dir"
)

var hgDetectURL = regexp.MustCompile("default = (?P<foo>.+)\n")

// HgGet is used to perform an initial clone of a repository.
func HgGet(g Getter, rev ...Rev) (string, error) {
	var output string
	var err error
	if rev == nil || ( rev != nil && rev[0] == "" ) {
		output, err = run("hg", "clone", "-U", g.Remote(), g.WkspcPath())
	} else {
		output, err = run("hg", "clone", "-u", string(rev[0]), "-U", g.Remote(), g.WkspcPath())
	}
	return output, err
}

// HgUpdate performs a Mercurial pull (like git fetch) + mercurial update (like git pull)
// into the workspace checkout.  Note that one can optionally identify a specific rev
// to update/merge to.  It will return any output of the cmd and an error that occurs.
// Note that there will be a pull and a merge class of functionality in dvln but
// pull is likely Mercurial pull (ie: git fetch) and merge is similar to git/hg,
// whereas update is like a fetch/merge in git or pull/upd(/merge) in hg.
func HgUpdate(u Updater, rev ...Rev) (string, error) {
	//FIXME: erik: should support a "date:<datestr>" class of rev,
	//       if that is passed in use "-d <date>" for update, so should
	//       all other VCS's that can support it... other option is to
	//       switch to a *Revision as the param but a special revision
	//       that only has *1* revision set (raw, tags, branches, semver,
	//       time)... and a silly routine to get that rev (then no need
	//       to mark up the 'Rev' type (which is a string), but a strong
	//       need to pass in the right thing of course if that is done. ;)
	output, err := runFromWkspcDir(u.WkspcPath(), "hg", "pull")
	if err != nil {
		return output, err
	}
	var updOut string
	if rev == nil || ( rev != nil && rev[0] == "" ) {
		updOut, err = runFromWkspcDir(u.WkspcPath(), "hg", "update")
	} else {
		updOut, err = runFromWkspcDir(u.WkspcPath(), "hg", "update", "-r", string(rev[0]))
	}
	output = output + updOut
	return output, err
}

// HgRevSet sets the wkspc revision of a pkg currently checked out via Hg.
// Note that a single specific revision must be given vs a generic
// Revision structure (since it may have <N> different valid rev's
// that reference the revision, this one decides exactly the one
// the client wishes to "set" or checkout in the wkspc).
func HgRevSet(r RevSetter, rev Rev) (string, error) {
	return runFromWkspcDir(r.WkspcPath(), "hg", "update", string(rev))
}

// HgRevRead retrieves the current version.
func HgRevRead(r RevReader, scope ...ReadScope) (*Revision, string, error) {
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
		output, err = exec.Command("hg", "identify").CombinedOutput()
		if err != nil {
			return nil, string(output), err
		}
		parts := strings.SplitN(string(output), " ", 2)
		sha := strings.TrimSpace(parts[0])
		rev.SetCore(Rev(sha))
	} else {
		//FIXME: erik: implement more extensive hg data grab
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
		output, err = exec.Command("hg", "identify").CombinedOutput()
		if err != nil {
			return nil, string(output), err
		}
		parts := strings.SplitN(string(output), " ", 2)
		sha := strings.TrimSpace(parts[0])
		rev.SetCore(Rev(sha))
	}
	return rev, string(output), err
}

// HgExists verifies the wkspc or remote location is a Hg repo.
func HgExists(e Existence, l Location) (bool, error) {
	var err error
	if l == Wkspc {
		if there, err := dir.Exists(e.WkspcPath() + "/.hg"); there && err == nil {
			return true, nil
		}
		//FIXME: erik: if err != nil should use something like:
		//       out.WrapErrf(ErrNoExists, #, "%v hg location, \"%s\", does not exist, err: %s", l, e.WkspcPath(), err)
	} else {
		//FIXME: erik: need to actually check if remote repo exists ;)
		// should use this "ErrNoExist" from repo.go if doesn't exist
		return true, nil
	}
	return false, err
}

// HgCheckRemote  attempts to take a remote string (URL) and validate
// it against any local repo and try and set it when it is empty.  Returns:
// - string: this is the new remote (current remote returned if no new remote)
// - string: output of the Bzr command to try and determine the remote
// - error: non-nil if an error occurred
func HgCheckRemote (e Existence, remote string) (string, string, error) {
	// Make sure the wkspc Hg repo is configured the same as the remote when
	// A remote value was passed in.
	var outStr string
	if exists, err := e.Exists(Wkspc); err == nil && exists {
		// An Hg repo was found so test that the URL there matches
		// the repo passed in here.
		oldDir, err := os.Getwd()
		if err != nil {
			return remote, "", err
		}
		err = os.Chdir(e.WkspcPath())
		if err != nil {
			return remote, "", err
		}
		defer os.Chdir(oldDir)
		output, err := exec.Command("hg", "paths").CombinedOutput()
		if err != nil {
			return remote, string(output), err
		}

		outStr = string(output)
		m := hgDetectURL.FindStringSubmatch(outStr)
		//FIXME: erik: added that remote != "", think it's needed, check
		if remote != "" && m[1] != "" && m[1] != remote {
			return remote, outStr, ErrWrongRemote
		}

		// If no remote was passed in but one is configured for the locally
		// checked out Hg repo use that one.
		if remote == "" && m[1] != "" {
			return m[1], outStr, nil
		}
	}
	return remote, outStr, nil
}

