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

// Updater provides a small interface to work with different VCS systems with
// respect to bringing in and "merging" remote VCS changes into an existing
// local workspace clone (and any local changes).  A Describer interface
// comes in as a parameter to get local wkspc path and remote location info
// and Exists from the Existence intfc is also needed (sometimes)..
// Note: do not use the interface Existence as it will conflict and result
//       in build errors since the Existence itfc also is a Describer
type Updater interface {
	// Describer access to VCS system details (Remote, WkspcPath, ..)
    Describer

	// Exists will determine if the repo exists (remotely or in local wkspc)
	Exists(Location) (bool, error)

	// Update is used to merge with new central repo changes to local
    // workspace, optionally at a given revision (specific single revision)
	Update(...Rev) (string, error)
}

// NewUpdater returns a VCS Updater based on the given VCS description info about
// the remote (URL typically) and workspace (dir/path) locations.  If the VCS info
// is minimal (eg: not a full URL with scheme) then this will try and detect
// VCS type (if unable to determine an ErrCannotDetectVCS will be returned).
// Note: This function can make network calls to try to determine the VCS
func NewUpdater(remote, wkspc string, vcsType ...Type) (Updater, error) {
	vtype, remote, err := detectVCSType(remote, wkspc, vcsType...)
	if err != nil {
		return nil, err
	}
	switch vtype {
	case Git:
		return NewGitUpdater(remote, wkspc)
	case Svn:
		return NewSvnUpdater(remote, wkspc)
	case Hg:
		return NewHgUpdater(remote, wkspc)
	case Bzr:
		return NewBzrUpdater(remote, wkspc)
	}

	// Should never fall through to here but just in case.
	//FIXME: erik: I think we need an ErrVCSNotImplemented or
	//       something like that to indicate the VCS does
    //       not support the given operation (leading towards
    //       support for VCS's that only support some ops)
	return nil, ErrCannotDetectVCS
}

