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

	bzrGetter, err := NewBzrGetter("https://launchpad.net/dvlnbzrtest", tempDir+"/govcstestbzrrepo", false)
	if err != nil {
		t.Errorf("Unable to instantiate new Bzr VCS reader, Err: %s", err)
	}

	if bzrGetter.Vcs() != Bzr {
		t.Error("Bzr is detecting the wrong type")
	}

	// Check the basic getters.
	if bzrGetter.Remote() != "https://launchpad.net/dvlnbzrtest" {
		t.Error("Remote not set properly")
	}
	if bzrGetter.WkspcPath() != tempDir+"/govcstestbzrrepo" {
		t.Error("Local disk location not set properly")
	}

	// Do an initial clone.
	_, err = bzrGetter.Get()
	if err != nil {
		t.Errorf("Unable to clone Bzr repo. Err was %s", err)
	}

	// Verify Bzr repo is a Bzr repo
	path, err := bzrGetter.Exists(Wkspc)
	if err != nil {
		t.Errorf("Existence check failed on Bzr repo: %s", err)
	}
	if path == "" {
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
	bzrReader, err := NewReader("https://launchpad.net/dvlnbzrtest", tempDir+"/govcstestbzrrepo")
	if err != nil {
		t.Error(err)
	}
	// Verify the thing exists in the workspace
	path, err = bzrReader.Exists(Wkspc)
	if err != nil {
		t.Errorf("Existence check failed on Bzr repo: %s", err)
	}
	if path == "" {
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
	mirror := true
	bzrUpdater, err := NewUpdater("https://launchpad.net/dvlnbzrtest", tempDir+"/govcstestbzrrepo", !mirror, RebaseFalse)
	if err != nil {
		t.Error(err)
	}
	_, err = bzrUpdater.Update()
	if err != nil {
		t.Error(err)
	}

	v, _, err = bzrReader.RevRead(CoreRev)
	if string(v[0].Core()) != "3" {
		t.Errorf("Error, unexpected Bzr version read, wanted \"3\", received: %s", string(v[0].Core()))
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
	path, err := bzrReader.Exists(Wkspc)
	if err != nil {
		t.Errorf("Existence check failed on Bzr repo: %s", err)
	}
	if path != "" {
		t.Error("Bzr Exists does detects bzr repo where one is not")
	}

	// Test NewReader when there's no local. This should simply provide a working
	// instance without error based on looking at the remote localtion.
	_, nrerr := NewReader("https://launchpad.net/dvlnbzrtest", tempDir+"/govcstestbzrrepo")
	if nrerr != nil {
		t.Error(nrerr)
	}

	// Try remote Bzr existence checks via a Getter
	url1 := "launchpad.net/dvlnbzrtest"
	bzrGetter, err := NewBzrGetter(url1, tempDir, false)
	if err != nil {
		t.Fatalf("Failed to initialize new Bzr getter, error: %s", err)
	}
	path, err = bzrGetter.Exists(Remote)
	if err != nil {
		t.Fatalf("Failed to find remote repo that should exist (URL: %s), error: %s", url1, err)
	}
	if path != "https://launchpad.net/dvlnbzrtest" {
		t.Fatalf("Exists failed to return remote path with correct scheme (URL: %s), found: %s", url1, path)
	}

	if testing.Short() {
		t.Skip("skipping remaining existence checks in short mode.")
		return
	}

	url2 := "https://launchpad.net/dvlnbzrtest"
	bzrGetter, err = NewBzrGetter(url2, tempDir, false)
	if err != nil {
		t.Fatalf("Failed to initialize new Bzr getter, error: %s", err)
	}
	path, err = bzrGetter.Exists(Remote)
	if err != nil {
		t.Fatalf("Failed to find remote repo that should exist (URL: %s), error: %s", url2, err)
	}
	if path != url2 {
		t.Fatalf("Exists failed to return matching URL path (URL: %s), found: %s", url2, path)
	}

	badurl1 := "launchpad.net/dvlnnotexistbzrtest"
	bzrGetter, err = NewBzrGetter(badurl1, tempDir, false)
	if err != nil {
		t.Fatalf("Failed to initialize 1st \"bad\" Bzr getter, init should work, error: %s", err)
	}
	path, err = bzrGetter.Exists(Remote)
	if err == nil {
		t.Fatalf("Failed to detect an error scanning for 1st bad VCS location (loc: %s), error: %s", badurl1, err)
	}
	if path != "" {
		t.Fatalf("Unexpectedly found a repo when shouldn't have (URL: %s), found path: %s", badurl1, err)
	}

	badurl2 := "https://launchpad.net/dvlnnotexistbzrtest"
	bzrGetter, err = NewBzrGetter(badurl2, tempDir, false)
	if err != nil {
		t.Fatalf("Failed to initialize 2nd \"bad\" Bzr getter, init should work, error: %s", err)
	}
	path, err = bzrGetter.Exists(Remote)
	if err == nil {
		t.Fatalf("Failed to detect an error scanning for 2nd bad VCS location (loc: %s), error: %s", badurl2, err)
	}
	if path != "" {
		t.Fatalf("Unexpectedly found a repo when shouldn't have (URL: %s), found path: %s", badurl2, err)
	}
}
