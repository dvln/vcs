package vcs

// BzrGetter implements the Repo interface for the Bzr source control.
type BzrGetter struct {
	Description
	mirror bool
}

// NewBzrGetter creates a new instance of BzrGetter. The remote and wkspc
// directories need to be passed in.
func NewBzrGetter(remote, wkspc string, mirror bool) (Getter, error) {
	ltype, err := DetectVcsFromFS(wkspc)
	// Found a VCS other than Bzr. Need to report an error.
	if err == nil && ltype != Bzr {
		return nil, ErrWrongVCS
	}
	g := &BzrGetter{}
	g.mirror = mirror
	g.setDescription(remote, "", wkspc, defaultBzrSchemes, Bzr)
	if err == nil { // Have a local wkspc FS repo, try to improve the remote..
		remote, _, err = BzrCheckRemote(g, remote)
		if err != nil {
			return nil, err
		}
		g.setRemote(remote)
	}
	return g, nil
}

// Get support for bzr getter
func (g *BzrGetter) Get(rev ...Rev) (string, error) {
	return BzrGet(g, rev...)
}

// RevSet support for bzr getter
func (g *BzrGetter) RevSet(rev Rev) (string, error) {
	return BzrRevSet(g, rev)
}

// Exists support for bzr getter
func (g *BzrGetter) Exists(l Location) (string, error) {
	return BzrExists(g, l)
}
