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

	svnReader, err := NewSvnReader("https://github.com/Masterminds/VCSTestRepo/trunk", tempDir+"/VCSTestRepo")
	if err != nil {
		t.Errorf("Unable to instantiate new SVN VCS reader, Err: %s", err)
	}

	if svnReader.Vcs() != Svn {
		t.Error("Svn is detecting the wrong type")
	}

	// Check the basic getters.
	if svnReader.Remote() != "https://github.com/Masterminds/VCSTestRepo/trunk" {
		t.Error("Remote not set properly")
	}
	if svnReader.WkspcPath() != tempDir+"/VCSTestRepo" {
		t.Error("Local disk location not set properly")
	}

	//Logger = log.New(os.Stdout, "", log.LstdFlags)

	// Do an initial checkout.
	_, err = svnReader.Get()
	if err != nil {
		t.Errorf("Unable to checkout SVN repo. Err was %s", err)
	}

	// Verify SVN repo is a SVN repo
	exists, err := svnReader.Exists(Wkspc)
	if err != nil {
		t.Errorf("Existence check failed on svn repo: %s", err)
	}
	if exists == false {
		t.Error("Problem checking if SVN repo Exists in the workspace")
	}

	// Verify an incorrect remote is caught when NewSvnReader is used on an existing location
	_, nrerr := NewSvnReader("https://github.com/Masterminds/VCSTestRepo/unknownbranch", tempDir+"/VCSTestRepo")
	if nrerr != ErrWrongRemote {
		t.Error("ErrWrongRemote was not triggered for SVN")
	}

	// Test internal lookup mechanism used outside of Hg specific functionality.
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
	// nsvnReader, nrerr := NewReader("https://github.com/Masterminds/VCSTestRepo/trunk", tempDir+"/VCSTestRepo")
	// if nrerr != nil {
	// 	t.Error(nrerr)
	// }
	// // Verify the right oject is returned. It will check the local repo type.
	// exists, err = nsvnReader.Exists(Wkspc)
	// if err != nil {
	// 	t.Errorf("Existence check failed on svn repo: %s", err)
	// }
	// if exists == false {
	// 	t.Error("Wrong version returned from NewReader")
	// }

	// Update the version to a previous version.
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
	_, err = svnReader.Update()
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
	exists, err := svnReader.Exists(Wkspc)
	if err != nil {
		t.Errorf("Existence check failed on svn repo: %s", err)
	}
	if exists == true {
		t.Error("SVN repo exists check incorrectlyi indicating existence")
	}

	// Test NewReader when there's no local. This should simply provide a working
	// instance without error based on looking at the remote localtion.
	_, nrerr := NewReader("https://github.com/Masterminds/VCSTestRepo/trunk", tempDir+"/VCSTestRepo")
	if nrerr != nil {
		t.Error(nrerr)
	}
}
