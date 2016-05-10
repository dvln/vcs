package vcs

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/dvln/out"
	"github.com/dvln/util/file"
)

// Canary test to ensure GitReader implements the Reader interface.
var _ Reader = &GitReader{}

// To verify git is working we perform intergration testing
// with a known git service.

// This tests non-bare git repo's with the various bits of git functionality within the
// VCS package... at least those items applicable to non-bare git repo's which is pretty
// much all features
func TestGit(t *testing.T) {
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
	testClone := filepath.Join(tempDir, "VCSTestRepo")

	mirror := true
	gitGetter, err := NewGitGetter("https://github.com/dvln/git-test-repo", "", testClone, !mirror)
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
	if gitGetter.LocalRepoPath() != testClone {
		t.Error("Local disk location not set properly")
	}

	// Do an initial mirror clone.
	results, err := gitGetter.Get()
	if err != nil {
		t.Fatalf("Unable to clone Git repo using VCS reader Get(). Err was: %s, details:\n%s", err, results)
	}

	// Verify Git repo exists in the workspace
	path, _, err := gitGetter.Exists(LocalPath)
	if err != nil {
		t.Fatalf("Existence check failed on git repo: %s", err)
	}
	if path == "" {
		t.Error("Problem seeing if Git repo exists in workspace")
	}

	// Test internal lookup mechanism used outside of Git specific functionality.
	ltype, err := DetectVcsFromFS(testClone)
	if err != nil {
		t.Error("detectVcsFromFS unable to detect Git repo")
	}
	if ltype != Git {
		t.Errorf("detectVcsFromFS detected %s instead of Git type", ltype)
	}

	// Test NewReader on existing checkout. This should simply provide a working
	// instance without error based on looking at the local directory.
	gitReader, err := NewReader("https://github.com/dvln/git-test-repo", testClone)
	if err != nil {
		t.Fatal(err)
	}
	if gitReader == nil {
		t.Fatal("gitReader interface was unexpectedly nil")
	}

	// See if the new git VCS reader was instantiated in the workspace
	path, _, err = gitReader.Exists(LocalPath)
	if err != nil {
		t.Errorf("Existence check failed on git repo: %s", err)
	}
	if path == "" {
		t.Error("The git reader was not correctly instantiated in the workspace")
	}

	// Perform an update operation
	gitUpdater, err := NewUpdater("https://github.com/dvln/git-test-repo", "origin", testClone, !mirror, RebaseFalse, nil)
	if err != nil {
		t.Fatal(err)
	}
	results, err = gitUpdater.Update()
	if err != nil {
		t.Fatalf("Failed to run git update, error: %s, results:\n%s", err, results)
	}

	// Set the version (checkout) using a short sha1 that should exist
	results, err = gitReader.RevSet("3f690c9")
	if err != nil {
		t.Fatalf("Unable to update Git repo version. Err was: %s, results:\n%s", err, results)
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

	// Install a git hook, verify existence, then remove it, verify gone
	gitHookMgr, err := NewHookMgr(testClone)
	if err != nil {
		t.Fatalf("Failed to set up a new git hook manager, err:\n %s\n", err)
	}
	// First do a run with an invalid source path, insure that fails (this
	// is not a bare repo so that will not be found as it doesn't exist)
	hookSrc := filepath.Join(testClone, "hooks", "pre-push.sample")
	link := true
	hookLinkPath, err := gitHookMgr.Install(hookSrc, "pre-push", link)
	if err == nil {
		t.Fatal("Hook install of link should have failed as target is not there")
	}
	// Now do a run with the real non-bare path that should exist...
	hookSrc = filepath.Join(testClone, ".git", "hooks", "pre-push.sample")
	hookLinkPath, err = gitHookMgr.Install(hookSrc, "pre-push", link)
	if err != nil {
		t.Fatalf("Failed to install git hook symlink, error:\n%s\n", err)
	}
	fileInfo, err := os.Lstat(hookLinkPath)
	if err != nil {
		t.Fatalf("Should have set up a hook symlink but os.Lstat failed, err: %s\n", err)
	}
	if fileInfo.Mode()&os.ModeSymlink == 0 {
		t.Fatal("Should have created a symlink but file created was not a symlink")
	}
	originFile, err := os.Readlink(hookLinkPath)
	if err != nil {
		t.Fatalf("Should have created a symlink but failed to resolve symlink, err: %s", err)
	}
	if originFile != hookSrc {
		t.Fatalf("Installed hook symlink not pointing correctly:\n  found: %s\n  need: %s\n", originFile, hookSrc)
	}
	err = gitHookMgr.Remove("pre-push")
	if err != nil {
		t.Fatalf("Failed to remove hook symlink, err: %s", err)
	}
	if there, err := file.Exists(hookLinkPath); err != nil || there {
		t.Fatalf("Removal of hook symlink seems to have failed (err: %s, there: %s)\n", err, there)
	}

	// Now lets try a copy type hook install and removal
	hookCopyPath, err := gitHookMgr.Install(hookSrc, "pre-push", !link)
	if err != nil {
		t.Fatalf("Failed to install (copy) git hook\n  src:%s\n  tgt:%s\n  err: %s\n", hookSrc, hookCopyPath, err)
	}
	fileInfo, err = os.Stat(hookCopyPath)
	if err != nil {
		t.Fatalf("Should have just copied file but os.Stat failed, err: %s\n", err)
	}
	srcFileInfo, err := os.Stat(hookSrc)
	if err != nil {
		t.Fatalf("Should have copied a symlink but os.Stat failed, err: %s\n", err)
	}
	if fileInfo.Size() != srcFileInfo.Size() {
		t.Fatal("File size of copied file not matching source")
	}
	err = gitHookMgr.Remove("pre-push")
	if err != nil {
		t.Fatalf("Failed to remove hook file, err: %s", err)
	}
	if there, err := file.Exists(hookCopyPath); err != nil || there {
		t.Fatalf("Removal of hook file seems to have failed (err: %s, there: %s)\n", err, there)
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
	gitGetter, err := NewGitGetter("https://github.com/dvln/git-test-repo", "", tempDir+sep+"VCSTestRepo", mirror)
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
	if gitGetter.LocalRepoPath() != tempDir+sep+"VCSTestRepo" {
		t.Error("Local disk location not set properly")
	}

	// Do an initial clone.
	_, err = gitGetter.Get()
	if err != nil {
		t.Fatalf("Unable to clone Git repo using VCS reader Get(). Err was %s", err)
	}

	// Verify Git repo exists in the workspace
	path, _, err := gitGetter.Exists(LocalPath)
	if err != nil {
		t.Errorf("Existence check failed on git repo: %s", err)
	}
	if path == "" {
		t.Error("Problem seeing if Git repo exists in workspace")
	}

	// Test internal lookup mechanism used outside of Git specific functionality.
	ltype, err := DetectVcsFromFS(tempDir + sep + "VCSTestRepo")
	if err != nil {
		t.Error("detectVcsFromFS unable to detect Git repo")
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
	path, _, err = gitReader.Exists(LocalPath)
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
	gitUpdater, err := NewUpdater("https://github.com/dvln/git-test-repo", "origin", tempDir+sep+"VCSTestRepo", mirror, RebaseFalse, nil)
	if err != nil {
		t.Fatal(err)
	}
	_, err = gitUpdater.Update()
	if err != nil {
		t.Error(err)
	}
	runDir := gitUpdater.LocalRepoPath()
	runOpt := "-C"

	// See if the branch exists (should exist, this is a mirror)
	result, err := run(gitTool, runOpt, runDir, "rev-parse", "--verify", "testbr1")
	if err != nil {
		t.Fatalf("Failed to detect local testbr1, should be there: %s\n%s", err, result.Output)
	}

	// Perform specific fetch operations on one ref, delete another ref
	refs := make(map[string]RefOp)
	refs["refs/heads/master"] = RefFetch
	refs["refs/heads/testbr1"] = RefDelete
	gitUpdater2, err := NewUpdater("https://github.com/dvln/git-test-repo", "origin", tempDir+sep+"VCSTestRepo", mirror, RebaseFalse, refs)
	if err != nil {
		t.Fatal(err)
	}
	_, err = gitUpdater2.Update()
	if err != nil {
		t.Error(err)
	}

	// See if the branch no longer exists (should gone at this point)
	_, err = run(gitTool, runOpt, runDir, "rev-parse", "--verify", "testbr1")
	if err == nil {
		t.Fatalf("The testbr1 branch should have been deleted, it's still still there...")
	}
}

// This tests cloning into a mirror repo, verifying it exists, then re-cloning into the same
// directory/path/etc but wanting a regular clone (not mirror/bare) and seeing that it blows
// away the old clone and brings in the new "regular" clone on the 2nd pass.
func TestGitFlipCloneType(t *testing.T) {
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
	gitGetter, err := NewGitGetter("https://github.com/dvln/git-test-repo", "", tempDir+sep+"VCSTestRepo", mirror)
	if err != nil {
		t.Errorf("Unable to instantiate new Git VCS reader, Err: %s", err)
	}

	if gitGetter.Vcs() != Git {
		t.Error("Git is detecting the wrong type")
	}

	// Check the basic getters.
	if gitGetter.Remote() != "https://github.com/dvln/git-test-repo" {
		t.Error("Remote not set properly")
	}
	if gitGetter.LocalRepoPath() != tempDir+sep+"VCSTestRepo" {
		t.Error("Local disk location not set properly")
	}

	// Do an initial mirror clone
	_, err = gitGetter.Get()
	if err != nil {
		t.Fatalf("Unable to clone Git repo using VCS reader Get(). Err was %s", err)
	}

	// Verify Git repo exists in the workspace
	path, _, err := gitGetter.Exists(LocalPath)
	if err != nil || path == "" {
		t.Errorf("Existence check failed on git repo: %s", err)
	}

	// In this case we're going with a regular clone, not a mirror clone,
	// but we already have a mirror clone there, should detect, remove and
	// bring in a fresh "regular" clone for us.  What could go wrong?
	gitGetter2, err := NewGitGetter("https://github.com/dvln/git-test-repo", "", tempDir+sep+"VCSTestRepo", !mirror)
	if err != nil {
		t.Errorf("Unable to instantiate second Git VCS reader, Err: %s", err)
	}
	// Do an fresh non-mirror "regular" clone, mirror clone exists
	results, err := gitGetter2.Get()
	if err != nil {
		t.Fatalf("Unable to non-mirror clone Git repo over existing mirror clone using VCS reader Get(). Err was %s", err)
	}
	resultsStr := fmt.Sprintf("%s", results)
	if strings.Contains(resultsStr, "remote update") || strings.Contains(resultsStr, "--mirror") {
		t.Fatalf("Second Get() run, which should have overwritten a mirror clone, does not appear to have worked, results:\n%s", resultsStr)
	}

	// Verify Git repo exists in the workspace
	path, _, err = gitGetter.Exists(LocalPath)
	if err != nil || path == "" {
		t.Fatalf("Existence check failed on git repo: %s", err)
	}
	if _, err = os.Stat(tempDir + sep + "VCSTestRepo" + sep + ".git"); os.IsNotExist(err) {
		t.Fatalf("Existence check failed on what should be a non-bare git repo: %s", err)
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
	path, _, err := gitReader.Exists(LocalPath)
	if err == nil {
		t.Error("Existence check should have indicated not exists for git repo, but did not")
	}
	if path != "" {
		t.Error("Git Exists is not correctly identifying non-Git pkg/repo")
	}

	if !out.IsError(err, ErrNoExist) {
		t.Fatalf("Git Exists is not correctly identifying repo does not exist (looking for ErrNoExist): %s", err)
	}
	if !out.IsError(err, nil, 4500) {
		t.Fatalf("Git Exists is not correctly identifying repo does not exist, error (looking for code 4500): %s", err)
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
	gitGetter, err := NewGitGetter(url1, "", tempDir, !mirror)
	if err != nil {
		t.Fatalf("Failed to initialize new Git getter, error: %s", err)
	}
	path, _, err = gitGetter.Exists(Remote)
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
	gitGetter, err = NewGitGetter(badurl1, "", tempDir, !mirror)
	if err != nil {
		t.Fatalf("Failed to initialize 1st \"bad\" Git getter, init should work, error: %s", err)
	}
	path, results, err := gitGetter.Exists(Remote)
	if err == nil {
		t.Fatalf("Failed to detect an error scanning for 1st bad VCS location (loc: %s), error: nil, results:\n%s", badurl1, results)
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
			t.Error("Failed to remove temp workspace that should have existed, err:", err)
		}
	}()

	mirror := true
	gitGetter, err := NewGetter("https://github.com/dvln/git-test-repo", "", tempDir+sep+"VCSTestRepo", mirror)
	if err != nil {
		t.Fatalf("Unable to instantiate new Git VCS reader, Err: %s", err)
	}

	// Do a clone
	_, err = gitGetter.Get()
	if err != nil {
		t.Fatalf("Unable to clone Git repo using VCS reader Get(). Err was %s", err)
	}

	// Do a clone again, it should end up detecting the clone and doing a remote update
	results, err := gitGetter.Get()
	if err != nil {
		t.Fatalf("Unable to clone over existing Git repo using VCS reader Get(). Err was %s", err)
	}
	resultsStr := fmt.Sprintf("%s", results)
	if !strings.Contains(resultsStr, "remote update") {
		t.Fatalf("Appears 2nd clone didn't switch to using a remote update, results:\n%s", resultsStr)
	}

	// Perform an update operation
	gitUpdater, err := NewUpdater("https://github.com/dvln/git-test-repo", "origin", tempDir+sep+"VCSTestRepo", mirror, RebaseFalse, nil)
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
	gitUpdater2, err := NewUpdater("https://github.com/dvln/git-test-repo", "origin", tempDir+sep+"VCSTestRepo", mirror, RebaseFalse, refs)
	if err != nil {
		t.Fatal(err)
	}
	_, err = gitUpdater2.Update()
	if err != nil {
		t.Error(err)
	}
}
