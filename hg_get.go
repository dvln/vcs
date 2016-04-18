package vcs

// HgGetter implements the Repo interface for the Mercurial source control.
type HgGetter struct {
	Description
	mirror bool
}

// NewHgGetter creates a new instance of HgGetter. The remote and localPath directories
// need to be passed in.
func NewHgGetter(remote, localPath string, mirror bool) (Getter, error) {
	ltype, err := DetectVcsFromFS(localPath)
	// Found a VCS other than Hg. Need to report an error.
	if err == nil && ltype != Hg {
		return nil, ErrWrongVCS
	}
	g := &HgGetter{}
	g.mirror = mirror
	g.setDescription(remote, "", localPath, defaultHgSchemes, Hg)
	if err == nil { // Have a localPath FS repo, try to validate/upd remote
		remote, _, err = HgCheckRemote(g, remote)
		if err != nil {
			return nil, err
		}
		g.setRemote(remote)
	}
	return g, nil // note: above 'err' not used on purpose here..
}

// Get support for hg getter
func (g *HgGetter) Get(rev ...Rev) (Resulter, error) {
	return HgGet(g, rev...)
}

// RevSet support for hg getter
func (g *HgGetter) RevSet(rev Rev) (Resulter, error) {
	return HgRevSet(g, rev)
}

// Exists support for hg getter
func (g *HgGetter) Exists(l Location) (string, Resulter, error) {
	return HgExists(g, l)
}
