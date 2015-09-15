package vcs

// BzrReader implements the Repo interface for the Bzr source control.
type BzrReader struct {
	Description
}

// NewBzrReader creates a new instance of BzrReader. The remote and wkspc directories
// need to be passed in.
func NewBzrReader(remote, wkspc string) (*BzrReader, error) {
	ltype, err := DetectVcsFromFS(wkspc)
	// Found a VCS other than Bzr. Need to report an error.
	if err == nil && ltype != Bzr {
		return nil, ErrWrongVCS
	}
	r := &BzrReader{}
	r.setRemote(remote)
	r.setWkspcPath(wkspc)
	r.setVcs(Bzr)
	if err == nil { // Have a local wkspc FS repo, try to improve the remote..
		remote, _, err = BzrCheckRemote(r, remote)
		if err != nil {
			return nil, err
		}
		r.setRemote(remote)
	}
	return r, nil
}

// Update support for bzr reader
func (r *BzrReader) Update(rev ...Rev) (string, error) {
	return BzrUpdate(r, rev...)
}

// Get support for bzr reader
func (r *BzrReader) Get(rev ...Rev) (string, error) {
	return BzrGet(r, rev...)
}

// RevSet support for bzr reader
func (r *BzrReader) RevSet(rev Rev) (string, error) {
	return BzrRevSet(r, rev)
}

// RevRead support for bzr reader
func (r *BzrReader) RevRead(scope ...ReadScope) (*Revision, string, error) {
	return BzrRevRead(r, scope...)
}

// Exists support for bzr reader
func (r *BzrReader) Exists(l Location) (bool, error) {
	return BzrExists(r, l)
}

