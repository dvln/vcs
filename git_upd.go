// Copyright © 2015 Erik Brady <brady@dvln.org>
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
type GitUpdater struct {
	Description
	mirror bool
	rebase RebaseVal
}

// NewGitUpdater creates a new instance of GitUpdater. The remote and wkspc URL/dir
// need to be passed in.  Params:
//	remote (string): URL of remote repo
//	wkspc (string): Directory for the local workspace
//	mirror (bool): if a full mirroring of all content is desired
//	rebase (RebaseVal): if rebase wanted or not, what type
func NewGitUpdater(remote, wkspc string, mirror bool, rebase RebaseVal) (Updater, error) {
	ltype, err := DetectVcsFromFS(wkspc)
	// Found a VCS other than Git. Need to report an error.
	if err == nil && ltype != Git {
		return nil, ErrWrongVCS
	}
	u := &GitUpdater{}
	u.mirror = mirror
	u.rebase = rebase
	// Set up initial remote URL/path, repo name, wkspc path, URL schemes, VCS type
	// FIXME: weak, origin is hard coded here and in GitCheckRemote()
	u.setDescription(remote, "origin", wkspc, defaultGitSchemes, Git)
	if err == nil { // Have a local wkspc FS repo, try to validate/upd remote
		remote, _, err = GitCheckRemote(u, remote)
		if err != nil {
			return nil, err
		}
		u.setRemote(remote)
	}
	return u, nil // note: above 'err' not used on purpose here..
}

// Update allows generic git updater to update VCS's, like git fetch+merge
func (u *GitUpdater) Update(rev ...Rev) (string, error) {
	return GitUpdate(u, rev...)
}

// Exists support for git updater
func (u *GitUpdater) Exists(l Location) (string, error) {
	return GitExists(u, l)
}
