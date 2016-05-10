package vcs

import (
	"io/ioutil"
	"os"
	"testing"
)

// To verify svn is working we perform intergration testing
// with a known svn service.

// Canary test to ensure SvnReader implements the VCS Reader interface.
var _ Reader = &SvnReader{}

func TestSvn(t *testing.T) {

	tempDir, err := ioutil.TempDir("", "go-vcs-svn-tests")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		err = os.RemoveAll(tempDir)
		if err != nil {
			t.Error(err)
		}
	}()

	svnGetter, err := NewSvnGetter("https://svn.code.sf.net/p/dvlnsvntest/code/trunk", "", tempDir+"/VCSTestRepo", false)
	if err != nil {
		t.Fatalf("Unable to instantiate new SVN VCS reader, Err: %s", err)
	}

	if svnGetter.Vcs() != Svn {
		t.Error("Svn is detecting the wrong type")
	}

	// Check the basic getters.
	if svnGetter.Remote() != "https://svn.code.sf.net/p/dvlnsvntest/code/trunk" {
		t.Error("Remote not set properly")
	}
	if svnGetter.LocalRepoPath() != tempDir+"/VCSTestRepo" {
		t.Error("Local disk location not set properly")
	}

	// Do an initial checkout.
	_, err = svnGetter.Get()
	if err != nil {
		t.Fatalf("Unable to checkout SVN repo. Err was %s", err)
	}

	// Verify SVN repo is a SVN repo
	path, _, err := svnGetter.Exists(LocalPath)
	if err != nil {
		t.Fatalf("Existence check failed on svn repo: %s", err)
	}
	if path == "" {
		t.Error("Problem checking if SVN repo Exists in the workspace")
	}

	// Verify an incorrect remote is caught when NewSvnReader is used on an existing location
	_, err = NewSvnReader("https://svn.code.sf.net/p/dvlnsvntest/code/unknownbranch", tempDir+"/VCSTestRepo")
	if err != ErrWrongRemote {
		t.Fatal("ErrWrongRemote was not triggered for SVN")
	}

	// Test internal lookup mechanism used outside of Svn specific functionality.
	ltype, err := DetectVcsFromFS(tempDir + "/VCSTestRepo")
	if err != nil {
		t.Error("detectVcsFromFS unable to Svn repo")
	}
	if ltype != Svn {
		t.Errorf("detectVcsFromFS detected %s instead of Svn type", ltype)
	}

	// Commenting out auto-detection tests for SVN. NewReader automatically detects
	// GitHub to be a Git repo and that's an issue for this test. Need an
	// SVN host that can autodetect from before using this test again.
	//
	// Test NewReader on existing checkout. This should simply provide a working
	// instance without error based on looking at the local directory.
	// svnReader, err := NewReader("https://svn.code.sf.net/p/dvlnsvntest/code/trunk", tempDir+"/VCSTestRepo")
	// if err != nil {
	// 	t.Fatal(err)
	// }
	// // Verify the right oject is returned. It will check the local repo type.
	// path, err = svnReader.Exists(LocalPath)
	// if err != nil {
	// 	t.Fatalf("Existence check failed on svn repo: %s", err)
	// }
	// if path == "" {
	// 	t.Error("Wrong version returned from NewReader")
	// }

	// Change the version in the workspace to a previous version.
	svnReader, err := NewSvnReader("https://svn.code.sf.net/p/dvlnsvntest/code/trunk", tempDir+"/VCSTestRepo")
	if err != nil {
		t.Fatalf("Unable to instantiate new SVN VCS reader, Err: %s", err)
	}
	results, err := svnReader.RevSet("r2")
	if err != nil {
		t.Errorf("Unable to update SVN repo version. Err was %s, results:\n%s", err, results)
	}

	// Use RevRead to verify we are on the right version.
	v, _, err := svnReader.RevRead(CoreRev)
	if string(v[0].Core()) != "2" {
		t.Errorf("Error reading SVN version after revset/checkout, expected \"2\", found: %s", string(v[0].Core()))
	}
	if err != nil {
		t.Error(err)
	}

	// Perform an update which should take up back to the latest version.
	mirror := true
	svnUpdater, err := NewSvnUpdater("https://svn.code.sf.net/p/dvlnsvntest/code/trunk", "", tempDir+"/VCSTestRepo", !mirror, RebaseFalse, nil)
	if err != nil {
		t.Fatal(err)
	}
	_, err = svnUpdater.Update()
	if err != nil {
		t.Error(err)
	}

	// Make sure we are on a newer version because of the update.
	v, _, err = svnReader.RevRead(CoreRev)
	if string(v[0].Core()) != "3" {
		t.Errorf("Unexpected version found after update, should be \"3\", found: %s", string(v[0].Core()))
	}
	if err != nil {
		t.Error(err)
	}
}

func TestSvnExists(t *testing.T) {
	// Verify svnReader.Exists fails for non-SVN directories.
	// TestSvn is already checking on a valid repo
	tempDir, err := ioutil.TempDir("", "go-vcs-svn-tests")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		err = os.RemoveAll(tempDir)
		if err != nil {
			t.Error(err)
		}
	}()

	svnReader, _ := NewSvnReader("", tempDir)
	path, _, err := svnReader.Exists(LocalPath)
	if err == nil {
		t.Error("Existence check should have indicated not exists for svn repo, but did not")
	}
	if path != "" {
		t.Fatal("SVN repo exists check incorrectlyi indicating existence")
	}

	// Test NewReader when there's no local. This should simply provide a working
	// instance without error based on looking at the remote localtion.
	_, err = NewReader("https://svn.code.sf.net/p/dvlnsvntest/code/trunk", tempDir+"/VCSTestRepo")
	if err != nil {
		t.Fatalf("Unable to instantiate new SVN VCS reader (using generic init), Err: %s", err)
	}

	// Try remote Svn existence checks via a Getter
	url1 := "svn.code.sf.net/p/dvlnsvntest/code/trunk"
	svnGetter, err := NewSvnGetter(url1, "", tempDir, false)
	if err != nil {
		t.Fatalf("Failed to initialize new Svn getter, error: %s", err)
	}
	path, _, err = svnGetter.Exists(Remote)
	if err != nil {
		t.Fatalf("Failed to find remote repo that should exist (URL: %s), error: %s", url1, err)
	}
	if path != "https://svn.code.sf.net/p/dvlnsvntest/code/trunk" {
		t.Fatalf("Exists failed to return remote path with correct scheme (URL: %s), found: %s", url1, path)
	}

	if testing.Short() {
		t.Skip("skipping remaining existence checks in short mode.")
		return
	}

	url2 := "https://svn.code.sf.net/p/dvlnsvntest/code/trunk"
	svnGetter, err = NewSvnGetter(url2, "", tempDir, false)
	if err != nil {
		t.Fatalf("Failed to initialize new Svn getter, error: %s", err)
	}
	path, _, err = svnGetter.Exists(Remote)
	if err != nil {
		t.Fatalf("Failed to find remote repo that should exist (URL: %s), error: %s", url2, err)
	}
	if path != url2 {
		t.Fatalf("Exists failed to return matching URL path (URL: %s), found: %s", url2, path)
	}

	badurl1 := "svn://svn.code.sf.net/p/notexistdvlnsvntest/code/trunk"
	svnGetter, err = NewSvnGetter(badurl1, "", tempDir, false)
	if err != nil {
		t.Fatalf("Failed to initialize 1st \"bad\" Svn getter, init should work, error: %s", err)
	}
	path, _, err = svnGetter.Exists(Remote)
	if err == nil {
		t.Fatalf("Failed to detect an error scanning for 1st bad VCS location (loc: %s), error: %s", badurl1, err)
	}
	if path != "" {
		t.Fatalf("Unexpectedly found a repo when shouldn't have (URL: %s), found path: %s", badurl1, err)
	}

	badurl2 := "https://svn.code.sf.net/p/notexistdvlnsvntest/code/trunk"
	svnGetter, err = NewSvnGetter(badurl2, "", tempDir, false)
	if err != nil {
		t.Fatalf("Failed to initialize 2nd \"bad\" Svn getter, init should work, error: %s", err)
	}
	path, _, err = svnGetter.Exists(Remote)
	if err == nil {
		t.Fatalf("Failed to detect an error scanning for 2nd bad VCS location (loc: %s), error: %s", badurl2, err)
	}
	if path != "" {
		t.Fatalf("Unexpectedly found a repo when shouldn't have (URL: %s), found path: %s", badurl2, err)
	}
}
