package vcs

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/dvln/out"
	"github.com/dvln/util/dir"
	"github.com/dvln/util/file"
	"github.com/dvln/util/url"
)

var defaultGitSchemes []string
var refsRegex = regexp.MustCompile(`^refs/heads/(.*)$`)

// RemoteMode describes how remote URL and checking/updating works
type RemoteMode string

// Remote URL/name checking behavior
const (
	// CheckRemote indicates to just validate remote URL vs remote name
	CheckRemote RemoteMode = "check"
 	// UpdateRemote says to force remote name point to given URL
	UpdateRemote RemoteMode = "update"
)

func init() {
	SetDefaultGitSchemes(nil)
}

// GitGet is used to perform an initial clone of a repository, optionally
// can check out a rev, params:
//	g (*GitGetter): the getter data we need to run the pull
//	rev (Rev): optional; revision to checkout after getting the clone
// Returns results (vcs cmds run, output) and any error that may have occurred
func GitGet(g *GitGetter, rev ...Rev) (Resulter, error) {
	results := newResults()
	mirrorStr := ""
	if g.mirror { // if in mirror clone mode add in --mirror
		mirrorStr = "--mirror"
	}
	result, err := run("git", "clone", mirrorStr, g.Remote(), g.LocalRepoPath())
	results.add(result)
	if err != nil && rev != nil {
		// Be careful to append more results from cmds run in RevSet
		var setResults Resulter
		setResults, err = g.RevSet(rev[0])
		if setResults != nil {
			for _, revResult := range setResults.All() {
				results.add(revResult)
			}
		}
	}
	return results, err
}

// gitUpdateRefs is fired if GitUpdate() gets specific refs to operate
// on... meaning fetch or delete ops (at this point).  Params:
//	u (*GitUpdater): has all the data we need to run the update
// Returns results (vcs cmds run, output) and any error that may have occurred
func gitUpdateRefs(u *GitUpdater) (Resulter, error) {
	var err error
	results := newResults()
	runOpt := "-C"
	runDir := u.LocalRepoPath()
	for ref, refOp := range u.refs {
		var result *Result
		switch refOp {
		case RefDelete:
			result, err = run("git", runOpt, runDir, "update-ref", "-d", ref)
			results.add(result)
		case RefFetch:
			if u.mirror { // request is to mirror refs exactly, do so
				refSpec := fmt.Sprintf("+%s:%s", ref, ref)
				result, err = run("git", runOpt, runDir, "fetch", u.RemoteRepoName(), refSpec)
			} else { // normal fetch requested, heads remapped, all else comes in "as-is"
				m := refsRegex.FindStringSubmatch(ref) // look for refs/heads/<name> refs
				if m[1] != "" {                        // if it was a refs/heads then map it:
					remoteRef := fmt.Sprintf("refs/remotes/%s/%s", u.RemoteRepoName(), m[1])
					refSpec := fmt.Sprintf("+%s:%s", ref, remoteRef)
					result, err = run("git", runOpt, runDir, "fetch", u.RemoteRepoName(), refSpec)
				} else { // bring in tags/etc under the same namespace
					refSpec := fmt.Sprintf("+%s:%s", ref, ref)
					result, err = run("git", runOpt, runDir, "fetch", u.RemoteRepoName(), refSpec)
				}
			}
			results.add(result)
		default:
			err = out.NewErrf(4502, "Update refs: invalid ref operation given \"%v\", clone: %s", refOp, u.LocalRepoPath())
		}
	}
	return results, err
}

// GitUpdate performs a git fetch and merge to an existing checkout (ie:
// a git pull).  Params:
//	u (*GitUpdater): git upd struct, gives kind of update needed, stores cmds run
//	rev (Rev): optional; revision to update to (if given only 1 used)
// Returns results (vcs cmds run, output) and any error that may have occurred
func GitUpdate(u *GitUpdater, rev ...Rev) (Resulter, error) {
	// Perform required fetches optionally with pulls as well as handling
	// more specific fetches on single refs (or deletion of refs)... has
	// some handling of mirror/bare clones vs local clones and for std
	// clones can do rebase type pulls (if that section of the routine is
	// reached).
	var err error
	runOpt := "-C"
	runDir := u.LocalRepoPath()
	if u.refs != nil {
		return gitUpdateRefs(u)
	}
	results := newResults()
	var result *Result
	if u.mirror {
		result, err = run("git", runOpt, runDir, "remote", "update", "--prune", u.RemoteRepoName())
	} else {
		result, err = run("git", runOpt, runDir, "fetch", u.RemoteRepoName())
	}
	results.add(result)
	if err != nil {
		return results, err
	}

	bareRepo := false
	gitDir, workTree, err := findGitDirs(runDir)
	if err != nil {
		return nil, err
	}
	if gitDir == runDir && workTree == "" {
		bareRepo = true
	}
	if !u.mirror && !bareRepo { // if not a mirror and a regular clone
		// Try and run a git pull to do the merge|rebase op
		rebaseStr := ""
		switch u.rebase {
		case RebaseFalse:
			rebaseStr = "--rebase=false"
		case RebaseTrue:
			rebaseStr = "--rebase=true"
		case RebasePreserve:
			rebaseStr = "--rebase=preserve"
		default: // likely RebaseUser, meaning don't provide any rebase opt
		}
		var pullResult *Result
		if rev == nil || (rev != nil && rev[0] == "") {
			pullResult, err = run("git", runOpt, runDir, "pull", rebaseStr, u.RemoteRepoName())
		} else { // if user asks for a specific version on pull, use that
			pullResult, err = run("git", runOpt, runDir, "pull", rebaseStr, u.RemoteRepoName(), string(rev[0]))
		}
		results.add(pullResult)
	}
	return results, err
}

// GitRevSet sets the local repo rev of a pkg currently checked out via Git.
// Note that a single specific revision must be given vs a generic
// Revision structure (since it may have <N> different valid rev's
// that reference the revision, this one decides exactly the one
// the client wishes to "set" or checkout in the repo path).
func GitRevSet(r RevSetter, rev Rev) (Resulter, error) {
	runOpt := "-C"
	runDir := r.LocalRepoPath()
	results := newResults()
	result, err := run("git", runOpt, runDir, "checkout", string(rev))
	results.add(result)
	return results, err
}

// GitRevRead retrieves the given or current local repo rev.  A Revision struct
// pointer is returned (how filled out depends upon if the read is just the
// basic core/raw VCS revision or full data for the given VCS which will
// include tags, branches, timestamp info, author/committer, date, comment).
// Note: this reads one version but that could be expanded to take <N>
// revisions or a range, eg GitRevRead(reader, <scope>, rev1, "..", rev2),
// without changing this methods params or return signature (but code
// changes  would be needed)
func GitRevRead(r RevReader, scope ReadScope, vcsRev ...Rev) ([]Revisioner, Resulter, error) {
	results := newResults()
	runOpt := "-C"
	runDir := r.LocalRepoPath()
	specificRev := ""
	if vcsRev != nil && vcsRev[0] != "" {
		specificRev = string(vcsRev[0])
	}
	rev := &Revision{}
	var revs []Revisioner
	var err error
	var result *Result
	if scope == CoreRev {
		// client just wants the core/base VCS revision only..
		if specificRev != "" {
			result, err = run("git", runOpt, runDir, "log", "-1", "--format=%H", specificRev)
		} else {
			result, err = run("git", runOpt, runDir, "log", "-1", "--format=%H")
		}
		results.add(result)
		if err != nil {
			return nil, results, err
		}
		rev.SetCore(Rev(strings.TrimSpace(result.Output)))
		revs = append(revs, rev)
	} else {
		//FIXME: correct the full data one to run something like this:
		//% git log -1 --format='%H [%cD]%d'
		//a862506d017d643091368d53128447d032a03f54 [Thu, 11 Sep 2014 17:45:32 -0700] (HEAD -> topic, tag: main/7353, tag: acme__main__new__1410482753, origin/main, origin/HEAD)
		//should also add author+authorid+committer+committerid and then add in the
		//revision comment on the line following that data
		if specificRev != "" {
			result, err = run("git", runOpt, runDir, "log", "-1", "--format=%H", specificRev)
		} else {
			result, err = run("git", runOpt, runDir, "log", "-1", "--format=%H")
		}
		results.add(result)
		if err != nil {
			return nil, results, err
		}
		rev.SetCore(Rev(strings.TrimSpace(result.Output)))
		revs = append(revs, rev)
	}
	return revs, results, nil
}

// GitExists verifies the local repo or remote location is a Git repo,
// returns where it was found (or "" if not found), the results
// of any git cmds run (cmds and related output) and any error.
// Note that if no git cmds run then Resulter won't have any data
// (which occurs if the git repo is local)
func GitExists(e Existence, l Location) (string, Resulter, error) {
	results := newResults()
	var err error
	path := ""
	if l == LocalPath {
		_, _, err = findGitDirs(e.LocalRepoPath()) // see if git clone there
		if err == nil {
			return e.LocalRepoPath(), nil, nil // it's a local git clone, success
		}
	} else { // checking remote "URL" as well as possible for current VCS..
		remote := e.Remote()
		scheme := url.GetScheme(remote)
		if scheme != "" { // if we have a scheme then see if the repo exists...
			var result *Result
			result, err = run("git", "ls-remote", remote)
			results.add(result)
			if err == nil {
				path = remote
			}
		} else {
			vcsSchemes := e.Schemes()
			for _, scheme = range vcsSchemes {
				var result *Result
				result, err = run("git", "ls-remote", scheme+"://"+remote)
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
		err = out.WrapErrf(ErrNoExist, 4501, "Remote git location, \"%s\", does not exist, err: %s", e.LocalRepoPath(), err)
	}
	return path, results, err
}

// GitCheckRemote  attempts to take a remote string (URL) and validate
// it against any local repo and try and set it when it is empty.  Returns:
// - string: this is the new remote (current remote returned if no new remote)
// - Resulter: cmds and output of all git cmds attempted
// - error: non-nil if an error occurred
func GitCheckRemote(e Existence, remote string, mode ...RemoteMode) (string, Resulter, error) {
	// Make sure the local Git repo is configured the same as the remote when
	// a remote value was passed in, if no remote try and determine it here
	currMode := CheckRemote // default to just checking the remote
	if mode != nil && len(mode) == 1 {
		currMode = mode[0] // set it to whatever was passed in otherwise (upd|check)
	}
	results := newResults()
	var outStr string
	if loc, existResults, err := e.Exists(LocalPath); err == nil && loc != "" {
		if existResults != nil {
			for _, existResult := range existResults.All() {
				results.add(existResult)
			}
		}
		runOpt := "-C"
		runDir := loc
		remoteName := e.RemoteRepoName()
		gitString := fmt.Sprintf("remote.%s.url", remoteName)
		result, err := run("git", runOpt, runDir, "config", "--get", gitString)
		results.add(result)
		if err != nil {
			return remote, results, err
		}
		outStr = result.Output
		localRemote := strings.TrimSpace(outStr)
		if remote != "" && localRemote != remote {
			// If remote is given and it doesn't match what the remoteName
			// (eg: "origin") points to, the error if just checking and if
			// told to update instead update the remoteName's URL to 'remote'
			if currMode == UpdateRemote {
				remResult, err := run("git", runOpt, runDir, "remote", "set-url", remoteName, remote)
				results.add(remResult)
				if err == nil {
					return remote, results, nil
				}
			}
			return remote, results, ErrWrongRemote
		}

		// If no remote was passed in but one is configured for the locally
		// checked out Git repo use that one.
		if remote == "" && localRemote != "" {
			return localRemote, results, nil
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

// SetDefaultGitSchemes allows one to override the default ordering
// and set of git remote URL schemes to try for any remote that has
// no scheme provided, defaults to Go core list for now.
func SetDefaultGitSchemes(schemes []string) {
	if schemes == nil {
		defaultGitSchemes = []string{"git", "https", "http", "git+ssh"}
	} else {
		defaultGitSchemes = schemes
	}
}

// findGitDirs expects to be pointed at a git workspace, either
// bare or standard.  It'll find the gitdir and worktree dirs
// and return them, if it fails it'll return non-nil err.  Params:
//	path (string): path to the git workspace
// Returns:
//	gitDir (string): path to git metadata location
//	workTreeDir (string): working tree, "" if bare clone
//	err (error): a valid error if unable to find a git repo
func findGitDirs(path string) (string, string, error) {
	gitDir := filepath.Join(path, ".git") // see if std git clone
	var err error
	var exists bool
	if exists, err = dir.Exists(gitDir); exists && err == nil {
		return gitDir, path, nil
	}
	gitRefsDir := filepath.Join(path, "refs")
	if exists, err = dir.Exists(gitRefsDir); exists && err == nil {
		gitConfigFile := filepath.Join(path, "config")
		if exists, err = file.Exists(gitConfigFile); exists && err == nil {
			return path, "", nil
		}
	}
	return "", "", out.WrapErrf(ErrNoExist, 4500, "Unable to find valid git dir under path: %s, err: %s", path, err)
}
