package vcs

// HgUpdater implements the Repo interface for the Mercurial source control.
type HgUpdater struct {
	Description
	mirror bool
	rebase RebaseVal
	refs   map[string]RefOp
}

// NewHgUpdater creates a new instance of HgUpdater. The remote and wkspc directories
// need to be passed in.  Params:
//	remote (string): URL of remote repo
//	wkspc (string): Directory for the local workspace
//	mirror (bool): if a full mirroring of all content is desired
//	rebase (RebaseVal): if rebase wanted or not, what type
//	refs (map[string]RefOp): list of refs to act on w/given operation (or nil)
// Currently ignores mirror, rebase and refs.
func NewHgUpdater(remote, wkspc string, mirror bool, rebase RebaseVal, refs map[string]RefOp) (Updater, error) {
	ltype, err := DetectVcsFromFS(wkspc)
	// Found a VCS other than Hg. Need to report an error.
	if err == nil && ltype != Hg {
		return nil, ErrWrongVCS
	}
	u := &HgUpdater{}
	u.mirror = mirror
	u.rebase = rebase
	if refs != nil { // if refs given, then set up refs to act on w/ops
		u.refs = make(map[string]RefOp)
		u.refs = refs
	}
	u.setDescription(remote, "", wkspc, defaultHgSchemes, Hg)
	if err == nil { // Have a local wkspc FS repo, try to validate/upd remote
		remote, _, err = HgCheckRemote(u, remote)
		if err != nil {
			return nil, err
		}
		u.setRemote(remote)
	}
	return u, nil // note: above 'err' not used on purpose here..
}

// Update support for hg updater
func (u *HgUpdater) Update(rev ...Rev) (string, error) {
	return HgUpdate(u, rev...)
}

// Exists support for hg updater
func (u *HgUpdater) Exists(l Location) (string, error) {
	return HgExists(u, l)
}
