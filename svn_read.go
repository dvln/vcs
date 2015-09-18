package vcs

// SvnReader implements the Repo interface for the Svn source control.
type SvnReader struct {
	Description
}

// NewSvnReader creates a new instance of SvnReader. The remote and local directories
// need to be passed in. The remote location should include the branch for SVN.
// For example, if the package is https://github.com/Masterminds/cookoo/ the remote
// should be https://github.com/Masterminds/cookoo/trunk for the trunk branch.
func NewSvnReader(remote, wkspc string) (*SvnReader, error) {
	ltype, err := DetectVcsFromFS(wkspc)
	// Found a VCS other than Svn. Need to report an error.
	if err == nil && ltype != Svn {
		return nil, ErrWrongVCS
	}
	r := &SvnReader{}
    r.setDescription(remote, "", wkspc, defaultSvnSchemes, Svn)
	if err == nil { // Have a local wkspc FS repo, try to validate/upd remote
		remote, _, err = SvnCheckRemote(r, remote)
		if err != nil {
			return nil, err
		}
		r.setRemote(remote)
	}
	return r, nil
}

// RevSet support for svn reader
func (r *SvnReader) RevSet(rev Rev) (string, error) {
	return SvnRevSet(r, rev)
}

// RevRead support for svn reader
func (r *SvnReader) RevRead(scope ReadScope, vcsRev ...Rev) ([]Revisioner, string, error) {
	return SvnRevRead(r, scope, vcsRev...)
}

// Exists support for svn reader
func (r *SvnReader) Exists(l Location) (string, error) {
	return SvnExists(r, l)
}

