package vcs

import (
	"io/ioutil"
	"os"
	"testing"
)

// Canary test to ensure GitReader implements the Reader interface.
var _ Reader = &GitReader{}

// To verify git is working we perform intergration testing
// with a known git service.

func TestGit(t *testing.T) {

	tempDir, err := ioutil.TempDir("", "go-vcs-git-tests")
	if err != nil {
		t.Error(err)
	}
	defer func() {
		err = os.RemoveAll(tempDir)
		if err != nil {
			t.Error(err)
		}
	}()

	gitReader, err := NewGitReader("https://github.com/Masterminds/VCSTestRepo", tempDir+"/VCSTestRepo")
	if err != nil {
		t.Errorf("Unable to instantiate new Git VCS reader, Err: %s", err)
	}

	if gitReader.Vcs() != Git {
		t.Error("Git is detecting the wrong type")
	}

	// Check the basic getters.
	if gitReader.Remote() != "https://github.com/Masterminds/VCSTestRepo" {
		t.Error("Remote not set properly")
	}
	if gitReader.WkspcPath() != tempDir+"/VCSTestRepo" {
		t.Error("Local disk location not set properly")
	}

	// Do an initial clone.
	_, err = gitReader.Get()
	if err != nil {
		t.Errorf("Unable to clone Git repo using VCS reader Get(). Err was %s", err)
	}

	// Verify Git repo exists in the workspace
	exists, err := gitReader.Exists(Wkspc)
	if err != nil {
		t.Errorf("Existence check failed on git repo: %s", err)
	}
	if exists == false {
		t.Error("Problem seeing if Git repo exists in workspace")
	}

	// Test internal lookup mechanism used outside of Git specific functionality.
	ltype, err := DetectVcsFromFS(tempDir + "/VCSTestRepo")
	if err != nil {
		t.Error("detectVcsFromFS unable to Git repo")
	}
	if ltype != Git {
		t.Errorf("detectVcsFromFS detected %s instead of Git type", ltype)
	}

	// Test NewReader on existing checkout. This should simply provide a working
	// instance without error based on looking at the local directory.
	ngitReader, nrerr := NewReader("https://github.com/Masterminds/VCSTestRepo", tempDir+"/VCSTestRepo")
	if nrerr != nil {
		t.Error(nrerr)
	}
	// See if the new git VCS reader was instantiated in the workspace
	exists, err = ngitReader.Exists(Wkspc)
	if err != nil {
		t.Errorf("Existence check failed on git repo: %s", err)
	}
	if exists == false {
		t.Error("The git reader was not correctly instantiated in the workspace")
	}

	// Perform an update operation
	_, err = gitReader.Update()
	if err != nil {
		t.Error(err)
	}

	// Set the version using the short hash.
	_, err = gitReader.RevSet("806b07b")
	if err != nil {
		t.Errorf("Unable to update Git repo version. Err was %s", err)
	}

	// Use RevRead to verify we are on the right version.
	v, _, err := gitReader.RevRead(CoreRev)
	if string(v.Core()) != "806b07b08faa21cfbdae93027904f80174679402" {
		t.Errorf("Error checking checked out Git version, found: \"%s\"\n", string(v.Core()))
	}
	if err != nil {
		t.Error(err)
	}

	// Verify that we can set the version something other than short hash
	_, err = gitReader.RevSet("master")
	if err != nil {
		t.Errorf("Unable to update Git repo version. Err was %s", err)
	}
	_, err = gitReader.RevSet("806b07b08faa21cfbdae93027904f80174679402")
	if err != nil {
		t.Errorf("Unable to update Git repo version. Err was %s", err)
	}
	v, _, err = gitReader.RevRead(CoreRev)
	if string(v.Core()) != "806b07b08faa21cfbdae93027904f80174679402" {
		t.Errorf("Error checking checked out Git version, found: \"%s\"\n", string(v.Core()))
	}
	if err != nil {
		t.Error(err)
	}

}

func TestGitCheckLocal(t *testing.T) {
	// Verify repo.CheckLocal fails for non-Git directories.
	// TestGit is already checking on a valid repo
	tempDir, err := ioutil.TempDir("", "go-vcs-git-tests")
	if err != nil {
		t.Error(err)
	}
	defer func() {
		err = os.RemoveAll(tempDir)
		if err != nil {
			t.Error(err)
		}
	}()

	gitReader, _ := NewGitReader("", tempDir)
	exists, err := gitReader.Exists(Wkspc)
	if exists == true {
		t.Error("Git Exists is not correctly identifying non-Git pkg/repo")
	}

	// Test NewReader when there's no local. This should simply provide a working
	// instance without error based on looking at the remote localtion.
	_, nrerr := NewReader("https://github.com/Masterminds/VCSTestRepo", tempDir+"/VCSTestRepo")
	if nrerr != nil {
		t.Error(nrerr)
	}
}
