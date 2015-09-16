package vcs

import (
	"io/ioutil"
	"os"
	"testing"
)

// Canary test to ensure HgReader implements the Reader interface.
var _ Reader = &HgReader{}

// To verify hg is working we perform intergration testing
// with a known hg service.

func TestHg(t *testing.T) {

	tempDir, err := ioutil.TempDir("", "go-vcs-hg-tests")
	if err != nil {
		t.Error(err)
	}
	defer func() {
		err = os.RemoveAll(tempDir)
		if err != nil {
			t.Error(err)
		}
	}()

	hgReader, err := NewHgReader("https://bitbucket.org/mattfarina/testhgrepo", tempDir+"/testhgrepo")
	if err != nil {
		t.Errorf("Unable to instantiate new Hg VCS reader, Err: %s", err)
	}

	if hgReader.Vcs() != Hg {
		t.Error("Hg is detecting the wrong type")
	}

	// Check the basic getters.
	if hgReader.Remote() != "https://bitbucket.org/mattfarina/testhgrepo" {
		t.Error("Remote not set properly")
	}
	if hgReader.WkspcPath() != tempDir+"/testhgrepo" {
		t.Error("Local disk location not set properly")
	}

	//Logger = log.New(os.Stdout, "", log.LstdFlags)

	// Do an initial clone.
	_, err = hgReader.Get()
	if err != nil {
		t.Errorf("Unable to clone Hg repo. Err was %s", err)
	}

	// Verify Hg repo is a Hg repo
	exists, err := hgReader.Exists(Wkspc)
	if err != nil {
		t.Errorf("Existence check failed on hg repo: %s", err)
	}
	if exists == false {
		t.Error("Problem checking out repo or Hg Exists(Wkspc) not working")
	}

	// Test internal lookup mechanism used outside of Hg specific functionality.
	ltype, err := DetectVcsFromFS(tempDir + "/testhgrepo")
	if err != nil {
		t.Error("detectVcsFromFS unable to Hg repo")
	}
	if ltype != Hg {
		t.Errorf("detectVcsFromFS detected %s instead of Hg type", ltype)
	}

	// Test NewReader on existing checkout. This should simply provide a working
	// instance without error based on looking at the local directory.
	nhgReader, nrerr := NewReader("https://bitbucket.org/mattfarina/testhgrepo", tempDir+"/testhgrepo")
	if nrerr != nil {
		t.Error(nrerr)
	}
	// Verify the right oject is returned. It will check the local repo type.
	exists, err = nhgReader.Exists(Wkspc)
	if err != nil {
		t.Errorf("Existence check failed on hg repo: %s", err)
	}
	if exists == false {
		t.Error("Wrong version returned from NewReader")
	}

	// Set the version using the short hash.
	_, err = hgReader.RevSet("a5494ba2177f")
	if err != nil {
		t.Errorf("Unable to update Hg repo version. Err was %s", err)
	}

	// Use RevRead to verify we are on the right version.
	v, _, err := hgReader.RevRead(CoreRev)
	if string(v[0].Core()) != "a5494ba2177f" {
		t.Error("Error checking checked out Hg version")
	}
	if err != nil {
		t.Error(err)
	}

	// Perform an update.
	_, err = hgReader.Update()
	if err != nil {
		t.Error(err)
	}

	v, _, err = hgReader.RevRead(CoreRev)
	if string(v[0].Core()) != "d680e82228d2" {
		t.Error("Error checking checked out Hg version")
	}
	if err != nil {
		t.Error(err)
	}

}

func TestHgExists(t *testing.T) {
	// Verify hgReader.Exists fails for non-Hg directories.
	// TestHg is already checking on a valid repo
	tempDir, err := ioutil.TempDir("", "go-vcs-hg-tests")
	if err != nil {
		t.Error(err)
	}
	defer func() {
		err = os.RemoveAll(tempDir)
		if err != nil {
			t.Error(err)
		}
	}()

	hgReader, _ := NewHgReader("", tempDir)
	exists, err := hgReader.Exists(Wkspc)
	if err != nil {
		t.Errorf("Existence check failed on hg repo: %s", err)
	}
	if exists == true {
		t.Error("Hg Exists() does not identify non-Hg location")
	}

	// Test NewReader when there's no local. This should simply provide a working
	// instance without error based on looking at the remote localtion.
	_, nrerr := NewReader("https://bitbucket.org/mattfarina/testhgrepo", tempDir+"/testhgrepo")
	if nrerr != nil {
		t.Error(nrerr)
	}
}
