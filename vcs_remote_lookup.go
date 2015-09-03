// ws_remote_lookup is targeted towards trying to figure out a "remote" VCS's
// type when it is unknown (is it git, hg, svn, bzr?).  Note that "remote" means
// not in the workspace (and could be a locally visible filesystem or URL). This
// set of routines is only used for code bases that use pkg names and dvln cfg
// such that the repo/VCS path/URL and repo/VCS type can be gleaned.  This set
// of routines does the VCS *type* figuring out (it is given a repo path).  Eg:
// 1) Go source "pkgs" are "typically" VCS/repo paths (well, mostly, no scheme)
//    - can use the pkg names for repo path, here we "discover" backing VCS type
// 2) If DVLN_VCS_PATH env or 'vcs_path' cfg is set to something like:
//      % export DVLN_VCS_PATH="https://github/myorg /nfs/dir/clones"
//    and all pkg names (at least those not defined in a related codebase defn)
//    are visible relative to the given search "paths" then 'dvln' can derive
//    the likely repo/VCS path/URL's and these routines can determine repo type:
//      repo path: https://github.com/myorg/pkg1 or /nfs/dir/clones/pkg1
//      -> existence for which path to pass in here happens *elsewhere*
//      repo type: given one of those paths (whichever exists) then these
//                 routines can (try to) derive the type of those repo/VCS's
// The whole idea behind these routines is to derive the repo/VCS type given a
// repo path.  However, one should have checked existence of already derived
// repo/vcs "paths" before calling into this (to see which exists so the below
// routines can figure the type).  Example: the above path has 2 entries, maybe
// pkg1 only exists under "/nfs/dir/clones/pkg1", we need to use this repo/VCS
// path to figure out the VCS type in play (maybe it's hg).  If we used the other
// possible path (https://github.com/myorg/pkg1) we would have determined git
// here without even checking if it exists (these routines don't hit the network
// unless they have to, they assume the given repo/VCS path/uri is correct).

package vcs

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"strings"
)

type vcsInfo struct {
	host     string
	pattern  string
	vcs      Type
	addCheck func(m map[string]string) (Type, error)
	regex    *regexp.Regexp
}

var vcsList = []*vcsInfo{
	{
		host:    "github.com",
		vcs:     Git,
		pattern: `^(github\.com/[A-Za-z0-9_.\-]+/[A-Za-z0-9_.\-]+)(/[A-Za-z0-9_.\-]+)*$`,
	},
	{
		host:     "bitbucket.org",
		pattern:  `^(bitbucket\.org/(?P<name>[A-Za-z0-9_.\-]+/[A-Za-z0-9_.\-]+))(/[A-Za-z0-9_.\-]+)*$`,
		addCheck: checkBitbucket,
	},
	{
		host:    "launchpad.net",
		pattern: `^(launchpad\.net/(([A-Za-z0-9_.\-]+)(/[A-Za-z0-9_.\-]+)?|~[A-Za-z0-9_.\-]+/(\+junk|[A-Za-z0-9_.\-]+)/[A-Za-z0-9_.\-]+))(/[A-Za-z0-9_.\-]+)*$`,
		vcs:     Bzr,
	},
	{
		host:    "git.launchpad.net",
		vcs:     Git,
		pattern: `^(git\.launchpad\.net/(([A-Za-z0-9_.\-]+)|~[A-Za-z0-9_.\-]+/(\+git|[A-Za-z0-9_.\-]+)/[A-Za-z0-9_.\-]+))$`,
	},
	{
		host:    "go.googlesource.com",
		vcs:     Git,
		pattern: `^(go\.googlesource\.com/[A-Za-z0-9_.\-]+/?)$`,
	},
	// TODO: Once Google Code becomes fully deprecated this can be removed.
	{
		host:     "code.google.com",
		addCheck: checkGoogle,
		pattern:  `^(code\.google\.com/[pr]/(?P<project>[a-z0-9\-]+)(\.(?P<repo>[a-z0-9\-]+))?)(/[A-Za-z0-9_.\-]+)*$`,
	},
	// Alternative Google setup. This is the previous structure but it still works... until Google Code goes away.
	{
		addCheck: checkURL,
		pattern:  `^([a-z0-9_\-.]+)\.googlecode\.com/(?P<type>git|hg|svn)(/.*)?$`,
	},
	// If none of the previous detect the type they will fall to this looking for the type in a generic sense
	// by the extension to the path.
	{
		addCheck: checkURL,
		pattern:  `\.(?P<type>git|hg|svn|bzr)$`,
	},
}

func init() {
	// Precompile the regular expressions used to check VCS locations.
	for _, v := range vcsList {
		v.regex = regexp.MustCompile(v.pattern)
	}
}

// detectVcsFromRemote is a bit of a hack.  It tries to figure out what type
// of VCS (git,hg,bzr,svn) is the given vcsURI pointing at.  The vcsURI can be:
// 1) a local path that starts with '/'  (note: all OS's should use forward /)
// ---> if exists then "ping/scan it" determine VCS type (eg: repo on NFS mount)
// 2) a VCS URL: scan the URL for "known" naming (Go style scan)
// 3) a "redirect" URL: a Go style redirect can point to another URL & give VCS
// ---> note that this last one will hit the network if it is reached
// Return this data:
// - vcs.Type (currently Git, Hg, Bzr, Svn or noVCS if none)
// - vcsURI: will be the same as vcsURI passed in unless it's a Go-like redirect
// - error: if any issues, ErrCannotDetectVCS indicates normal but w/no match
// Note: this routine does NOT always check if the actual repo exists (although
// for some systems like bitbucket there are add-on routines thta do check)
func detectVcsFromRemote(vcsURI string) (Type, string, error) {
	//TODO: consider check for local path (ie: starts with '/'), to do this
	//      -> convert to OS specific path (forward/backward/etc)
	//      -> use this: DetectVcsFromFS(vcsPath string) (Type, error)
	//      Note: also might be good to out.WrapErr any errors here and there,
    //            if so then tests will need to be updated that check these Errs
	//      eg: support NFS path in pkg/codebase search "/nfs/somedir/<name>"
	//      eg: maybe support 'Rcs' type for codebase defn optionally (?)
	t, e := detectVcsFromURL(vcsURI)
	if e == nil {
		return t, vcsURI, nil
	}

	// Need to test for vanity or paths like golang.org/x/

	// TODO: Test for 3xx redirect codes and handle appropriately.

	// Pages like https://golang.org/x/net provide an html document with
	// meta tags containing a location to work with. The go tool uses
	// a meta tag with the name go-import which is what we use here.
	// godoc.org also has one call go-source that we do not need to use.
	// The value of go-import is in the form "prefix vcs repo". The prefix
	// should match the vcsURI and the repo is a location that can be
	// checked out. Note, to get the html document you you need to add
	// ?go-get=1 to the url.
	u, err := url.Parse(vcsURI)
	if err != nil {
		return NoVCS, "", err
	}
	if u.RawQuery == "" {
		u.RawQuery = "go-get=1"
	} else {
		u.RawQuery = u.RawQuery + "+go-get=1"
	}
	checkURL := u.String()
	resp, err := http.Get(checkURL)
	if err != nil {
		return NoVCS, "", ErrCannotDetectVCS
	}
	defer resp.Body.Close()

	t, nu, err := parseImportFromBody(u, resp.Body)
	if err != nil {
		return NoVCS, "", err
	} else if t == "" || nu == "" {
		return NoVCS, "", ErrCannotDetectVCS
	}

	return t, nu, nil
}

// detectVcsFromURL uses Go like matching to see if the given URL
// matches "known" VCS serving systems (eg: github) or, failing that,
// known VCS extensions on the repo name (eg: .git, .bzr, .hg, .svn).
// It will return the type of VCS found (or NoVCS and any error hit.
func detectVcsFromURL(vcsURL string) (Type, error) {
	u, err := url.Parse(vcsURL)
	if err != nil {
		return "", err
	}

	// If there is no host found we cannot detect the VCS from
	// the url. Note, URIs beginning with git@github using the ssh
	// syntax fail this check.
	if u.Host == "" {
		return "", ErrCannotDetectVCS
	}

	// Try to detect from known hosts, such as Github
	for _, v := range vcsList {
		if v.host != "" && v.host != u.Host {
			continue
		}

		// Make sure the pattern matches for an actual repo location. For example,
		// we should fail if the VCS listed is github.com/masterminds as that's
		// not actually a repo.
		uCheck := u.Host + u.Path
		m := v.regex.FindStringSubmatch(uCheck)
		if m == nil {
			if v.host != "" {
				return "", ErrCannotDetectVCS
			}

			continue
		}

		// If we are here the host matches. If the host has a singular
		// VCS type, such as Github, we can return the type right away.
		if v.vcs != "" {
			return v.vcs, nil
		}

		// Run additional checks to try and determine the repo
		// for the matched service.
		info := make(map[string]string)
		for i, name := range v.regex.SubexpNames() {
			if name != "" {
				info[name] = m[i]
			}
		}
		t, err := v.addCheck(info)
		if err != nil {
			return "", ErrCannotDetectVCS
		}

		return t, nil
	}

	// Unable to determine the vcs from the url.
	return "", ErrCannotDetectVCS
}

// Bitbucket provides an API for checking the VCS.
func checkBitbucket(i map[string]string) (Type, error) {
	// The part of the response we care about.
	var response struct {
		SCM Type `json:"scm"`
	}

	u := expand(i, "https://api.bitbucket.org/1.0/repositories/{name}")
	data, err := get(u)
	if err != nil {
		return "", err
	}

	if err := json.Unmarshal(data, &response); err != nil {
		return "", fmt.Errorf("Decoding error %s: %v", u, err)
	}

	return response.SCM, nil
}

// Google supports Git, Hg, and Svn. The SVN style is only
// supported through their legacy setup at <project>.googlecode.com.
// I wonder if anyone is actually using SVN support.
func checkGoogle(i map[string]string) (Type, error) {

	// To figure out which of the VCS types is used in Google Code you need
	// to parse a web page and find it. Ugh. I mean... ugh.
	var hack = regexp.MustCompile(`id="checkoutcmd">(hg|git|svn)`)

	d, err := get(expand(i, "https://code.google.com/p/{project}/source/checkout?repo={repo}"))
	if err != nil {
		return "", err
	}

	if m := hack.FindSubmatch(d); m != nil {
		if vcs := string(m[1]); vcs != "" {
			if vcs == "svn" {
				// While Google supports SVN it can only be used with the legacy
				// urls of <project>.googlecode.com. I considered creating a new
				// error for this problem but Google Code is going away and there
				// is support for the legacy structure.
				return "", ErrCannotDetectVCS
			}

			return Type(vcs), nil
		}
	}

	return "", ErrCannotDetectVCS
}

// Expect a type key on i with the exact type detected from the regex.
func checkURL(i map[string]string) (Type, error) {
	return Type(i["type"]), nil
}

func get(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("%s: %s", url, resp.Status)
	}
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("%s: %v", url, err)
	}
	return b, nil
}

func expand(match map[string]string, s string) string {
	for k, v := range match {
		s = strings.Replace(s, "{"+k+"}", v, -1)
	}
	return s
}

func parseImportFromBody(ur *url.URL, r io.ReadCloser) (tp Type, u string, err error) {
	d := xml.NewDecoder(r)
	d.CharsetReader = charsetReader
	d.Strict = false
	var t xml.Token
	for {
		t, err = d.Token()
		if err != nil {
			if err == io.EOF {
				err = nil
			}
			return
		}
		if e, ok := t.(xml.StartElement); ok && strings.EqualFold(e.Name.Local, "body") {
			return
		}
		if e, ok := t.(xml.EndElement); ok && strings.EqualFold(e.Name.Local, "head") {
			return
		}
		e, ok := t.(xml.StartElement)
		if !ok || !strings.EqualFold(e.Name.Local, "meta") {
			continue
		}
		if attrValue(e.Attr, "name") != "go-import" {
			continue
		}
		if f := strings.Fields(attrValue(e.Attr, "content")); len(f) == 3 {

			// If this the second time a go-import statement has been detected
			// return an error. There should only be one import statement per
			// html file. We don't simply return the first found in order to
			// detect pages including more than one.
			// Should this be a different error?
			if tp != "" || u != "" {
				tp = NoVCS
				u = ""
				err = ErrCannotDetectVCS
				return
			}

			// If the prefix supplied by the remote system isn't a prefix to the
			// url we're fetching return an error. This will work for exact
			// matches and prefixes. For example, golang.org/x/net as a prefix
			// will match for golang.org/x/net and golang.org/x/net/context.
			// Should this be a different error?
			vcsURL := ur.Host + ur.Path
			if !strings.HasPrefix(vcsURL, f[0]) {
				err = ErrCannotDetectVCS
				return
			}

			// We check to make sure the string in the html document is one of
			// the VCS we support. Do not want to blindly trust a string value
			// in an HTML doc.
			switch Type(f[1]) {
			case Git:
				tp = Git
			case Svn:
				tp = Svn
			case Bzr:
				tp = Bzr
			case Hg:
				tp = Hg
			}

			u = f[2]
		}
	}
}

func charsetReader(charset string, input io.Reader) (io.Reader, error) {
	switch strings.ToLower(charset) {
	case "ascii":
		return input, nil
	default:
		return nil, fmt.Errorf("can't decode XML document using charset %q", charset)
	}
}

func attrValue(attrs []xml.Attr, name string) string {
	for _, a := range attrs {
		if strings.EqualFold(a.Name.Local, name) {
			return a.Value
		}
	}
	return ""
}
