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

// Describer provides a small interface to get basic vcs definition data such
// as where the VCS lives remotely and where it belongs locally, as well as
// a method for determining the type of repo it is (if it is known yet)
// Note: if you change any of these method sig's please check for their
//       use across all files in this package, eg: get.go has copied the
//       Remote() & WkspcPath() func signatures (see comment there)
type Describer interface {
	// Vcs retrieves the underlying VCS being implemented.
	Vcs() Type

	// Remote retrieves the remote location/URL for a repo.
	Remote() string

	// WkspcPath retrieves the wkspc local file system path for a repo.
	WkspcPath() string

	// RemoteRepoName retrieves any VCS specific naming for the remote repo,
	// eg: "origin" is the default for the repo we cloned from in git, in
	//     other VCS's this can come back unused with a ""
	RemoteRepoName() string
}

// Description is a structure that satisfies the VCS Describer implementation, used
// by the <VCS>Reader and other <VCS> implementations (eg: GitPulller)
type Description struct {
	remote, wkspc, remoteRepoName string
	vcsType Type
}

// Remote retrieves the remote location for a repo.
func (b *Description) Remote() string {
	return b.remote
}

// RemoteRepoName is only needed for some VCS's, but if the remote repo
// has a "name" to identify it (eg: "origin" in git) then one can get
// the current remote repo name here (to set: see SetRemoteRepoName())
func (b *Description) RemoteRepoName() string {
	return b.remoteRepoName
}

// WkspcPath retrieves the wkspc file system location for a repo.
func (b *Description) WkspcPath() string {
	return b.wkspc
}

// Vcs retrieves the VCS type if we have one
func (b *Description) Vcs() Type {
	return b.vcsType
}

func (b *Description) setRemote(remote string) {
	b.remote = remote
}

func (b *Description) setRemoteRepoName(remRepoName string) {
	b.remoteRepoName = remRepoName
}

func (b *Description) setWkspcPath(wkspc string) {
	b.wkspc = wkspc
}

func (b *Description) setVcs(vcsType Type) {
	b.vcsType = vcsType
}

