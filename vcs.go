// This package is a derivative of the github.com/Masterminds/vcs repo, thanks
// for the base ideas folks!  Copyright at time of fork was MIT.
//
// Further dvln related restructuring/changes licensed via:
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

// Package vcs provides the ability to work with varying version control systems
// (VCS),  also known as source control systems (SCM) though the same interface.
//
// This package includes a function that attempts to detect the repo type from
// the remote URL and return the proper type. For example,
//
//     remote := "https://github.com/Masterminds/vcs"
//     wkspc, _ := ioutil.TempDir("", "go-vcs")
//     vcsReader, err := vcs.NewReader(remote, wkspc)   // add VCS Type if known
//
// In this case vcs will use a GitReader instance. NewReader can detect the VCS
// for numerous popular VCS and from the URL. For example, a URL ending in .git
// that's not from one of the popular VCS will be detected as a Git repo and
// the correct reader type will be returned.
//
// If you know the VCS system type and would like to create an instance of a
// specific type you can add it as an optional 3rd param to NewReader or you
// can use one of the constructurs for specific VCS types, via calls to
// NewGitReader, NewSvnReader, NewBzrReader, and NewHgReader.
//
// Once you have an object implementing a VCS Reader interface the operations
// are the same no matter which VCS you're using. There are some caveats. For
// example, each VCS has its own revision formats that need to be respected and
// to checkout a branch, if a branch is being worked with, is different in
// each VCS.  These revisions are passed as type 'Rev' which is just essentially
// a string but when revisions are read a Revision struct can be populated with
// more extensive data about a given revision if desired (raw VCS revision, tags,
// branches, timestamp, etc)
package vcs

import (
	"errors"
	"os"
	"os/exec"
)

var (
	// ErrNoExist is returned when a repo/pkg can't be found (wkspc|remote)
	ErrNoExist = errors.New("VCS does not exist (ie: could not be found)")

	// ErrWrongVCS is returned when an action is tried on the wrong VCS.
	ErrWrongVCS = errors.New("Wrong VCS detected")

	// ErrCannotDetectVCS is returned when VCS cannot be detected from URI string.
	ErrCannotDetectVCS = errors.New("Cannot detect VCS")

	// ErrWrongRemote occurs when the passed in remote does not match the VCS
	// configured endpoint.
	ErrWrongRemote = errors.New("The Remote does not match the VCS endpoint")
)

// Type describes the type of VCS
type Type string

// VCS Types
const (
	NoVCS Type = ""
	Bzr   Type = "bzr"
	Git   Type = "git"
	Hg    Type = "hg"
	Svn   Type = "svn"
)

// Location describes the location to check for a repo (vcs pkg)
type Location string

// Location settings
const (
	Wkspc  Location = "wkspc"
	Remote Location = "remote"
)

// run will execute the given cmd and args and return the output and
// any error that occurred.
func run(cmd string, args ...string) (string, error) {
	output, err := exec.Command(cmd, args...).CombinedOutput()
	return string(output), err
}

// runFromWkspcDir attempts to cd into the pkg's workspace root dir (VCS root)
// and run the command from that location (then it cd's back).  The command
// output is returned along with any error.
func runFromWkspcDir(wkspcDir, cmd string, args ...string) (string, error) {
	oldDir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	err = os.Chdir(wkspcDir)
	if err != nil {
		return "", err
	}
	defer os.Chdir(oldDir)

	return run(cmd, args...)
}

// detectVCSType tries to determine what VCS we are working with and can
// return an more complete remote URL/path.  Note that sub-methods can
// access the network in some situations to help determine this (although
// it tries not to, making some assumptions and run as quickly as it can).
// It will return the VCS type it determines along with the same or an
// update 'remote' URL/string and any errors that occurred.
func detectVCSType(remote, wkspc string, vcsType ...Type) (Type, string, error) {
	var err error
	vtype := NoVCS
	if vcsType != nil && len(vcsType) == 1 && vcsType[0] != NoVCS {
		vtype = vcsType[0]
	} else {
		vtype, remote, err = detectVcsFromRemote(remote)

		// If from the remote URL the VCS could not be detected, see if the wkspc
		// repo contains enough information to figure out the VCS. The reason the
		// wkspc repo is not checked first is because of the potential for VCS type
		// switches which will be detected in each of the type builders.
		if err == ErrCannotDetectVCS {
			vtype, err = DetectVcsFromFS(wkspc)
		}
	}
	return vtype, remote, err
}
