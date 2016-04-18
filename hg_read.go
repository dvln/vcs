package vcs

// HgReader implements the Repo interface for the Mercurial source control.
type HgReader struct {
	Description
}

// NewHgReader creates a new instance of HgReader. The remote and localPath directories
// need to be passed in.
func NewHgReader(remote, localPath string) (*HgReader, error) {
	ltype, err := DetectVcsFromFS(localPath)
	// Found a VCS other than Hg. Need to report an error.
	if err == nil && ltype != Hg {
		return nil, ErrWrongVCS
	}
	r := &HgReader{}
	r.setDescription(remote, "", localPath, defaultHgSchemes, Hg)
	if err == nil { // Have a localPath FS repo, try to validate/upd remote
		remote, _, err = HgCheckRemote(r, remote)
		if err != nil {
			return nil, err
		}
		r.setRemote(remote)
	}
	return r, nil // note: above 'err' not used on purpose here..
}

// RevSet support for hg reader
func (r *HgReader) RevSet(rev Rev) (Resulter, error) {
	return HgRevSet(r, rev)
}

// RevRead support for hg reader
func (r *HgReader) RevRead(scope ReadScope, vcsRev ...Rev) ([]Revisioner, Resulter, error) {
	return HgRevRead(r, scope, vcsRev...)
}

// Exists support for hg reader
func (r *HgReader) Exists(l Location) (string, Resulter, error) {
	return HgExists(r, l)
}
