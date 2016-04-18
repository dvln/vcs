// Copyright © 2016 Erik Brady <brady@dvln.org>
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
	"fmt"

	"github.com/dvln/out"
)

// Resulter provides a small interface to get basic vcs definition data such
// as where the VCS lives remotely and where it belongs locally, as well as
// a method for determining the type of repo it is (if it is known yet)
// Note: if you change any of these method sig's please check for their
//       use across all files in this package, eg: get.go has copied the
//       Remote() & LocalRepoPath() func signatures (see comment there)
type Resulter interface {
	// All returns the list of commands run and any output from each cmd
	All() []*Result

	// Last gives the last raw/core SCM cmd run and it's output
	// via the Result structure returned
	Last() *Result

	// Add adds a new result to the list of results
	add(*Result)
}

// Result is a structure that satisfies the VCS Resulter implementation, used
// by the <VCS>Reader and other <VCS> implementations (eg: GitUpdater)
type Result struct {
	cmd    string
	output string
}

// Results is a structure that contains all the commands run and their
// output for the most recent run of any method that runs VCS cmds
type Results struct {
	results []*Result
}

// newResult generates a new single result
func newResult() *Result {
	return &Result{}
}

// newResults generates a new results structure
func newResults() *Results {
	return &Results{}
}

// All returns all cmds/output (results) from the most recent SCM op,
// keep in mind that some SCM ops (eg: update) might need multiple raw SCM
// cmds to accomplish what the overall function needs to accomplish.
func (r *Results) All() []*Result {
	return r.results
}

// Last returns the absolutely last result (cmd/output) that
// was done (raw SCM cmd/output)... useful for finding the cmd and
// output of any failed command
func (r *Results) Last() *Result {
	length := len(r.results)
	if length == 0 {
		return nil
	}
	return r.results[length-1]
}

// Add is a method on Results that allows one to add a new result
func (r *Results) add(result *Result) {
	r.results = append(r.results, result)
}

// String implements a stringer for the *Result type so we can print out string
// representations for any result
func (r *Result) String() string {
	indentedOut := out.InsertPrefix(r.output, "  ", out.AlwaysInsert, 0)
	return fmt.Sprintf("cmd: %s, output:\n%s", r.cmd, indentedOut)
}

// String implements a stringer for the *Results type so we can print out string
// representations for all results
func (r *Results) String() string {
	results := r.All()
	resultsStr := ""
	for i, result := range results {
		indentedOut := out.InsertPrefix(result.output, "  ", out.AlwaysInsert, 0)
		if indentedOut == "" {
			indentedOut = "[No output from command]\n"
		}
		resultsStr += fmt.Sprintf("cmd %d: %s, output %d:\n%s", i+1, result.cmd, i+1, indentedOut)
	}
	return resultsStr
}
