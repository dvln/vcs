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

	gitGetter, err := NewGitGetter("https://github.com/Masterminds/VCSTestRepo", tempDir+"/VCSTestRepo")
	if err != nil {
		t.Errorf("Unable to instantiate new Git VCS reader, Err: %s", err)
	}

	if gitGetter.Vcs() != Git {
		t.Error("Git is detecting the wrong type")
	}

	// Check the basic getters.
	if gitGetter.Remote() != "https://github.com/Masterminds/VCSTestRepo" {
		t.Error("Remote not set properly")
	}
	if gitGetter.WkspcPath() != tempDir+"/VCSTestRepo" {
		t.Error("Local disk location not set properly")
	}

	// Do an initial clone.
	_, err = gitGetter.Get()
	if err != nil {
		t.Errorf("Unable to clone Git repo using VCS reader Get(). Err was %s", err)
	}

	// Verify Git repo exists in the workspace
	path, err := gitGetter.Exists(Wkspc)
	if err != nil {
		t.Errorf("Existence check failed on git repo: %s", err)
	}
	if path == "" {
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
	gitReader, err := NewReader("https://github.com/Masterminds/VCSTestRepo", tempDir+"/VCSTestRepo")
	if err != nil {
		t.Error(err)
	}
	// See if the new git VCS reader was instantiated in the workspace
	path, err = gitReader.Exists(Wkspc)
	if err != nil {
		t.Errorf("Existence check failed on git repo: %s", err)
	}
	if path == "" {
		t.Error("The git reader was not correctly instantiated in the workspace")
	}

	// Perform an update operation
	gitUpdater, err := NewUpdater("https://github.com/Masterminds/VCSTestRepo", tempDir+"/VCSTestRepo")
	if err != nil {
		t.Error(err)
	}
	_, err = gitUpdater.Update()
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
	if string(v[0].Core()) != "806b07b08faa21cfbdae93027904f80174679402" {
		t.Errorf("Error checking checked out Git version, found: \"%s\"\n", string(v[0].Core()))
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
	if string(v[0].Core()) != "806b07b08faa21cfbdae93027904f80174679402" {
		t.Errorf("Error checking checked out Git version, found: \"%s\"\n", string(v[0].Core()))
	}
	if err != nil {
		t.Error(err)
	}

}

func TestGitExists(t *testing.T) {
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
	path, err := gitReader.Exists(Wkspc)
	if path != "" {
		t.Error("Git Exists is not correctly identifying non-Git pkg/repo")
	}

	// Test NewReader when there's no local. This should simply provide a working
	// instance without error based on looking at the remote localtion.
	_, nrerr := NewReader("https://github.com/Masterminds/VCSTestRepo", tempDir+"/VCSTestRepo")
	if nrerr != nil {
		t.Error(nrerr)
	}

	// Try remote Git existence checks via a Getter
	url1 := "github.com/dvln/vcs"
	gitGetter, err := NewGitGetter(url1, tempDir)
	if err != nil {
		t.Fatalf("Failed to initialize new Git getter, error: %s", err)
	}
	path, err = gitGetter.Exists(Remote)
	if err != nil {
		t.Fatalf("Failed to find remote repo that should exist (URL: %s), error: %s", url1, err)
	}
	if !(path == "https://github.com/dvln/vcs" || path == "git://github.com/dvln/vcs") {
		t.Fatalf("Exists failed to return remote path with correct scheme (URL: %s), found: %s", url1, path)
	}

    if testing.Short() {
        t.Skip("skipping remaining existence checks in short mode.")
		return
    }

	url2 := "https://github.com/dvln/vcs"
	gitGetter, err = NewGitGetter(url2, tempDir)
	if err != nil {
		t.Fatalf("Failed to initialize new Git getter, error: %s", err)
	}
	path, err = gitGetter.Exists(Remote)
	if err != nil {
		t.Fatalf("Failed to find remote repo that should exist (URL: %s), error: %s", url2, err)
	}
	if path != url2 {
		t.Fatalf("Exists failed to return matching URL path (URL: %s), found: %s", url2, path)
	}

	badurl1 := "github.com/dvln/notexistvcs"
	gitGetter, err = NewGitGetter(badurl1, tempDir)
	if err != nil {
		t.Fatalf("Failed to initialize 1st \"bad\" Git getter, init should work, error: %s", err)
	}
	path, err = gitGetter.Exists(Remote)
	if err == nil {
		t.Fatalf("Failed to detect an error scanning for 1st bad VCS location (loc: %s), error: %s", badurl1, err)
	}
	if path != "" {
		t.Fatalf("Unexpectedly found a repo when shouldn't have (URL: %s), found path: %s", badurl1, err)
	}

	badurl2 := "https://github.com/dvln/notexistvcs"
	gitGetter, err = NewGitGetter(badurl2, tempDir)
	if err != nil {
		t.Fatalf("Failed to initialize 2nd \"bad\" Git getter, init should work, error: %s", err)
	}
	path, err = gitGetter.Exists(Remote)
	if err == nil {
		t.Fatalf("Failed to detect an error scanning for 2nd bad VCS location (loc: %s), error: %s", badurl2, err)
	}
	if path != "" {
		t.Fatalf("Unexpectedly found a repo when shouldn't have (URL: %s), found path: %s", badurl2, err)
	}
}
