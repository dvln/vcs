package vcs

// HgUpdater implements the Repo interface for the Mercurial source control.
type HgUpdater struct {
	Description
}

// NewHgUpdater creates a new instance of HgUpdater. The remote and wkspc directories
// need to be passed in.
func NewHgUpdater(remote, wkspc string) (Updater, error) {
	ltype, err := DetectVcsFromFS(wkspc)
	// Found a VCS other than Hg. Need to report an error.
	if err == nil && ltype != Hg {
		return nil, ErrWrongVCS
	}
	u := &HgUpdater{}
	u.setRemote(remote)
	u.setWkspcPath(wkspc)
	u.setVcs(Hg)
	if err == nil { // Have a local wkspc FS repo, try to validate/upd remote
		remote, _, err = HgCheckRemote(u, remote)
		if err != nil {
			return nil, err
		}
		u.setRemote(remote)
	}
	return u, nil
}

// Update support for hg updater
func (u *HgUpdater) Update(rev ...Rev) (string, error) {
	return HgUpdate(u, rev...)
}

// Exists support for hg updater
func (u *HgUpdater) Exists(l Location) (bool, error) {
	return HgExists(u, l)
}

