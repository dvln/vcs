package vcs

// HgGetter implements the Repo interface for the Mercurial source control.
type HgGetter struct {
	Description
}

// NewHgGetter creates a new instance of HgGetter. The remote and wkspc directories
// need to be passed in.
func NewHgGetter(remote, wkspc string) (Getter, error) {
	ltype, err := DetectVcsFromFS(wkspc)
	// Found a VCS other than Hg. Need to report an error.
	if err == nil && ltype != Hg {
		return nil, ErrWrongVCS
	}
	g := &HgGetter{}
	g.setRemote(remote)
	g.setWkspcPath(wkspc)
	g.setVcs(Hg)
	if err == nil { // Have a local wkspc FS repo, try to validate/upd remote
		remote, _, err = HgCheckRemote(g, remote)
		if err != nil {
			return nil, err
		}
		g.setRemote(remote)
	}
	return g, nil
}

// Get support for hg getter
func (g *HgGetter) Get(rev ...Rev) (string, error) {
	return HgGet(g, rev...)
}

// RevSet support for hg getter
func (g *HgGetter) RevSet(rev Rev) (string, error) {
	return HgRevSet(g, rev)
}

// Exists support for hg getter
func (g *HgGetter) Exists(l Location) (bool, error) {
	return HgExists(g, l)
}

