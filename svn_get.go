package vcs

// SvnGetter implements the Repo interface for the Svn source control.
type SvnGetter struct {
	Description
}

// NewSvnGetter creates a new instance of SvnGetter. The remote and local directories
// need to be passed in. The remote location should include the branch for SVN.
// For example, if the package is https://github.com/Masterminds/cookoo/ the remote
// should be https://github.com/Masterminds/cookoo/trunk for the trunk branch.
func NewSvnGetter(remote, wkspc string) (Getter, error) {
	ltype, err := DetectVcsFromFS(wkspc)
	// Found a VCS other than Svn. Need to report an error.
	if err == nil && ltype != Svn {
		return nil, ErrWrongVCS
	}
	g := &SvnGetter{}
	g.setDescription(remote, "", wkspc, defaultSvnSchemes, Svn)
	if err == nil { // Have a local wkspc FS repo, try to validate/upd remote
		remote, _, err = SvnCheckRemote(g, remote)
		if err != nil {
			return nil, err
		}
		g.setRemote(remote)
	}
	return g, nil
}

// Get support for svn getter
func (g *SvnGetter) Get(rev ...Rev) (string, error) {
	return SvnGet(g, rev...)
}

// RevSet support for svn getter
func (g *SvnGetter) RevSet(rev Rev) (string, error) {
	return SvnRevSet(g, rev)
}

// Exists support for svn getter
func (g *SvnGetter) Exists(l Location) (string, error) {
	return SvnExists(g, l)
}
