package vcs

import (
	"io/ioutil"
	"os"
	"testing"
)

// Canary test to ensure BzrReader implements the Reader interface.
var _ Reader = &BzrReader{}

// To verify bzr is working we perform intergration testing
// with a known bzr service.

func TestBzr(t *testing.T) {

	tempDir, err := ioutil.TempDir("", "go-vcs-bzr-tests")
	if err != nil {
		t.Error(err)
	}
	defer func() {
		err = os.RemoveAll(tempDir)
		if err != nil {
			t.Error(err)
		}
	}()

	bzrReader, err := NewBzrReader("https://launchpad.net/govcstestbzrrepo", tempDir+"/govcstestbzrrepo")
	if err != nil {
		t.Errorf("Unable to instantiate new Bzr VCS reader, Err: %s", err)
	}

	if bzrReader.Vcs() != Bzr {
		t.Error("Bzr is detecting the wrong type")
	}

	// Check the basic getters.
	if bzrReader.Remote() != "https://launchpad.net/govcstestbzrrepo" {
		t.Error("Remote not set properly")
	}
	if bzrReader.WkspcPath() != tempDir+"/govcstestbzrrepo" {
		t.Error("Local disk location not set properly")
	}

	// Do an initial clone.
	_, err = bzrReader.Get()
	if err != nil {
		t.Errorf("Unable to clone Bzr repo. Err was %s", err)
	}

	// Verify Bzr repo is a Bzr repo
	exists, err := bzrReader.Exists(Wkspc)
	if err != nil {
		t.Errorf("Existence check failed on Bzr repo: %s", err)
	}
	if exists == false {
		t.Error("Problem checking out repo via Bzr Exists is not working")
	}

	// Test internal lookup mechanism used outside of Bzr specific functionality.
	ltype, err := DetectVcsFromFS(tempDir + "/govcstestbzrrepo")
	if err != nil {
		t.Error("detectVcsFromFS unable to Bzr repo")
	}
	if ltype != Bzr {
		t.Errorf("detectVcsFromFS detected %s instead of Bzr type", ltype)
	}

	// Test NewReader on existing checkout. This should simply provide a working
	// instance without error based on looking at the local directory.
	nbzrReader, nrerr := NewReader("https://launchpad.net/govcstestbzrrepo", tempDir+"/govcstestbzrrepo")
	if nrerr != nil {
		t.Error(nrerr)
	}
	// Verify the thing exists in the workspace
	exists, err = nbzrReader.Exists(Wkspc)
	if err != nil {
		t.Errorf("Existence check failed on Bzr repo: %s", err)
	}
	if exists == false {
		t.Error("Don't see the new bzr repo in the workspace")
	}

	output, err := bzrReader.RevSet("2")
	if err != nil {
		t.Errorf("Unable to update Bzr repo version. Err was %s, output was:\n%s", err, output)
	}

	// Use Version to verify we are on the right version.
	v, _, err := bzrReader.RevRead(CoreRev)
	if string(v[0].Core()) != "2" {
		t.Error("Error checking checked out Bzr version")
	}
	if err != nil {
		t.Error(err)
	}

	// Perform an update.
	_, err = bzrReader.Update()
	if err != nil {
		t.Error(err)
	}

	v, _, err = bzrReader.RevRead(CoreRev)
	if string(v[0].Core()) != "3" {
		t.Error("Error checking checked out Bzr version")
	}
	if err != nil {
		t.Error(err)
	}

}

func TestBzrExists(t *testing.T) {
	// Verify bzrReader.Exists fails for non-Bzr directories.
	// TestBzr is already checking on a valid repo
	tempDir, err := ioutil.TempDir("", "go-vcs-bzr-tests")
	if err != nil {
		t.Error(err)
	}
	defer func() {
		err = os.RemoveAll(tempDir)
		if err != nil {
			t.Error(err)
		}
	}()

	bzrReader, _ := NewBzrReader("", tempDir)
	exists, err := bzrReader.Exists(Wkspc)
	if err != nil {
		t.Errorf("Existence check failed on Bzr repo: %s", err)
	}
	if exists == true {
		t.Error("Bzr Exists does detects bzr repo where one is not")
	}

	// Test NewReader when there's no local. This should simply provide a working
	// instance without error based on looking at the remote localtion.
	_, nrerr := NewReader("https://launchpad.net/govcstestbzrrepo", tempDir+"/govcstestbzrrepo")
	if nrerr != nil {
		t.Error(nrerr)
	}
}
