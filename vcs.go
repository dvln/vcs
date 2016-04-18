// This package is a derivative of the github.com/Masterminds/vcs repo, thanks
// for the base ideas folks!  Copyright at time of fork was MIT.
//
// Further dvln related restructuring/changes licensed via:
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

// Package vcs provides the ability to work with varying version control systems
// (VCS),  also known as source control systems (SCM) though the same interface.
//
// This package includes a function that attempts to detect the repo type from
// the remote URL and return the proper type. For example,
//
//     remote := "https://github.com/Masterminds/vcs"
//     localPath, _ := ioutil.TempDir("", "go-vcs")
//     vcsReader, err := vcs.NewReader(remote, localPath)   // add VCS Type if known
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
	"fmt"
	"os"
	"os/exec"
	"strings"
)

var (
	// ErrNoExist is returned when a repo/pkg can't be found (local|remote)
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
	// LocalPath indicates we have a local clone in a work area
	LocalPath Location = "local"
	// Remote indicates we have a remote clone not on the local host
	Remote Location = "remote"
)

// RebaseVal indicates what rebase mode is active
type RebaseVal int

// Rebase settings that are valid
const (
	// RebaseFalse indicates we are in merge mode, not rebase mode
	RebaseFalse RebaseVal = 0
	// RebaseTrue means we are standard rebase merge mode
	RebaseTrue RebaseVal = 1
	// RebasePreserve means rebase but local merge commits aren't flattened
	RebasePreserve RebaseVal = 2
	// RebaseUser means don't give a rebase option, use the users setting
	RebaseUser RebaseVal = 3
)

// RefOp describes operations that can be done with references
type RefOp string

// RefOp possibilities
const (
	// RefFetch indicates to fetch the specific reference from our source
	RefFetch RefOp = "fetch"
	// RefDelete indicates to delete the ref from the local clone, poof
	RefDelete RefOp = "delete"
)

// run will execute the given cmd and args and return the results and
// any error that occurred.  Params:
//	cmd (string): top level cmd (eg: "git" or "/path/to/git")
//	args (...string): what will be space separated args, empty args ignored
// Note: the args should not have strings like "-o blah", instead: "-o", "blah"
// Returns:
//	*Result: a single result structure (command run, raw output from cmd)
//	error: a Go error if anything goes astray in the exec.Command()
func run(cmd string, args ...string) (*Result, error) {
	var finalArgs []string
	for _, arg := range args {
		if arg != "" {
			finalArgs = append(finalArgs, arg)
		}
	}
	output, err := exec.Command(cmd, finalArgs...).CombinedOutput()
	result := newResult()
	result.cmd = fmt.Sprintf("%s %s", cmd, strings.Join(finalArgs, " "))
	result.output = string(output)
	return result, err
}

// runFromLocalRepoDir attempts to cd into the pkg's workspace root dir (VCS root)
// and run the command from that location (then it cd's back).  The command
// result is returned along with any error (see run() for details).
// WARNING: any SCM using this must avoid goroutines, Chdir() is not safe!!!
// WARNING: instead use the run() routine with SCM opts to specify repo location
func runFromLocalRepoDir(localRepoDir, cmd string, args ...string) (*Result, error) {
	oldDir, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	err = os.Chdir(localRepoDir)
	if err != nil {
		return nil, err
	}
	defer os.Chdir(oldDir)
	var finalArgs []string
	for _, arg := range args {
		if arg != "" {
			finalArgs = append(finalArgs, arg)
		}
	}
	return run(cmd, finalArgs...)
}

// detectVCSType tries to determine what VCS we are working with and can
// return an more complete remote URL/path.  Note that sub-methods can
// access the network in some situations to help determine this (although
// it tries not to, making some assumptions and run as quickly as it can).
// It will return the VCS type it determines along with the same or an
// update 'remote' URL/string and any errors that occurred.
func detectVCSType(remote, localPath string, vcsType ...Type) (Type, string, error) {
	var err error
	vtype := NoVCS
	if vcsType != nil && len(vcsType) == 1 && vcsType[0] != NoVCS {
		vtype = vcsType[0]
	} else {
		vtype, remote, err = detectVcsFromRemote(remote)

		// If from the remote URL the VCS could not be detected, see if the localPath
		// repo contains enough information to figure out the VCS. The reason the
		// localPath repo is not checked first is because of the potential for VCS type
		// switches which will be detected in each of the type builders.
		if err == ErrCannotDetectVCS {
			vtype, err = DetectVcsFromFS(localPath)
		}
	}
	return vtype, remote, err
}
