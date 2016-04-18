package vcs

// HgUpdater implements the Repo interface for the Mercurial source control.
type HgUpdater struct {
	Description
	Results
	mirror bool
	rebase RebaseVal
	refs   map[string]RefOp
}

// NewHgUpdater creates a new instance of HgUpdater. The remote and localPath directories
// need to be passed in.  Params:
//	remote (string): URL of remote repo
//	remoteName (string): If there is a name for the remote repo (hg: use "", not used)
//	localPath (string): Directory for the local repo/clone/workspace to update
//	mirror (bool): if a full mirroring of all content is desired
//	rebase (RebaseVal): if rebase wanted or not, what type
//	refs (map[string]RefOp): list of refs to act on w/given operation (or nil)
// Currently ignores mirror, rebase and refs.
func NewHgUpdater(remote, remoteName, localPath string, mirror bool, rebase RebaseVal, refs map[string]RefOp) (Updater, error) {
	ltype, err := DetectVcsFromFS(localPath)
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
	u.setDescription(remote, remoteName, localPath, defaultHgSchemes, Hg)
	if err == nil { // Have a localPath FS repo, try to validate/upd remote
		remote, _, err = HgCheckRemote(u, remote)
		if err != nil {
			return nil, err
		}
		u.setRemote(remote)
	}
	return u, nil // note: above 'err' not used on purpose here..
}

// Update support for hg updater
func (u *HgUpdater) Update(rev ...Rev) (Resulter, error) {
	return HgUpdate(u, rev...)
}

// Exists support for hg updater
func (u *HgUpdater) Exists(l Location) (string, Resulter, error) {
	return HgExists(u, l)
}
