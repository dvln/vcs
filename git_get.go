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

import (
	"os"
	"os/exec"
	"strings"
)

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

	r := &GitGetter{}
	r.setRemote(remote)
	r.setWkspcPath(wkspc)
	r.setVcs(Git)
	r.setRemoteRepoName("origin")
	//FIXME: erik: this is weak, need to support flexibility depending upon
	//       detail a client may have wanted (eg: update from vendor repo or
	//       joe's repo or whatever, etc)... consider improvements, see
	//       "git config --get remote.origin.url" hard coded below as well
    //         gitdir% git config --get remote.origin.url
    //         ssh://sjc-acmegit-v01:29718/acme/acme
    //         gitdir% remote -v
    //         origin	ssh://sjc-acmegit-v01:29718/acme/acme (fetch)
    //         origin	ssh://sjc-acmegit-v01:29718/acme/acme (push)
    //         %

	// Make sure the wkspc Git repo is configured the same as the remote when
	// A remote value was passed in.
	if exists, chkErr := r.Exists(Wkspc); err == nil && chkErr == nil && exists {
		oldDir, err := os.Getwd()
		if err != nil {
			return nil, err
		}
		err = os.Chdir(wkspc)
		if err != nil {
			return nil, err
		}
		defer os.Chdir(oldDir)
		output, err := exec.Command("git", "config", "--get", "remote.origin.url").CombinedOutput()
		if err != nil {
			return nil, err
		}

		localRemote := strings.TrimSpace(string(output))
		if remote != "" && localRemote != remote {
			return nil, ErrWrongRemote
		}

		// If no remote was passed in but one is configured for the locally
		// checked out Git repo use that one.
		if remote == "" && localRemote != "" {
			r.setRemote(localRemote)
		}
	}

	return r, nil
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
func (g *GitGetter) Exists(l Location) (bool, error) {
	return GitExists(g, l)
}

