// Copyright © 2015,2016 Erik Brady <brady@dvln.org>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package vcs

// GitUpdater implements the VCS Updater interface for the Git source control,
// start out by adding a base VCS description structure (implements Describer)
// and some update focused controls set via New[SCM]Updater, also stores all
// git cmds run and their output (the cmds/out for most recent Update() run)
type GitUpdater struct {
	Description
	mirror bool
	rebase RebaseVal
	refs   map[string]RefOp
}

// NewGitUpdater creates a new instance of GitUpdater. The remote and localPath URL/dir
// need to be passed in amongst other parameters.  Params:
//	remote (string): URL of remote repo (optional if remoteNAme is given)
//	remoteName (string): If there is a name for the remote repo (eg: "origin" default)
//	localPath (string): Directory for the local repo/clone/workspace to update
//	mirror (bool): if a full mirroring of all content is desired
//	rebase (RebaseVal): if rebase wanted or not, what type
//	refs (map[string]RefOp): list of refs to act on w/given operation (or nil)
// Note that this will populate/validate remote using remoteName (default: origin)
// Returns an updater interface and any error that may have occurred
func NewGitUpdater(remote, remoteName, localPath string, mirror bool, rebase RebaseVal, refs map[string]RefOp) (Updater, error) {
	ltype, err := DetectVcsFromFS(localPath)
	// Found a VCS other than Git. Need to report an error.
	if err == nil && ltype != Git {
		return nil, ErrWrongVCS
	}
	u := &GitUpdater{}
	u.mirror = mirror
	u.rebase = rebase
	if refs != nil { // if refs given, then set up refs to act on w/ops
		u.refs = make(map[string]RefOp)
		u.refs = refs
	}
	// Set up initial remote URL/path, repo name, localPath path, URL schemes, VCS type
	if remoteName == "" {
		remoteName = "origin"
	}
	u.setDescription(remote, remoteName, localPath, defaultGitSchemes, Git)
	if err == nil { // Have a localPath FS repo, try to validate/upd remote
		remote, _, err = GitCheckRemote(u, remote)
		if err != nil {
			return nil, err
		}
		u.setRemote(remote)
	}
	return u, nil // note: above 'err' not used on purpose here..
}

// Update allows generic git updater to update VCS's, like git fetch+merge
func (u *GitUpdater) Update(rev ...Rev) (Resulter, error) {
	return GitUpdate(u, rev...)
}

// Exists support for git updater
func (u *GitUpdater) Exists(l Location) (string, Resulter, error) {
	return GitExists(u, l)
}
