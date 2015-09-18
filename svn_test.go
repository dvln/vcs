package vcs

import (
	"io/ioutil"
	//"log"
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
		t.Error(err)
	}
	defer func() {
		err = os.RemoveAll(tempDir)
		if err != nil {
			t.Error(err)
		}
	}()

	svnGetter, err := NewSvnGetter("https://github.com/Masterminds/VCSTestRepo/trunk", tempDir+"/VCSTestRepo")
	if err != nil {
		t.Errorf("Unable to instantiate new SVN VCS reader, Err: %s", err)
	}

	if svnGetter.Vcs() != Svn {
		t.Error("Svn is detecting the wrong type")
	}

	// Check the basic getters.
	if svnGetter.Remote() != "https://github.com/Masterminds/VCSTestRepo/trunk" {
		t.Error("Remote not set properly")
	}
	if svnGetter.WkspcPath() != tempDir+"/VCSTestRepo" {
		t.Error("Local disk location not set properly")
	}

	// Do an initial checkout.
	_, err = svnGetter.Get()
	if err != nil {
		t.Errorf("Unable to checkout SVN repo. Err was %s", err)
	}

	// Verify SVN repo is a SVN repo
	path, err := svnGetter.Exists(Wkspc)
	if err != nil {
		t.Errorf("Existence check failed on svn repo: %s", err)
	}
	if path == "" {
		t.Error("Problem checking if SVN repo Exists in the workspace")
	}

	// Verify an incorrect remote is caught when NewSvnReader is used on an existing location
	_, err = NewSvnReader("https://github.com/Masterminds/VCSTestRepo/unknownbranch", tempDir+"/VCSTestRepo")
	if err != ErrWrongRemote {
		t.Error("ErrWrongRemote was not triggered for SVN")
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
	// svnReader, err := NewReader("https://github.com/Masterminds/VCSTestRepo/trunk", tempDir+"/VCSTestRepo")
	// if err != nil {
	// 	t.Error(err)
	// }
	// // Verify the right oject is returned. It will check the local repo type.
	// path, err = svnReader.Exists(Wkspc)
	// if err != nil {
	// 	t.Errorf("Existence check failed on svn repo: %s", err)
	// }
	// if path == "" {
	// 	t.Error("Wrong version returned from NewReader")
	// }

	// Change the version in the workspace to a previous version.
	svnReader, err := NewSvnReader("https://github.com/Masterminds/VCSTestRepo/trunk", tempDir+"/VCSTestRepo")
	if err != nil {
		t.Errorf("Unable to instantiate new SVN VCS reader, Err: %s", err)
	}
	output, err := svnReader.RevSet("r2")
	if err != nil {
		t.Errorf("Unable to update SVN repo version. Err was %s, output:\n%s", err, output)
	}

	// Use RevRead to verify we are on the right version.
	v, _, err := svnReader.RevRead(CoreRev)
	if string(v[0].Core()) != "2" {
		t.Error("Error checking checked SVN out version")
	}
	if err != nil {
		t.Error(err)
	}

	// Perform an update which should take up back to the latest version.
	svnUpdater, err := NewSvnUpdater("https://github.com/Masterminds/VCSTestRepo/trunk", tempDir+"/VCSTestRepo")
	if err != nil {
		t.Error(err)
	}
	_, err = svnUpdater.Update()
	if err != nil {
		t.Error(err)
	}

	// Make sure we are on a newer version because of the update.
	v, _, err = svnReader.RevRead(CoreRev)
	if string(v[0].Core()) == "2" {
		t.Error("Error with version. Still on old version. Update failed")
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
		t.Error(err)
	}
	defer func() {
		err = os.RemoveAll(tempDir)
		if err != nil {
			t.Error(err)
		}
	}()

	svnReader, _ := NewSvnReader("", tempDir)
	path, err := svnReader.Exists(Wkspc)
	if err != nil {
		t.Errorf("Existence check failed on svn repo: %s", err)
	}
	if path != "" {
		t.Error("SVN repo exists check incorrectlyi indicating existence")
	}

	// Test NewReader when there's no local. This should simply provide a working
	// instance without error based on looking at the remote localtion.
	_, nrerr := NewReader("https://github.com/Masterminds/VCSTestRepo/trunk", tempDir+"/VCSTestRepo")
	if nrerr != nil {
		t.Error(nrerr)
	}

	// Try remote Svn existence checks via a Getter
	url1 := "github.com/Masterminds/VCSTestRepo/trunk"
	svnGetter, err := NewSvnGetter(url1, tempDir)
	if err != nil {
		t.Fatalf("Failed to initialize new Svn getter, error: %s", err)
	}
	path, err = svnGetter.Exists(Remote)
	if err != nil {
		t.Fatalf("Failed to find remote repo that should exist (URL: %s), error: %s", url1, err)
	}
	if path != "https://github.com/Masterminds/VCSTestRepo/trunk" {
		t.Fatalf("Exists failed to return remote path with correct scheme (URL: %s), found: %s", url1, path)
	}

    if testing.Short() {
        t.Skip("skipping remaining existence checks in short mode.")
		return
    }

	url2 := "https://github.com/Masterminds/VCSTestRepo/trunk"
	svnGetter, err = NewSvnGetter(url2, tempDir)
	if err != nil {
		t.Fatalf("Failed to initialize new Svn getter, error: %s", err)
	}
	path, err = svnGetter.Exists(Remote)
	if err != nil {
		t.Fatalf("Failed to find remote repo that should exist (URL: %s), error: %s", url2, err)
	}
	if path != url2 {
		t.Fatalf("Exists failed to return matching URL path (URL: %s), found: %s", url2, path)
	}

	badurl1 := "github.com/Masterminds/notexistVCSTestRepo/trunk"
	svnGetter, err = NewSvnGetter(badurl1, tempDir)
	if err != nil {
		t.Fatalf("Failed to initialize 1st \"bad\" Svn getter, init should work, error: %s", err)
	}
	path, err = svnGetter.Exists(Remote)
	if err == nil {
		t.Fatalf("Failed to detect an error scanning for 1st bad VCS location (loc: %s), error: %s", badurl1, err)
	}
	if path != "" {
		t.Fatalf("Unexpectedly found a repo when shouldn't have (URL: %s), found path: %s", badurl1, err)
	}

	badurl2 := "https://github.com/Masterminds/notexistVCSTestRepo/trunk"
	svnGetter, err = NewSvnGetter(badurl2, tempDir)
	if err != nil {
		t.Fatalf("Failed to initialize 2nd \"bad\" Svn getter, init should work, error: %s", err)
	}
	path, err = svnGetter.Exists(Remote)
	if err == nil {
		t.Fatalf("Failed to detect an error scanning for 2nd bad VCS location (loc: %s), error: %s", badurl2, err)
	}
	if path != "" {
		t.Fatalf("Unexpectedly found a repo when shouldn't have (URL: %s), found path: %s", badurl2, err)
	}
}
