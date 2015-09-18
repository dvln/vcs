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

// GitGetter implements the VCS Getter interface for the Git source control,
// start out by adding a base VCS description structure (implements Describer)
type GitGetter struct {
	Description
}

// NewGitGetter creates a new instance of GitGetter. The remote and wkspc URL/dir
// need to be passed in.
func NewGitGetter(remote, wkspc string) (Getter, error) {
	ltype, err := DetectVcsFromFS(wkspc)
	// Found a VCS other than Git. Need to report an error.
	if err == nil && ltype != Git {
		return nil, ErrWrongVCS
	}
	g := &GitGetter{}
    g.setDescription(remote, "origin", wkspc, defaultGitSchemes, Git)
	if err == nil { // Have a local wkspc FS repo, try to validate/upd remote
		remote, _, err = GitCheckRemote(g, remote)
		if err != nil {
			return nil, err
		}
		g.setRemote(remote)
	}
	return g, nil	// note: above 'err' not used on purpose here..
}

// Get support for git getter
func (g *GitGetter) Get(rev ...Rev) (string, error) {
	return GitGet(g, rev...)
}

// RevSet support for git getter
func (g *GitGetter) RevSet(rev Rev) (string, error) {
	return GitRevSet(g, rev)
}

// Exists support for git getter
func (g *GitGetter) Exists(l Location) (string, error) {
	return GitExists(g, l)
}

