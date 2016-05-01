package vcs

// HookMgr composes the basic characteristics a vcs pkg (repo/clone) hook
// management interface..  If all these interfaces are met then hook mgmt
// support is available.  Aside: the reason a standard vcs Existence interface is
// not used is: https://groups.google.com/forum/#!topic/golang-nuts/OKgbtTW-5YQ
type HookMgr interface {
	// Describer interfaces has methods to determine info about a repo (remote/localRepo URL/path, VCS Type)
	Describer

	// Exists is the key Existence intfc func to see if the VCS is there or not,
	// see URL above for reason why the Existence interface isn't used directly
	Exists(Location) (string, Resulter, error)

	// Install is for installing a hook, params are path to hook file to install,
	// name of hook file to install, and if it should be a link or full copy,
	// returns full path to the hook installed and any hook install error
	Install(string, string, bool) (string, error)

	// Remove is for removing an installed hook (or symllink to a hook),
	// the parameter is the hook name (any error detected is returned)
	Remove(string) error
}

// NewHookMgr returns a VCS HookMgr interface to allow one to install or
// remove hooks (or links to hook) within a given VCS system.  It only
// works with local VCS's so doesn't accept remotes.  The HookMgr
// interface will be returned or an ErrCannotDetectVCS if the VCS
// type cannot be detected or ErrNoExist if the repo isn't there.
func NewHookMgr(localPath string, vcsType ...Type) (HookMgr, error) {
	vtype := NoVCS
	if vcsType != nil && len(vcsType) == 1 && vcsType[0] != NoVCS {
		vtype = vcsType[0]
	} else {
		var err error
		vtype, err = DetectVcsFromFS(localPath)
		if err != nil {
			return nil, err
		}
	}
	switch vtype {
	case Git:
		return NewGitHookMgr(localPath)
	case Svn:
		return nil, ErrNotImplemented
	case Hg:
		return nil, ErrNotImplemented
	case Bzr:
		return nil, ErrNotImplemented
	}
	// Should never fall through to here but just in case.
	return nil, ErrCannotDetectVCS
}
