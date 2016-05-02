package vcs

// GitHookMgr implements the VCS Reader interface for the Git source control,
// start out by adding a base VCS description structure (implements Describer)
type GitHookMgr struct {
	Description
}

// NewHookMgr creates a new instance of GitHookMgr. The localPath dir for the
// clone should be passed in.
func NewGitHookMgr(localPath string) (*GitHookMgr, error) {
	ltype, err := DetectVcsFromFS(localPath)
	// Found a VCS other than Git. Need to report an error.
	if err == nil && ltype != Git {
		return nil, ErrWrongVCS
	} else if err != nil {
		return nil, err
	}
	r := &GitHookMgr{}
	r.setDescription("", "origin", localPath, defaultGitSchemes, Git)
	return r, nil
}

// Install is targeted at installing a git hook into a git clone.  Params:
//	hookPath (string): path to the git hook to install
//	hookName (string): git friendly name for this hook (so git will fire it)
//	link (book): true if this should be a symlink, false if copy of hook wanted
func (h *GitHookMgr) Install(hookPath, hookName string, link bool) (string, error) {
	return GitHookInstall(h, hookPath, hookName, link)
}

// Installed is targeted at checking if git hooks are installed as specified
// or not yet... if so then true, if not then false (see Install() to install)
//	hookPath (string): path to the git hook to install
//	hookName (string): git friendly name for this hook (so git will fire it)
//	link (book): true if this should be a symlink, false if copy of hook wanted
func (h *GitHookMgr) Installed(hookPath, hookName string, link bool) bool {
	return GitHookInstalled(h, hookPath, hookName, link)
}

// Remove is for removing an installed git hook from a git clone.  Params:
//	name (string): the name of the hook to remove (actual name in hooks/ dir)
func (h *GitHookMgr) Remove(name string) error {
	return GitHookRemove(h, name)
}

// Exists support for git hook manager
func (h *GitHookMgr) Exists(l Location) (string, Resulter, error) {
	return GitExists(h, l)
}
