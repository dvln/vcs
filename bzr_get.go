package vcs

// BzrGetter implements the Repo interface for the Bzr source control.
type BzrGetter struct {
	Description
	mirror bool
}

// NewBzrGetter creates a new instance of BzrGetter. The remote and localPath
// directories need to be passed in.
func NewBzrGetter(remote, remoteName, localPath string, mirror bool) (Getter, error) {
	ltype, err := DetectVcsFromFS(localPath)
	// Found a VCS other than Bzr. Need to report an error.
	if err == nil && ltype != Bzr {
		return nil, ErrWrongVCS
	}
	g := &BzrGetter{}
	g.mirror = mirror
	g.setDescription(remote, "", localPath, defaultBzrSchemes, Bzr)
	if err == nil { // Have a localPath FS repo, try to improve the remote..
		remote, _, err = BzrCheckRemote(g, remote)
		if err != nil {
			return nil, err
		}
		g.setRemote(remote)
	}
	return g, nil
}

// Get support for bzr getter
func (g *BzrGetter) Get(rev ...Rev) (Resulter, error) {
	return BzrGet(g, rev...)
}

// RevSet support for bzr getter
func (g *BzrGetter) RevSet(rev Rev) (Resulter, error) {
	return BzrRevSet(g, rev)
}

// Exists support for bzr getter
func (g *BzrGetter) Exists(l Location) (string, Resulter, error) {
	return BzrExists(g, l)
}
