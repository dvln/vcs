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

	hgGetter, err := NewHgGetter("https://bitbucket.org/mattfarina/testhgrepo", tempDir+"/testhgrepo")
	if err != nil {
		t.Errorf("Unable to instantiate new Hg VCS reader, Err: %s", err)
	}

	if hgGetter.Vcs() != Hg {
		t.Error("Hg is detecting the wrong type")
	}

	// Check the basic getters.
	if hgGetter.Remote() != "https://bitbucket.org/mattfarina/testhgrepo" {
		t.Error("Remote not set properly")
	}
	if hgGetter.WkspcPath() != tempDir+"/testhgrepo" {
		t.Error("Local disk location not set properly")
	}

	//Logger = log.New(os.Stdout, "", log.LstdFlags)

	// Do an initial clone.
	_, err = hgGetter.Get()
	if err != nil {
		t.Errorf("Unable to clone Hg repo. Err was %s", err)
	}

	// Verify Hg repo is a Hg repo
	path, err := hgGetter.Exists(Wkspc)
	if err != nil {
		t.Errorf("Existence check failed on hg repo: %s", err)
	}
	if path == "" {
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
	hgReader, err := NewReader("https://bitbucket.org/mattfarina/testhgrepo", tempDir+"/testhgrepo")
	if err != nil {
		t.Error(err)
	}
	// Verify the right oject is returned. It will check the local repo type.
	path, err = hgReader.Exists(Wkspc)
	if err != nil {
		t.Errorf("Existence check failed on hg repo: %s", err)
	}
	if path == "" {
		t.Errorf("Existence check failed to find workspace path: %s", path)
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
	hgUpdater, err := NewUpdater("https://bitbucket.org/mattfarina/testhgrepo", tempDir+"/testhgrepo")
	if err != nil {
		t.Error(err)
	}
	_, err = hgUpdater.Update()
	if err != nil {
		t.Error(err)
	}

	v, _, err = hgReader.RevRead(CoreRev)
	if string(v[0].Core()) != "9c6ccbca73e8" {
		t.Errorf("Error checking checked out Hg version, expeced \"9c6ccbca73e8\", found: %s", string(v[0].Core()))
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
	path, err := hgReader.Exists(Wkspc)
	if err != nil {
		t.Errorf("Existence check failed on hg repo: %s", err)
	}
	if path != "" {
		t.Error("Hg Exists() does not identify non-Hg location")
	}

	// Test NewReader when there's no local. This should simply provide a working
	// instance without error based on looking at the remote localtion.
	_, nrerr := NewReader("https://bitbucket.org/mattfarina/testhgrepo", tempDir+"/testhgrepo")
	if nrerr != nil {
		t.Error(nrerr)
	}

	// Try remote Hg existence checks via a Getter
	url1 := "bitbucket.org/mattfarina/testhgrepo"
	hgGetter, err := NewHgGetter(url1, tempDir)
	if err != nil {
		t.Fatalf("Failed to initialize new Hg getter, error: %s", err)
	}
	path, err = hgGetter.Exists(Remote)
	if err != nil {
		t.Fatalf("Failed to find remote repo that should exist (URL: %s), error: %s", url1, err)
	}
	if path != "https://bitbucket.org/mattfarina/testhgrepo" {
		t.Fatalf("Exists failed to return remote path with correct scheme (URL: %s), found: %s", url1, path)
	}

    if testing.Short() {
        t.Skip("skipping remaining existence checks in short mode.")
		return
    }

	url2 := "https://bitbucket.org/mattfarina/testhgrepo"
	hgGetter, err = NewHgGetter(url2, tempDir)
	if err != nil {
		t.Fatalf("Failed to initialize new Hg getter, error: %s", err)
	}
	path, err = hgGetter.Exists(Remote)
	if err != nil {
		t.Fatalf("Failed to find remote repo that should exist (URL: %s), error: %s", url2, err)
	}
	if path != url2 {
		t.Fatalf("Exists failed to return matching URL path (URL: %s), found: %s", url2, path)
	}

	badurl1 := "bitbucket.org/mattfarina/notexisttesthgrepo"
	hgGetter, err = NewHgGetter(badurl1, tempDir)
	if err != nil {
		t.Fatalf("Failed to initialize 1st \"bad\" Hg getter, init should work, error: %s", err)
	}
	path, err = hgGetter.Exists(Remote)
	if err == nil {
		t.Fatalf("Failed to detect an error scanning for 1st bad VCS location (loc: %s), error: %s", badurl1, err)
	}
	if path != "" {
		t.Fatalf("Unexpectedly found a repo when shouldn't have (URL: %s), found path: %s", badurl1, err)
	}

	badurl2 := "https://bitbucket.org/mattfarina/notexisttesthgrepo"
	hgGetter, err = NewHgGetter(badurl2, tempDir)
	if err != nil {
		t.Fatalf("Failed to initialize 2nd \"bad\" Hg getter, init should work, error: %s", err)
	}
	path, err = hgGetter.Exists(Remote)
	if err == nil {
		t.Fatalf("Failed to detect an error scanning for 2nd bad VCS location (loc: %s), error: %s", badurl2, err)
	}
	if path != "" {
		t.Fatalf("Unexpectedly found a repo when shouldn't have (URL: %s), found path: %s", badurl2, err)
	}
}
