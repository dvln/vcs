package vcs

// BzrReader implements the Repo interface for the Bzr source control.
type BzrReader struct {
	Description
}

// NewBzrReader creates a new instance of BzrReader. The remote and localPath directories
// need to be passed in.
func NewBzrReader(remote, localPath string) (*BzrReader, error) {
	ltype, err := DetectVcsFromFS(localPath)
	// Found a VCS other than Bzr. Need to report an error.
	if err == nil && ltype != Bzr {
		return nil, ErrWrongVCS
	}
	r := &BzrReader{}
	r.setDescription(remote, "", localPath, defaultBzrSchemes, Bzr)
	if err == nil { // Have a localPath FS repo, try to improve the remote..
		remote, _, err = BzrCheckRemote(r, remote)
		if err != nil {
			return nil, err
		}
		r.setRemote(remote)
	}
	return r, nil
}

// RevSet support for bzr reader
func (r *BzrReader) RevSet(rev Rev) (Resulter, error) {
	return BzrRevSet(r, rev)
}

// RevRead support for bzr reader
func (r *BzrReader) RevRead(scope ReadScope, vcsRev ...Rev) ([]Revisioner, Resulter, error) {
	return BzrRevRead(r, scope, vcsRev...)
}

// Exists support for bzr reader
func (r *BzrReader) Exists(l Location) (string, Resulter, error) {
	return BzrExists(r, l)
}
