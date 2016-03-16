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

// Existence is an interface used to determine if a repo (vcs pkg)
// exists in the local workspace or at a remote URL (depending
// upon the location being checked).
type Existence interface {
	// Describer access to VCS system details (Remote, WkspcPath, ..)
	Describer

	// Exists will determine if the repo exists (remotely or in local wkspc)
	Exists(Location) (string, error)
}
