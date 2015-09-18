package vcs

// HgReader implements the Repo interface for the Mercurial source control.
type HgReader struct {
	Description
}

// NewHgReader creates a new instance of HgReader. The remote and wkspc directories
// need to be passed in.
func NewHgReader(remote, wkspc string) (*HgReader, error) {
	ltype, err := DetectVcsFromFS(wkspc)
	// Found a VCS other than Hg. Need to report an error.
	if err == nil && ltype != Hg {
		return nil, ErrWrongVCS
	}
	r := &HgReader{}
    r.setDescription(remote, "", wkspc, defaultHgSchemes, Hg)
	if err == nil { // Have a local wkspc FS repo, try to validate/upd remote
		remote, _, err = HgCheckRemote(r, remote)
		if err != nil {
			return nil, err
		}
		r.setRemote(remote)
	}
	return r, nil	// note: above 'err' not used on purpose here..
}

// Update support for hg reader
func (r *HgReader) Update(rev ...Rev) (string, error) {
	return HgUpdate(r, rev...)
}

// Get support for hg reader
func (r *HgReader) Get(rev ...Rev) (string, error) {
	return HgGet(r, rev...)
}

// RevSet support for hg reader
func (r *HgReader) RevSet(rev Rev) (string, error) {
	return HgRevSet(r, rev)
}

// RevRead support for hg reader
func (r *HgReader) RevRead(scope ReadScope, vcsRev ...Rev) ([]Revisioner, string, error) {
	return HgRevRead(r, scope, vcsRev...)
}

// Exists support for hg reader
func (r *HgReader) Exists(l Location) (string, error) {
	return HgExists(r, l)
}

