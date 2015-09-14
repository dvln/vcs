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

// Getter provides a small interface to work with different VCS systems with
// respect to "getting" a VCS and bringing it into a workspace, it requires
// one to pass in a vcs.Describer so it can find the remote VCS and identify
// where to put it locally
type Getter interface {
	// Describer access to VCS system details (Remote, WkspcPath, ..)
    Describer

	// RevSet matches the RevSet() interface sig from RevSetter. Due to Go's
    // design here can't just make this a RevSetter since the overall Reader
	// in repo.go includes RevReader and the Getter interface (and how Go works
    // is a compile error indicating duplicate method names)
    //   see: https://groups.google.com/forum/#!topic/golang-nuts/OKgbtTW-5YQ
	RevSet(Rev) (string, error)

	// Exists will determine if the repo exists (remotely or in local wkspc)
	Exists(Location) (bool, error)

	// Get is used to perform an initial clone/checkout of a repository.
	Get(...Rev) (string, error)
}

// NewGetter returns a VCS Getter based on the given VCS description info about
// the remote (URL typically) and workspace (dir/path) locations.  If the VCS info
// is minimal (eg: not a full URL with scheme) then this will try and detect
// VCS type (if unable to determine an ErrCannotDetectVCS will be returned).
// Note: This function can make network calls to try to determine the VCS
func NewGetter(remote, wkspc string, vcsType ...Type) (Getter, error) {
	vtype, remote, err := detectVCSType(remote, wkspc, vcsType...)
	if err != nil {
		return nil, err
	}
	switch vtype {
	case Git:
		return NewGitGetter(remote, wkspc)
	case Svn:
		return NewSvnGetter(remote, wkspc)
	case Hg:
		return NewHgGetter(remote, wkspc)
	case Bzr:
		return NewBzrGetter(remote, wkspc)
	}

	// Should never fall through to here but just in case.
	//FIXME: erik: I think we need an ErrVCSNotImplemented or
	//       something like that to indicate the VCS does
    //       not support the given operation (leading towards
    //       support for VCS's that only support some ops)
	return nil, ErrCannotDetectVCS
}

