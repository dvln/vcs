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

// GitReader implements the VCS Reader interface for the Git source control,
// start out by adding a base VCS description structure (implements Describer)
type GitReader struct {
	Description
}

// NewGitReader creates a new instance of GitReader. The remote and wkspc URL/dir
// need to be passed in.
func NewGitReader(remote, wkspc string) (*GitReader, error) {
	ltype, err := DetectVcsFromFS(wkspc)
	// Found a VCS other than Git. Need to report an error.
	if err == nil && ltype != Git {
		return nil, ErrWrongVCS
	}
	r := &GitReader{}
    r.setDescription(remote, "origin", wkspc, defaultGitSchemes, Git)
	if err == nil {
		newRemote, _, err := GitCheckRemote(r, remote)
		if err != nil {
			return nil, err
		}
		r.setRemote(newRemote)
	}
	return r, nil	// note: above 'err' not used on purpose here..
}

// RevSet support for git reader
func (r *GitReader) RevSet(rev Rev) (string, error) {
	return GitRevSet(r, rev)
}

// RevRead support for git reader
func (r *GitReader) RevRead(scope ReadScope, vcsRev ...Rev) ([]Revisioner, string, error) {
	return GitRevRead(r, scope, vcsRev...)
}

// Exists support for git reader
func (r *GitReader) Exists(l Location) (string, error) {
	return GitExists(r, l)
}

