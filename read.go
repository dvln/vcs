package vcs

// Reader composes the needed characteristics for full vcs pkg (repo) reader
// (no commits to VCS).  If all these interfaces are met then full pkg reading
// support is available.
type Reader interface {
	// Describer interfaces has methods to determine info about a repo (remote/wkspc URL/path, VCS Type)
    Describer

	// Get is the key Getter interface func (eg: git clone), cannot use a Getter:
    //   https://groups.google.com/forum/#!topic/golang-nuts/OKgbtTW-5YQ
	// ie: Describer intfc used in Getter intfc also, causes duplicate func errs
    // This is like a 'git clone ..'
	Get(...Rev) (string, error)

	// Update is the key Updater interface func (eg: git clone), cannot use an
	// Updater (see URL above for details).  This  is like a 'git fetch+merge'.
	Update(...Rev) (string, error)

	// RevRead is the key RevReader intfc func (eg: git clone), cannot use intfc
    // (see URL above), this is like a "git log -1 --format=.." type of op
	RevRead(...ReadScope) (*Revision, string, error)

	// RevSet is the key RevSetter intfc func (eg: git clone), cannot use intfc
    // (see URL above), this is like a 'git checkout <rev>' type of op
	RevSet(Rev) (string, error)

	// Exists is the key Existence intfc func, cannot use intfc (see URL above),
	Exists(Location) (bool, error)
}

// NewReader returns a VCS Reader based on trying to detect the VCS sys from the
// remote and wkspc locations.  The appropriate implementation will be returned
// or an ErrCannotDetectVCS if the VCS type cannot be detected.
// Note: This function can make network calls to try to determine the VCS
//       (unless grabbing the repo from a wkspc/local mount)
// FIXME: erik: use the optoinal vcsType argument going forward for speed optimizing
func NewReader(remote, wkspc string, vcsType ...Type) (Reader, error) {
	vtype, remote, err := detectVCSType(remote, wkspc, vcsType...)
	if err != nil {
		return nil, err
	}
	switch vtype {
	case Git:
		return NewGitReader(remote, wkspc)
	case Svn:
		return NewSvnReader(remote, wkspc)
	case Hg:
		return NewHgReader(remote, wkspc)
	case Bzr:
		return NewBzrReader(remote, wkspc)
	}

	// Should never fall through to here but just in case.
	return nil, ErrCannotDetectVCS
}

