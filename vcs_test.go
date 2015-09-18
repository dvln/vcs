package vcs

import (
	"io/ioutil"
	"os"
	"testing"
)

func ExampleNewReader() {
	remote := "https://github.com/Masterminds/vcs"
	local, _ := ioutil.TempDir("", "go-vcs")
	reader, _ := NewReader(remote, local)
	// Returns: instance of GitRepo

	reader.Vcs()
	// Returns Git as this is a Git VCS reader type
}

func TestTypeSwitch(t *testing.T) {

	// To test VCS reader type switching we checkout as SVN and then
    // try to get it as a git VCS reader afterwards.
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

	getter, err := NewSvnGetter("https://github.com/Masterminds/VCSTestRepo/trunk", tempDir+"/VCSTestRepo")
	if err != nil {
		t.Error(err)
	}
	output, err := getter.Get()
	if err != nil {
		t.Errorf("Unable to checkout SVN VCS reader for reader switching tests. Err was %s, output:\n%s", err, output)
	}

	_, err = NewReader("https://github.com/Masterminds/VCSTestRepo", tempDir+"/VCSTestRepo")
	if err != ErrWrongVCS {
		t.Errorf("Not detecting VCS/repo reader switch from SVN to Git")
	}
}
