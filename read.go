package vcs

// Reader composes the needed characteristics for full vcs pkg (repo) reader
// (no commits to VCS).  If all these interfaces are met then full pkg reading
// support is available.
type Reader interface {
	// Describer interfaces has methods to determine info about a repo (remote/localRepo URL/path, VCS Type)
	Describer

	// RevRead is the key RevReader intfc func (eg: git clone), cannot use intfc
	// (see URL above), this is like a "git log -1 --format=.." type of op
	RevRead(ReadScope, ...Rev) ([]Revisioner, Resulter, error)

	// RevSet is the key RevSetter intfc func (eg: git clone), cannot use intfc
	// (see URL above), this is like a 'git checkout <rev>' type of op
	RevSet(Rev) (Resulter, error)

	// Exists is the key Existence intfc func, cannot use intfc (see URL above),
	Exists(Location) (string, Resulter, error)
}

// NewReader returns a VCS Reader based on trying to detect the VCS sys from the
// remote and local repo locations.  The appropriate implementation will be returned
// or an ErrCannotDetectVCS if the VCS type cannot be detected.
// Note: This function can make network calls to try to determine the VCS
//       (unless grabbing the repo from a local filesystem/mount)
func NewReader(remote, localPath string, vcsType ...Type) (Reader, error) {
	vtype, remote, err := detectVCSType(remote, localPath, vcsType...)
	if err != nil {
		return nil, err
	}
	switch vtype {
	case Git:
		return NewGitReader(remote, localPath)
	case Svn:
		return NewSvnReader(remote, localPath)
	case Hg:
		return NewHgReader(remote, localPath)
	case Bzr:
		return NewBzrReader(remote, localPath)
	}

	// Should never fall through to here but just in case.
	return nil, ErrCannotDetectVCS
}
