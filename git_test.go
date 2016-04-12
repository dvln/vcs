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

// This tests non-bare git repo's with the various bits of git functionality within the
// VCS package... at least those items applicable to non-bare git repo's which is pretty
// much all features
func TestGit(t *testing.T) {
	sep := string(os.PathSeparator)
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

	mirror := true
	gitGetter, err := NewGitGetter("https://github.com/Masterminds/VCSTestRepo", tempDir+sep+"VCSTestRepo", !mirror)
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
	if gitGetter.WkspcPath() != tempDir+sep+"VCSTestRepo" {
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
	ltype, err := DetectVcsFromFS(tempDir + sep + "VCSTestRepo")
	if err != nil {
		t.Error("detectVcsFromFS unable to Git repo")
	}
	if ltype != Git {
		t.Errorf("detectVcsFromFS detected %s instead of Git type", ltype)
	}

	// Test NewReader on existing checkout. This should simply provide a working
	// instance without error based on looking at the local directory.
	gitReader, err := NewReader("https://github.com/Masterminds/VCSTestRepo", tempDir+sep+"VCSTestRepo")
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
	gitUpdater, err := NewUpdater("https://github.com/Masterminds/VCSTestRepo", tempDir+sep+"VCSTestRepo", !mirror, RebaseFalse)
	if err != nil {
		t.Error(err)
	}
	_, err = gitUpdater.Update()
	if err != nil {
		t.Error(err)
	}

	// Set the version (checkout) using a short sha1 that should exist
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

// This tests bare git repo's with the various bits of git functionality within the
// VCS package... at least those items applicable to bare git repo's (which is more
// limited)
func TestBareGit(t *testing.T) {
	sep := string(os.PathSeparator)
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

	mirror := true
	gitGetter, err := NewGitGetter("https://github.com/Masterminds/VCSTestRepo", tempDir+sep+"VCSTestRepo", mirror)
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
	if gitGetter.WkspcPath() != tempDir+sep+"VCSTestRepo" {
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
	ltype, err := DetectVcsFromFS(tempDir + sep + "VCSTestRepo")
	if err != nil {
		t.Error("detectVcsFromFS unable to Git repo")
	}
	if ltype != Git {
		t.Errorf("detectVcsFromFS detected %s instead of Git type", ltype)
	}

	// Test NewReader on existing checkout. This should simply provide a working
	// instance without error based on looking at the local directory.
	gitReader, err := NewReader("https://github.com/Masterminds/VCSTestRepo", tempDir+sep+"VCSTestRepo")
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

	// Use RevRead to read a version, see if that works
	v, _, err := gitReader.RevRead(CoreRev, "806b07b08faa21cfbdae93027904f80174679402")
	if string(v[0].Core()) != "806b07b08faa21cfbdae93027904f80174679402" {
		t.Errorf("Error checking checked out Git version, found: \"%s\"\n", string(v[0].Core()))
	}
	if err != nil {
		t.Error(err)
	}

	// Perform an update operation
	gitUpdater, err := NewUpdater("https://github.com/Masterminds/VCSTestRepo", tempDir+sep+"VCSTestRepo", mirror, RebaseFalse)
	if err != nil {
		t.Error(err)
	}
	_, err = gitUpdater.Update()
	if err != nil {
		t.Error(err)
	}
}

// TestgitExists focuses on existence checks
func TestGitExists(t *testing.T) {
	sep := string(os.PathSeparator)
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
	_, nrerr := NewReader("https://github.com/Masterminds/VCSTestRepo", tempDir+sep+"VCSTestRepo")
	if nrerr != nil {
		t.Error(nrerr)
	}

	// Try remote Git existence checks via a Getter
	url1 := "github.com/dvln/vcs"
	mirror := true
	gitGetter, err := NewGitGetter(url1, tempDir, !mirror)
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

	badurl1 := "github.com/dvln/notexistvcs"
	gitGetter, err = NewGitGetter(badurl1, tempDir, !mirror)
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
}
