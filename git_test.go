package vcs

import (
	"io/ioutil"
	"os"
	"sync"
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
		t.Fatal(err)
	}
	defer func() {
		err = os.RemoveAll(tempDir)
		if err != nil {
			t.Error(err)
		}
	}()

	mirror := true
	gitGetter, err := NewGitGetter("https://github.com/dvln/git-test-repo", tempDir+sep+"VCSTestRepo", !mirror)
	if err != nil {
		t.Fatalf("Unable to instantiate new Git VCS reader, Err: %s", err)
	}

	if gitGetter.Vcs() != Git {
		t.Error("Git is detecting the wrong type")
	}

	// Check the basic getters.
	if gitGetter.Remote() != "https://github.com/dvln/git-test-repo" {
		t.Error("Remote not set properly")
	}
	if gitGetter.WkspcPath() != tempDir+sep+"VCSTestRepo" {
		t.Error("Local disk location not set properly")
	}

	// Do an initial clone.
	_, err = gitGetter.Get()
	if err != nil {
		t.Fatalf("Unable to clone Git repo using VCS reader Get(). Err was %s", err)
	}

	// Verify Git repo exists in the workspace
	path, err := gitGetter.Exists(Wkspc)
	if err != nil {
		t.Fatalf("Existence check failed on git repo: %s", err)
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
	gitReader, err := NewReader("https://github.com/dvln/git-test-repo", tempDir+sep+"VCSTestRepo")
	if err != nil {
		t.Fatal(err)
	}
	if gitReader == nil {
		t.Fatal("gitReader interface was unexpectedly nil")
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
	gitUpdater, err := NewUpdater("https://github.com/dvln/git-test-repo", tempDir+sep+"VCSTestRepo", !mirror, RebaseFalse, nil)
	if err != nil {
		t.Fatal(err)
	}
	_, err = gitUpdater.Update()
	if err != nil {
		t.Error(err)
	}

	// Set the version (checkout) using a short sha1 that should exist
	_, err = gitReader.RevSet("3f690c9")
	if err != nil {
		t.Errorf("Unable to update Git repo version. Err was %s", err)
	}

	// Use RevRead to verify we are on the right version.
	v, _, err := gitReader.RevRead(CoreRev)
	if string(v[0].Core()) != "3f690c91af378fbd09628f9833abb1c3d6828c5e" {
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
	_, err = gitReader.RevSet("28d488c8deda544076f56b279824657fa691ef01")
	if err != nil {
		t.Errorf("Unable to update Git repo version. Err was %s", err)
	}
	v, _, err = gitReader.RevRead(CoreRev)
	if string(v[0].Core()) != "28d488c8deda544076f56b279824657fa691ef01" {
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
	gitGetter, err := NewGitGetter("https://github.com/dvln/git-test-repo", tempDir+sep+"VCSTestRepo", mirror)
	if err != nil {
		t.Fatalf("Unable to instantiate new Git VCS reader, Err: %s", err)
	}

	if gitGetter.Vcs() != Git {
		t.Error("Git is detecting the wrong type")
	}

	// Check the basic getters.
	if gitGetter.Remote() != "https://github.com/dvln/git-test-repo" {
		t.Error("Remote not set properly")
	}
	if gitGetter.WkspcPath() != tempDir+sep+"VCSTestRepo" {
		t.Error("Local disk location not set properly")
	}

	// Do an initial clone.
	_, err = gitGetter.Get()
	if err != nil {
		t.Fatalf("Unable to clone Git repo using VCS reader Get(). Err was %s", err)
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
	gitReader, err := NewReader("https://github.com/dvln/git-test-repo", tempDir+sep+"VCSTestRepo")
	if err != nil {
		t.Fatal(err)
	}
	// See if the new git VCS reader was instantiated in the workspace
	path, err = gitReader.Exists(Wkspc)
	if err != nil {
		t.Fatalf("Existence check failed on git repo: %s", err)
	}
	if path == "" {
		t.Fatal("The git reader was not correctly instantiated in the workspace")
	}

	// Use RevRead to read a version, see if that works
	v, _, err := gitReader.RevRead(CoreRev, "28d488c8deda544076f56b279824657fa691ef01")
	if string(v[0].Core()) != "28d488c8deda544076f56b279824657fa691ef01" {
		t.Errorf("Error checking checked out Git version, found: \"%s\"\n", string(v[0].Core()))
	}
	if err != nil {
		t.Error(err)
	}

	// Perform a remote update class operation (ie: mirror update w/prune of deleted refs)
	gitUpdater, err := NewUpdater("https://github.com/dvln/git-test-repo", tempDir+sep+"VCSTestRepo", mirror, RebaseFalse, nil)
	if err != nil {
		t.Fatal(err)
	}
	_, err = gitUpdater.Update()
	if err != nil {
		t.Error(err)
	}
	runDir := gitUpdater.WkspcPath()
	runOpt := "-C"

	// See if the branch exists (should exist, this is a mirror)
	output, err := run("git", runOpt, runDir, "rev-parse", "--verify", "testbr1")
	if err != nil {
		t.Fatalf("Failed to detect local testbr1, should be there: %s\n%s", err, output)
	}

	// Perform specific fetch operations on one ref, delete another ref
	refs := make(map[string]RefOp)
	refs["refs/heads/master"] = RefFetch
	refs["refs/heads/testbr1"] = RefDelete
	gitUpdater2, err := NewUpdater("https://github.com/dvln/git-test-repo", tempDir+sep+"VCSTestRepo", mirror, RebaseFalse, refs)
	if err != nil {
		t.Fatal(err)
	}
	_, err = gitUpdater2.Update()
	if err != nil {
		t.Error(err)
	}

	// See if the branch no longer exists (should gone at this point)
	_, err = run("git", runOpt, runDir, "rev-parse", "--verify", "testbr1")
	if err == nil {
		t.Fatalf("The testbr1 branch should have been deleted, it's still still there...")
	}
}

// TestGitExists focuses on existence checks
func TestGitExists(t *testing.T) {
	sep := string(os.PathSeparator)
	// Verify repo.CheckLocal fails for non-Git directories.
	// TestGit is already checking on a valid repo
	tempDir, err := ioutil.TempDir("", "go-vcs-git-tests")
	if err != nil {
		t.Fatal(err)
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
	_, nrerr := NewReader("https://github.com/dvln/git-test-repo", tempDir+sep+"VCSTestRepo")
	if nrerr != nil {
		t.Fatal(nrerr)
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

func TestParallelGitGetUpd(t *testing.T) {
	var wg sync.WaitGroup
	wg.Add(6)
	for i := 1; i <= 6; i++ {
		go runGetUpd(t, &wg)
	}
	wg.Wait()
}

// runGetUpd is for multiple goroutine testing, look for race issues
func runGetUpd(t *testing.T, wg *sync.WaitGroup) {
	defer wg.Done()
	sep := string(os.PathSeparator)
	tempDir, err := ioutil.TempDir("", "go-vcs-git-tests")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		err = os.RemoveAll(tempDir)
		if err != nil {
			t.Error(err)
		}
	}()

	mirror := true
	gitGetter, err := NewGetter("https://github.com/dvln/git-test-repo", tempDir+sep+"VCSTestRepo", mirror)
	if err != nil {
		t.Fatalf("Unable to instantiate new Git VCS reader, Err: %s", err)
	}

	// Do a clone
	_, err = gitGetter.Get()
	if err != nil {
		t.Fatalf("Unable to clone Git repo using VCS reader Get(). Err was %s", err)
	}

	// Perform an update operation
	gitUpdater, err := NewUpdater("https://github.com/dvln/git-test-repo", tempDir+sep+"VCSTestRepo", mirror, RebaseFalse, nil)
	if err != nil {
		t.Fatal(err)
	}
	_, err = gitUpdater.Update()
	if err != nil {
		t.Error(err)
	}

	// Perform specific fetch operations on one ref, delete another ref
	refs := make(map[string]RefOp)
	refs["refs/heads/master"] = RefFetch
	refs["refs/heads/testbr1"] = RefDelete
	gitUpdater2, err := NewUpdater("https://github.com/dvln/git-test-repo", tempDir+sep+"VCSTestRepo", mirror, RebaseFalse, refs)
	if err != nil {
		t.Fatal(err)
	}
	_, err = gitUpdater2.Update()
	if err != nil {
		t.Error(err)
	}
}
