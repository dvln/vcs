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

// The rev.go part of the 'vcs' package is focused around reading and
// possibly updating an existing VCS revisions meta-data.  The metadata
// read may be partial or complete depending upon the desire (ie: basic
// core/raw VCS version) info or full data (committer, author, dates,
// comment, tags, known branch(es), known semantic versions or other
// key versions plus the raw/core VCS version identifier, eg: sha1).
// Reading more data can take more time.
// - caching will need to be added over time (and flushing, etc)
// - eventually greater flexibility than minimal or all should be
// added, perhaps via a bit configuration field or something

package vcs

import (
    "time"
)

/*
func init() {
// FIXME: erik: could set default regex for semversion "matching", although
// it may make more sense to put this within each VCS implemented and let
// it be overridden/controlled at the codebase and pkg level for 'dvln' as
// well as via config file (cfg would be global or per scm, codebase would
// be codebase global or per pkg, perhaps inherited for group/codebase pkgs)
}
*/


// ReadScope describes how revision read ops should be focused (*if* a choice for a given VCS)
type ReadScope string

// ReadScope settings, do we want to optimize for speed (core/raw VCS rev is
// all that is guaranteed there) or for as much data as we can populate about
// a revision?
const (
	CoreRev ReadScope = "CoreRev"
    AllData ReadScope = "AllData"
)

// UserType indicates which revision settings we are mucking with
// (time/date, name, userid), eg: Committer, Author
type UserType string

const (
	Author    UserType = "author"    // set author data
	Committer UserType = "committer" // set committer data
	AuthComm  UserType = "authcomm" // for setting both Auth & Committer
)

// Rev corresponds to a core/raw revision for a VCS, it might be a sha1,
// it might be a tag, could be a timestamp potentially, 
type Rev string

// FIXME: erik: these interfaces are not really used, just pondering

// RevAccesser is focused on getting a VCS's revision data stored in
// a Revision structure (typically, but really anything implementing
// the interface works of course).
type RevAccesser interface {
    Core() Rev
    SemVers() []Rev
    Tags() []Rev
    Branches() []Rev
    TStamp(UserType) *time.Time
    Comment() string
    UserInfo(UserType) (string, string)
}

// RevStorer is focused on pushing VCS data into an in-memory storage
// location (typically a Revision struct but anything that implements
// the interface could be used)
type RevStorer interface {
    SetCore(Rev)
    SetSemVers([]Rev)
    SetTags([]Rev)
    SetBranches([]Rev)
	SetTStamp(UserType, *time.Time)
    SetComment(string)
    SetUserInfo(UserType, string, string)
}

// Revisioner is a generic interface to a specific VCS revisions data.
// One can fill in the data via the RevStorer interface and one can
// access the data via the RevAccess interface.
type Revisioner interface {
	RevAccesser
	RevStorer
}


// Rev stores basic VCS revision information, as much as can be gleaned from
// a given VCS system with, at the very least, having support for the Core
// VCS version identifier.
type Revision struct {
	core Rev                // The core/raw SCM revision (string, eg: sha1 for git)
    semVers []Rev           // Semantic Versions (if any) on this VCS revision
    refVers []Rev           // Commit reference versions/names (VCS specific)
    tags []Rev              // Any tags found on this VCS revision (non Sem/Ref)
    branches []Rev          // Any branch "latest" pointers on this VCS revision
    ancestors []Rev         // List of ancesors of the commit
	comment string			// Full comment for the revision
    author string           // Full name of the revision author
	authorId string         // User id of the revision author
	authorTStamp *time.Time // Authors revision creation time
	committer string        // Full name of the revision committer
	committerId string      // User id of the revision committer
	committerTStamp *time.Time // Committers revision creation time
}

// RevReader examines the workspace to determine what *current* VCS rev
// is in the workspace local path
type RevReader interface {
	// Describer access to VCS system details (Remote, WkspcPath, ..)
    Describer

	// RevRead retrieves the current in-workspace VCS core/raw revision
	RevRead(ReadScope, ...Rev) ([]Revisioner, string, error)
}

// RevSetter changes the current workspace revision of a pkg/repo
// Note: if you change any of these method sig's please check for their
//       use across all files in this package, eg: get.go has copied the
//       RevSet() func signatures (see comment in that file)
type RevSetter interface {
	// Describer access to VCS system details (Remote, WkspcPath, ..)
    Describer

	// RevSet sets the revision of a package/repo (eg: git checkout)
	RevSet(Rev) (string, error)
}

// RevWriter will make sure data on a specified revision matches the given data
type RevWriter interface {
	// Describer access to VCS system details (Remote, WkspcPath, ..)
    Describer

	// RevCommit examines the revision and verifies that all revision
    // setting are applied (all tags, all branch latest, etc)... if
    // anything is changed 'true' will be returned, only changes the
    // workspace state (local clone for DVCS, in-memory structure
    // only for CVCS)... see RevPush() for updating remote (central) VCS
	RevCommit(*Revision) (bool, error)

    // RevPush will push local revision changes to central clone/repo
	RevPush() error
}

// RevReadSetWriter combines the ability to read, set and write (commit/push)
// a VCS revision details (meta-data).
type RevReadSetWriter interface {
	// Describer access to VCS system details (Remote, WkspcPath, ..)
    Describer

	// RevRead retrieves the current in-workspace VCS core/raw revision
	RevRead(ReadScope, ...Rev) ([]Revisioner, string, error)

	// RevSet sets the revision of a package/repo (eg: git checkout)
	RevSet(Rev) (string, error)

	// FIXME: erik: get RevCommit and RevPush added in here
}

// NewRevision will contruct a new empty revision structure
func NewRevision() *Revision {
	return &Revision{}
}

// Core returns the core/raw VCS version
func (r *Revision) Core() Rev {
	return r.core
}

// SemVers returns a list of semvers (if any stored in the revision)
func (r *Revision) SemVers() []Rev {
	return r.semVers
}

// Tags returns a list of tags (if any stored in the revision)
func (r *Revision) Tags() []Rev {
	return r.tags
}

// Branches returns a list of tags (if any stored in the revision)
func (r *Revision) Branches() []Rev {
	return r.branches
}

// TStamp returns the timestamp on the revision, one needs
// to identify if author or committer timestamp is desired
// (in many VCS's there is no difference).
func (r *Revision) TStamp(utype UserType) *time.Time {
	if utype == Author {
		return r.authorTStamp
	}
	return r.committerTStamp
}

// Comment returns the comment on the revision
func (r *Revision) Comment() string {
	return r.comment
}

// UserInfo returns the author or committer name & userid info (in that
// order).  If author is requested it is returned, otherwise the committer
// information will be returned (not that for many VCS's it is one and
// the same)
func (r *Revision) UserInfo(utype UserType) (string, string) {
	if utype == Author {
		return r.author, r.authorId
	}
	return r.committer, r.committerId
}

// SetCore sets the core VCS revision in the basic revision struct
// (note that when reading a revision these will be set as well as
// possible depending upon a fast read or a data-heavy read), these
// will only be written to a VCS if you write the revision.
func (r *Revision) SetCore(rev Rev) {
	r.core = rev
}

// SetSemVers can set one or more semvers compat tags on this VCS revision
// structure (note that when reading a revision these will be set as well
// as possible depending upon a fast read or a data-heavy read), these
// will only be written to a VCS if you write the revision.
func (r *Revision) SetSemVers(semVers []Rev) {
	r.semVers = semVers
}

// SetTags sets any valid tags in this VCS revision structure
// (note that when reading a revision these will be set as well as
// possible depending upon a fast read or a data-heavy read), these
// will only be written to a VCS if you write the revision.
func (r *Revision) SetTags(tags []Rev) {
	r.tags = tags
}

// SetBranches sets any valid branch (latest rev) that correspond to this vcs rev
// (note that when reading a revision these will be set as well as
// possible depending upon a fast read or a data-heavy read), these
// will only be written to a VCS if you write the revision.
func (r *Revision) SetBranches(branches []Rev) {
	r.branches = branches
}

// SetTStamp sets the timestamp of the version (if available)
// Note that when reading a revision the time will be set if
// possible (if doing a fast read perhaps only the core/raw rev),
// and the time is never written to the VCS (we leave that to
// the VCS) so it is ignored upon writes of new rev data.
func (r *Revision) SetTStamp(utype UserType, timestamp *time.Time) {
	if utype == Author || utype == AuthComm {
		r.authorTStamp = timestamp
	}
	if utype == Committer || utype == AuthComm {
		r.committerTStamp = timestamp
	}
}

// SetComment stashes the comment for a given revision in
// the *Revision structure.
func (r *Revision) SetComment(comment string) {
	r.comment = comment
}

// SetUserInfo stashes the user name and id (email) in
// the *Revision structure.
func (r *Revision) SetUserInfo(utype UserType, name string, userid string) {
	if utype == Author || utype == AuthComm {
		r.author = name
		r.authorId = userid
	}
	if utype == Committer || utype == AuthComm {
		r.committer = name
		r.committerId = userid
	}
}

