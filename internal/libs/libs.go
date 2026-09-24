// Package libs downloads the runtime jar dependencies the connector needs --
// the IBM MQ client and its transitive libraries, or the syslog encoder --
// over HTTPS into a destination directory. This is the only outbound-network
// path in the tool, so https is mandatory on the initial URL and on every
// redirect hop, the on-disk filename is derived from the URL and validated
// only after decoding, every download is size-capped on the read path and
// cross-checked against a truthful Content-Length when the server sends one,
// the write is atomic (temp file, then rename), and userinfo is stripped
// before a URL ever reaches an error or a log.
//
// Every jar is also verified against the ".sha1" sidecar Maven Central (and
// any well-behaved mirror) publishes alongside it: fetchSHA1Sidecar fetches
// it and writeAtomic hashes the bytes actually written -- while they are
// still streaming to the temp file, never by reading the file twice -- before
// the rename is ever allowed to happen. Be honest about the boundary this
// draws: the digest is served by the SAME HOST as the jar itself, so a match
// only rules out truncation, corruption and a partial transfer in flight --
// it is integrity, not authenticity, and proves nothing against a compromised
// or malicious repository. A Maven-resolved artifact requires this to
// succeed: a missing, unfetchable, malformed or mismatched sidecar fails that
// one artifact, never the whole run. An explicit --url is more lenient about
// absence only: a 404 on its sidecar is recorded as UNVERIFIED in
// Report.Unverified and the file is still written, but any other sidecar
// problem (a malformed body, a digest that does not match) fails it exactly
// like the Maven-resolved case. A response with no Content-Length
// (ContentLength == -1) that also has no verified digest is refused outright,
// since neither signal is then available to catch a connection that closes
// early.
//
// Two ways to choose what lands on disk. Given explicit --url values, exactly
// those URLs are downloaded and no Maven resolution happens at all. Otherwise
// the requested set (mq or syslog) selects a seed coordinate; resolveClosureAt
// and jarURL, in this package's own maven.go, resolve that seed -- pinned to
// Input.Version, or its latest stable release when Version is empty -- and walk its POM dependency closure through
// the same injected Doer this file downloads with, so no path in the package
// ever reaches the network except through it. A dependency keeps the version
// its POM declares -- resolved through the full parent chain when it is a
// property -- and falls back to that artifact's own latest stable release,
// noted in Report.Fallback, only when the chain never resolves it.
//
// Version-aware omission against the connector image's classpath. Once a
// Maven-resolved closure exists, and unless Input.IncludeProvided is set,
// every DEPENDENCY the image already provides at a version at least as new as
// the one just resolved is dropped before anything is fetched, and reported
// by name in Report.Omitted rather than silently disappearing -- an operator
// diffing Report.Written against a previous run must never wonder whether an
// artifact was skipped or simply forgotten. An artifact absent from the image,
// or present only at an older version, is downloaded as usual. The seed
// itself -- resolveClosureAt's result[0], but identified here by comparing
// Coord against the seed this call resolved rather than by trusting that
// position alone -- is NEVER a candidate for omission, no matter what the
// list says about it: it is the jar the operator ran the command to get, the
// same command must keep working against an older image that lacks it
// entirely, and it keeps a stale or hostile omit list from being able to
// suppress the one jar that matters (see image.go for the full rationale).
// The image's jar list comes from loadImageLibs (this package's image.go):
// Input.OmitLibFile names a captured list from a different image and
// REPLACES the built-in list completely -- there is no merging, so an empty
// file omits nothing -- and an empty Input.OmitLibFile uses the list embedded
// in the binary. Omission applies ONLY to a Maven-resolved closure -- an
// explicit --url is always downloaded verbatim, because the operator named
// it and the tool never second-guesses a named URL.
//
// Matching is by jar ARTIFACT BASE NAME plus version, never by groupId,
// because a jar filename on a classpath carries no groupId. This is a real
// limitation -- two different libraries that happen to share a base jar name
// would be confused -- but it is also why the Jackson 3 artifacts
// (tools.jackson.core) in the syslog closure still download even though the
// image ships Jackson 2 (com.fasterxml.jackson.core) under the same
// "jackson-core"/"jackson-databind" base names: Jackson 3's versions compare
// higher than the image's 2.x copies, so the version comparison alone gets
// the right answer without ever needing to know either side's groupId.
package libs

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// Doer is the injected HTTP boundary; *http.Client satisfies it.
type Doer interface {
	Do(req *http.Request) (*http.Response, error)
}

// Coord is a Maven groupId:artifactId pair, without a version.
type Coord struct {
	Group    string
	Artifact string
}

// CentralBase is the Maven Central layout root every metadata, POM and jar
// URL is built from.
const CentralBase = "https://repo1.maven.org/maven2"

// Set names: exported so the CLI model can be gated against them.
const (
	SetMQ     = "mq"
	SetSyslog = "syslog"
)

// defaultTimeout bounds one artifact's entire fetch end to end -- every
// redirect hop downloadOne follows plus reading the final response body --
// via a single shared context deadline, so a chain of slow redirects and a
// body that never ends cannot together hang the tool longer than this.
const defaultTimeout = 30 * time.Second

// maxRedirectHops caps how many redirects downloadOne will follow for a single
// artifact, so a redirect loop cannot spin forever.
const maxRedirectHops = 5

// maxDownloadArtifacts caps how many artifacts a single Download call will
// attempt, whether from an explicit --url list or a resolved closure, so a
// pathological input cannot turn one command into an unbounded fan-out.
const maxDownloadArtifacts = 128

// maxArtifactBytes is the per-artifact byte cap enforced on the read path,
// never on a trusted Content-Length. It is a var, not a const, only so a test
// can shrink it to force the cap without downloading real megabytes; production
// callers never change it. The largest jar in the current seed closures is
// bcprov-jdk18on at roughly 8.9 MB, so 32 MiB leaves headroom for growth while
// still bounding the worst case.
var maxArtifactBytes int64 = 32 * 1024 * 1024

// sha1SidecarSuffix is appended to a jar's own URL to name the checksum
// sidecar Maven Central (and any well-behaved mirror) publishes beside it.
const sha1SidecarSuffix = ".sha1"

// maxSHA1SidecarBytes bounds one fetched sidecar body. A real sidecar is a
// 40-character hex digest, sometimes followed by whitespace and a filename --
// well under 100 bytes; this only guards against a pathological response.
const maxSHA1SidecarBytes = 4096

// SetNames returns the valid Input.Set values, in the order CLI completion and
// help should present them.
func SetNames() []string {
	return []string{SetMQ, SetSyslog}
}

// resolveSeed maps a set to the Maven coordinate its closure is seeded from.
// An unknown set is a systemic error naming the offending value and listing the
// valid ones.
//
// The mq seed is the Jakarta build of the IBM client, and there is no choice to
// make: the connector image is a Jakarta stack (it ships jakarta.jms-api,
// Spring 6 and mq-jms-spring-boot-starter 3.x -- see the imagelibs list), and a
// client implementing javax.jms cannot satisfy a jakarta.jms binder. The javax
// build (com.ibm.mq:com.ibm.mq.allclient) was selectable until it became clear
// it could only ever produce a classpath that fails at run time -- and that
// having both on one classpath is worse still, since the two builds carry the
// same com.ibm.mq.* and com.ibm.msg.client.* classes and load order decides
// which wins.
func resolveSeed(set string) (Coord, error) {
	switch set {
	case SetMQ:
		return Coord{Group: "com.ibm.mq", Artifact: "com.ibm.mq.jakarta.client"}, nil
	case SetSyslog:
		return Coord{Group: "net.logstash.logback", Artifact: "logstash-logback-encoder"}, nil
	default:
		return Coord{}, fmt.Errorf("unknown set %q: must be one of %s", set, strings.Join(SetNames(), ", "))
	}
}

// Input is one Download request.
type Input struct {
	Dir             string   // destination directory, created if needed
	Set             string   // SetMQ or SetSyslog; ignored when URLs is non-empty
	Version         string   // pin the seed release; empty means latest stable
	URLs            []string // explicit overrides; when non-empty no Maven resolution and no omission happens
	Force           bool     // overwrite existing files
	OmitLibFile     string   // path to an omit list that REPLACES the embedded default completely; empty means the embedded default; an empty file omits nothing
	IncludeProvided bool     // download the whole closure even where the image already provides it
	HTTP            Doer     // nil means a default client with a timeout

	// DeployedImage is the connector image reference the operator is actually
	// deploying, as declared in env.yaml (e.g.
	// "solace/solace-pubsub-connector-ibmmq:2.14.1"). Empty when unknown --
	// download runs perfectly well with no config at all.
	//
	// Purely advisory: it never changes what is downloaded, only whether
	// Report.OmitListImageMismatch is set. Omission compares against the jar
	// list, and this says whether that list describes the right image.
	DeployedImage string
}

// Failure is one artifact that could not be downloaded.
type Failure struct {
	Name string // artifact filename, or the URL when no name is known
	Err  error
}

// Report is the outcome of a Download call.
type Report struct {
	Written  []string // paths written, sorted
	Skipped  []string // paths that already existed, sorted
	Omitted  []string // human-readable: "jakarta.jms-api-3.0.0.jar: the image has 3.1.0"
	Fallback []string // human-readable notes: an artifact whose version could not be
	// resolved from the POM chain and fell back to latest stable
	Unverified []string // human-readable: a jar written without sha1 verification (an
	// explicit --url whose sidecar 404s); never populated for a
	// Maven-resolved artifact, which always requires verification
	Failed []Failure // per-artifact failures, in resolution order

	// OmitListProvenance names which omit list was in effect for this run's
	// omission step: the operator's --omit-lib-file path verbatim, or a name
	// identifying the embedded default. Empty when omission never ran (an
	// --url download, or IncludeProvided). Surfacing this lets an operator
	// deploying against an image other than the one the built-in list came
	// from see the mismatch in the report instead of discovering it at
	// runtime.
	OmitListProvenance string
	// OmitListWarnings names the artifacts THIS closure had to download
	// because the omit list's own entry for them carried a version
	// loadImageLibs could not trust (see validateImageVersion) and so treated
	// as not-provided rather than compared.
	//
	// Scoped to the closure deliberately. It used to carry one warning per
	// rejected line in the list, which meant a `download jar mq` reported
	// netty and hibernate-validator -- artifacts no mq closure references --
	// on every single run. A rejection the download never consulted changed
	// nothing, so it is not worth an operator's attention.
	OmitListWarnings []string
	// OmitListImageMismatch is set when Input.DeployedImage names a connector
	// image the embedded jar list is not known to describe, so every omission
	// above may be a claim about a different classpath than the one being
	// deployed to. The list covers a range of releases, not the single one it
	// was captured from, so a newer image than that capture is silent.
	//
	// Its own field rather than an OmitListWarnings entry: that list means
	// "entries this closure could not compare", which is a per-artifact
	// finding. This is one statement about the whole run, and conflating them
	// would blur a distinction the report just gained.
	OmitListImageMismatch string
}

// downloadItem is one artifact queued for the download loop: a URL to fetch
// and, once known, the filename it will be written under. name is left empty
// for a seed-resolved artifact, whose filename is only derived (and validated)
// once its jar URL is in hand.
type downloadItem struct {
	url      string
	name     string
	fallback bool
}

// Download resolves Input to a set of artifacts -- either exactly in.URLs, or
// the POM dependency closure of the seed selected by in.Set -- and
// downloads each into in.Dir.
//
// error != nil is SYSTEMIC: a malformed or non-https --url, an unknown Set
// or Version value, a destination dir that cannot be created, the seed
// artifact's metadata being unreachable, or a named --omit-lib-file (or the
// embedded default) failing to load. Nothing is written in that case. A
// per-artifact problem -- one jar 404s, a redirect steps off https, an
// artifact exceeds the byte cap, arrives short of its own Content-Length, or
// fails sha1 verification -- never aborts the run; it is recorded in
// Report.Failed while every other artifact still downloads.
func Download(in Input) (Report, error) {
	var rep Report

	doer := in.HTTP
	if doer == nil {
		doer = defaultClient()
	}

	var items []downloadItem
	usingURLs := len(in.URLs) > 0

	var seed Coord
	if usingURLs {
		if len(in.URLs) > maxDownloadArtifacts {
			return Report{}, fmt.Errorf("%d urls exceeds the maximum of %d in a single download", len(in.URLs), maxDownloadArtifacts)
		}
		for _, raw := range in.URLs {
			parsed, err := url.Parse(raw)
			if err != nil {
				return Report{}, fmt.Errorf("parsing url %q: %w", raw, err)
			}
			if !parsed.IsAbs() {
				return Report{}, fmt.Errorf("url %q is not absolute", sanitizeURL(parsed))
			}
			if err := requireHTTPS(parsed); err != nil {
				return Report{}, fmt.Errorf("url %q: %w", sanitizeURL(parsed), err)
			}
			name, err := filenameFromEscapedPath(parsed.EscapedPath())
			if err != nil {
				return Report{}, fmt.Errorf("url %q: %w", sanitizeURL(parsed), err)
			}
			items = append(items, downloadItem{url: raw, name: name})
		}
	} else {
		s, err := resolveSeed(in.Set)
		if err != nil {
			return Report{}, err
		}
		seed = s
		// Input.Version is operator input arriving straight from the command
		// line, so it is validated against the same conservative charset the
		// resolver enforces on every XML-sourced coordinate part, before it
		// ever reaches a URL.
		if in.Version != "" {
			if err := validateCoordPart(in.Version); err != nil {
				return Report{}, fmt.Errorf("version %q: %w", in.Version, err)
			}
		}
	}

	if err := os.MkdirAll(in.Dir, 0o755); err != nil {
		return Report{}, fmt.Errorf("creating %q: %w", in.Dir, err)
	}

	if !usingURLs {
		// The Set-based path cannot know its target filenames -- and so
		// cannot check the destination for existing files -- until the
		// closure is resolved: a filename is derived from the resolved
		// artifact's version, not from the seed coordinate alone. Unlike the
		// --url path, which skips the network entirely when every named file
		// already exists, this path always contacts Maven Central first, on
		// every run, even when nothing will end up written. That is a
		// deliberate trade-off, not an oversight; see
		// TestDownloadSetPathAlwaysResolvesEvenWhenFilesExist.
		artifacts, err := resolveClosureAt(doer, seed, in.Version)
		if err != nil {
			return Report{}, fmt.Errorf("resolving %s:%s: %w", seed.Group, seed.Artifact, err)
		}
		if len(artifacts) > maxDownloadArtifacts {
			return Report{}, fmt.Errorf("%d artifacts exceeds the maximum of %d in a single download", len(artifacts), maxDownloadArtifacts)
		}

		if !in.IncludeProvided {
			loaded, err := loadImageLibs(in.OmitLibFile)
			if err != nil {
				if in.OmitLibFile != "" {
					// The operator named this file, so a bad one is a
					// systemic error: nothing is downloaded until it is
					// fixed, exactly like a bad --url or an unknown Set.
					// err already names the offending path.
					return Report{}, fmt.Errorf("loading omit list: %w", err)
				}
				// An empty OmitLibFile means the embedded default, and that
				// failing to load or parse is a bug baked into the binary --
				// there is no operator-supplied file to blame -- but Download
				// still cannot silently proceed, so it is surfaced the same
				// way any other systemic failure is.
				return Report{}, fmt.Errorf("loading embedded default omit list: %w", err)
			}
			rep.OmitListProvenance = loaded.Provenance
			// Only when the embedded default is in play: a named --omit-lib-file
			// is the operator's own statement about their image, and
			// second-guessing it would contradict the same rule that makes an
			// explicit --url immune to omission.
			if in.OmitLibFile == "" {
				rep.OmitListImageMismatch = imageMismatchNote(in.DeployedImage)
			}

			kept := artifacts[:0]
			for _, a := range artifacts {
				// The seed is never a candidate for omission -- identified by
				// Coord equality against the seed this call resolved, not by
				// trusting that resolveClosureAt always puts it at index 0.
				// See the package doc comment and image.go for why.
				if a.Coord == seed {
					kept = append(kept, a)
					continue
				}
				if provided, imageVersion := imageSatisfies(loaded.Libs, a); provided {
					rep.Omitted = append(rep.Omitted, fmt.Sprintf("%s-%s.jar: the image has %s", a.Artifact, a.Version, imageVersion))
					continue
				}
				// Not provided -- but if the image DID list this artifact and
				// its version was rejected as unparseable, that rejection is
				// why it is being downloaded, so say so here. Reporting it at
				// load time instead named every rejected line on every run,
				// including artifacts no closure ever asks about.
				if reason, wasRejected := loaded.Rejected[a.Artifact]; wasRejected {
					rep.OmitListWarnings = append(rep.OmitListWarnings, reason)
				}
				kept = append(kept, a)
			}
			artifacts = kept
		}

		for _, a := range artifacts {
			items = append(items, downloadItem{url: jarURL(a), fallback: a.Fallback})
		}
	}

	for _, it := range items {
		name := it.name
		if name == "" {
			parsed, err := url.Parse(it.url)
			if err != nil {
				rep.Failed = append(rep.Failed, Failure{Name: it.url, Err: fmt.Errorf("parsing resolved url: %w", err)})
				continue
			}
			n, err := filenameFromEscapedPath(parsed.EscapedPath())
			if err != nil {
				rep.Failed = append(rep.Failed, Failure{Name: sanitizeURL(parsed), Err: err})
				continue
			}
			name = n
		}

		dst := filepath.Join(in.Dir, name)
		if !in.Force {
			if _, statErr := os.Stat(dst); statErr == nil {
				rep.Skipped = append(rep.Skipped, dst)
				continue
			}
		}

		// requireVerification is uniform for the whole call: usingURLs is
		// decided once, before this loop, and never mixes --url items with
		// Maven-resolved ones.
		unverified, err := downloadWithVerification(doer, it.url, dst, !usingURLs)
		if err != nil {
			rep.Failed = append(rep.Failed, Failure{Name: name, Err: err})
			continue
		}

		rep.Written = append(rep.Written, dst)
		if it.fallback {
			rep.Fallback = append(rep.Fallback, fmt.Sprintf("%s: version could not be resolved from the POM chain, used latest stable", name))
		}
		if unverified {
			rep.Unverified = append(rep.Unverified, fmt.Sprintf("%s: no sha1 sidecar found for this url, written unverified", name))
		}
	}

	sort.Strings(rep.Written)
	sort.Strings(rep.Skipped)
	return rep, nil
}

// defaultClient is the Doer used when Input.HTTP is nil. CheckRedirect returns
// http.ErrUseLastResponse so Do never follows a redirect on its own: downloadOne
// inspects and re-validates every hop itself, the same way regardless of which
// Doer is in play.
func defaultClient() *http.Client {
	return &http.Client{
		Timeout: defaultTimeout,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
}

// requireHTTPS rejects anything but an https scheme.
func requireHTTPS(u *url.URL) error {
	if u.Scheme != "https" {
		return fmt.Errorf("scheme %q is not allowed: only https is permitted", u.Scheme)
	}
	return nil
}

// sanitizeURL renders u with any userinfo stripped, so a credential embedded
// in a URL never reaches an error message or a log line.
func sanitizeURL(u *url.URL) string {
	c := *u
	c.User = nil
	return c.String()
}

// filenameFromEscapedPath derives the on-disk filename from a URL's last path
// element and validates it. escaped is the URL's still percent-encoded path
// (url.URL.EscapedPath): splitting on "/" before decoding means a literal "/"
// can only be a real path separator, never one smuggled in as %2F inside a
// single segment. Only the last segment is decoded and validated -- decoding
// after the split, never before, so a value like %2e%2e%2f cannot pass as a
// clean-looking name and only reveal a traversal once written.
func filenameFromEscapedPath(escaped string) (string, error) {
	segments := strings.Split(escaped, "/")
	last := segments[len(segments)-1]
	decoded, err := url.PathUnescape(last)
	if err != nil {
		return "", fmt.Errorf("decoding filename: %w", err)
	}
	if err := validateFilenameShape(decoded); err != nil {
		return "", err
	}
	return decoded, nil
}

// validateFilenameShape rejects anything that is not a safe, plain jar
// filename: empty, ".", "..", a path separator (either / or \), an absolute
// path, a control character, or a name that does not end in ".jar".
func validateFilenameShape(name string) error {
	if name == "" {
		return fmt.Errorf("filename is empty")
	}
	if name == "." || name == ".." {
		return fmt.Errorf("filename %q is not allowed", name)
	}
	if strings.ContainsAny(name, "/\\") {
		return fmt.Errorf("filename %q contains a path separator", name)
	}
	if filepath.IsAbs(name) {
		return fmt.Errorf("filename %q is an absolute path", name)
	}
	for _, r := range name {
		if r < 0x20 || r == 0x7f {
			return fmt.Errorf("filename %q contains a control character", name)
		}
	}
	if !strings.HasSuffix(strings.ToLower(name), ".jar") {
		return fmt.Errorf("filename %q does not end in .jar", name)
	}
	return nil
}

// isRedirectStatus reports whether code is one of the HTTP redirect statuses
// downloadOne follows itself.
func isRedirectStatus(code int) bool {
	switch code {
	case http.StatusMovedPermanently, http.StatusFound, http.StatusSeeOther,
		http.StatusTemporaryRedirect, http.StatusPermanentRedirect:
		return true
	}
	return false
}

// downloadWithVerification fetches rawURL's sha1 sidecar and then, once the
// expected digest is known (or known to be absent), downloads the jar itself
// via downloadOne, which hashes the bytes actually written and compares them
// before the rename. requireVerification is true for a Maven-resolved
// artifact -- where any sidecar problem, including a plain 404, is always a
// per-artifact failure -- and false for an explicit --url, where a 404 alone
// is UNVERIFIED (unverified=true) rather than a failure; any other sidecar
// problem (a network error, a non-200/404 status, a malformed body) fails
// the artifact either way, because that is not "no sidecar here", it is a
// broken one.
func downloadWithVerification(doer Doer, rawURL, dst string, requireVerification bool) (unverified bool, err error) {
	sidecar, sErr := fetchSHA1Sidecar(doer, rawURL)
	if sErr != nil {
		return false, fmt.Errorf("sha1 sidecar: %w", sErr)
	}
	if sidecar.notFound {
		if requireVerification {
			return false, fmt.Errorf("sha1 sidecar not found (a Maven-resolved artifact always requires one)")
		}
		if err := downloadOne(doer, rawURL, dst, ""); err != nil {
			return false, err
		}
		return true, nil
	}
	if err := downloadOne(doer, rawURL, dst, sidecar.digest); err != nil {
		return false, err
	}
	return false, nil
}

// sha1Sidecar is the outcome of fetching one jar's ".sha1" sidecar. notFound
// is true only when the sidecar request itself returned 404 -- the one case
// downloadWithVerification lets an explicit --url treat as UNVERIFIED rather
// than a failure. Any other problem is returned as an error instead.
type sha1Sidecar struct {
	digest   string
	notFound bool
}

// fetchSHA1Sidecar fetches and strictly parses the ".sha1" sidecar Maven
// Central (and any well-behaved mirror) publishes beside a jar, at rawURL
// with sha1SidecarSuffix appended. It follows redirects itself, re-validating
// https on every hop exactly like downloadOne, under its own defaultTimeout
// deadline and its own maxSHA1SidecarBytes cap -- entirely independent of the
// jar fetch that follows it.
//
// Parsing is strict: the body is trimmed, split on whitespace, the first
// field is lowercased and must be exactly 40 hex characters. Maven Central
// sometimes follows the digest with whitespace and a filename (the
// sha1sum(1) two-column format); only the first field is ever used.
func fetchSHA1Sidecar(doer Doer, rawURL string) (sha1Sidecar, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	u := rawURL + sha1SidecarSuffix
	for hop := 0; ; hop++ {
		if hop > maxRedirectHops {
			return sha1Sidecar{}, fmt.Errorf("too many redirects (max %d)", maxRedirectHops)
		}

		parsed, err := url.Parse(u)
		if err != nil {
			return sha1Sidecar{}, fmt.Errorf("parsing sha1 sidecar url: %w", err)
		}
		if err := requireHTTPS(parsed); err != nil {
			return sha1Sidecar{}, err
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
		if err != nil {
			return sha1Sidecar{}, fmt.Errorf("building sha1 sidecar request: %w", err)
		}
		resp, err := doer.Do(req)
		if err != nil {
			return sha1Sidecar{}, fmt.Errorf("requesting sha1 sidecar %s: %w", sanitizeURL(parsed), err)
		}

		if isRedirectStatus(resp.StatusCode) {
			loc := resp.Header.Get("Location")
			resp.Body.Close()
			if loc == "" {
				return sha1Sidecar{}, fmt.Errorf("sha1 sidecar redirect from %s carries no location", sanitizeURL(parsed))
			}
			ref, err := url.Parse(loc)
			if err != nil {
				return sha1Sidecar{}, fmt.Errorf("parsing sha1 sidecar redirect location: %w", err)
			}
			next := parsed.ResolveReference(ref)
			u = next.String()
			continue
		}

		if resp.StatusCode == http.StatusNotFound {
			resp.Body.Close()
			return sha1Sidecar{notFound: true}, nil
		}
		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			return sha1Sidecar{}, fmt.Errorf("unexpected status %s from sha1 sidecar %s", resp.Status, sanitizeURL(parsed))
		}

		body, err := io.ReadAll(io.LimitReader(resp.Body, maxSHA1SidecarBytes+1))
		resp.Body.Close()
		if err != nil {
			return sha1Sidecar{}, fmt.Errorf("reading sha1 sidecar %s: %w", sanitizeURL(parsed), err)
		}
		if int64(len(body)) > maxSHA1SidecarBytes {
			return sha1Sidecar{}, fmt.Errorf("sha1 sidecar %s exceeds the %d byte cap", sanitizeURL(parsed), maxSHA1SidecarBytes)
		}
		digest, perr := parseSHA1Sidecar(body)
		if perr != nil {
			return sha1Sidecar{}, fmt.Errorf("parsing sha1 sidecar %s: %w", sanitizeURL(parsed), perr)
		}
		return sha1Sidecar{digest: digest}, nil
	}
}

// parseSHA1Sidecar strictly parses a .sha1 sidecar body: trim, take the
// first whitespace-separated field, lowercase, require exactly 40 hex
// characters. Anything else -- empty, short, non-hex, an HTML error page --
// is rejected rather than trusted.
func parseSHA1Sidecar(body []byte) (string, error) {
	fields := strings.Fields(string(body))
	if len(fields) == 0 {
		return "", fmt.Errorf("empty sidecar body")
	}
	digest := strings.ToLower(fields[0])
	if len(digest) != 40 {
		return "", fmt.Errorf("digest %q is not 40 hex characters", digest)
	}
	for _, r := range digest {
		if !((r >= '0' && r <= '9') || (r >= 'a' && r <= 'f')) {
			return "", fmt.Errorf("digest %q contains non-hex characters", digest)
		}
	}
	return digest, nil
}

// downloadOne fetches rawURL, following redirects itself (re-validating https
// on every hop, up to maxRedirectHops), and writes the body to dst, verifying
// it against expectedSHA1 (see writeAtomic) when non-empty. All hops for this
// one artifact share a single defaultTimeout deadline, so a chain of
// redirects cannot each buy themselves a fresh window.
func downloadOne(doer Doer, rawURL string, dst string, expectedSHA1 string) error {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	u := rawURL
	for hop := 0; ; hop++ {
		if hop > maxRedirectHops {
			return fmt.Errorf("too many redirects (max %d)", maxRedirectHops)
		}

		parsed, err := url.Parse(u)
		if err != nil {
			return fmt.Errorf("parsing url: %w", err)
		}
		if err := requireHTTPS(parsed); err != nil {
			return err
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
		if err != nil {
			return fmt.Errorf("building request: %w", err)
		}
		resp, err := doer.Do(req)
		if err != nil {
			return fmt.Errorf("requesting %s: %w", sanitizeURL(parsed), err)
		}

		if isRedirectStatus(resp.StatusCode) {
			loc := resp.Header.Get("Location")
			resp.Body.Close()
			if loc == "" {
				return fmt.Errorf("redirect from %s carries no location", sanitizeURL(parsed))
			}
			ref, err := url.Parse(loc)
			if err != nil {
				return fmt.Errorf("parsing redirect location: %w", err)
			}
			next := parsed.ResolveReference(ref)
			u = next.String()
			continue
		}

		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			return fmt.Errorf("unexpected status %s from %s", resp.Status, sanitizeURL(parsed))
		}

		return writeAtomic(resp, dst, expectedSHA1)
	}
}

// writeAtomic streams resp's body into a temp file in dst's own directory,
// hashing it with sha1 as it copies -- alongside the io.Copy, never by
// re-reading the file -- capped at maxArtifactBytes, then renames it over dst
// only once every check below has passed. The temp file is removed on every
// error path -- a cap trip, an empty body, a byte count that disagrees with a
// truthful Content-Length, a sha1 mismatch, or an unverifiable download --
// so an interrupted, truncated, corrupted or unverified download never
// leaves a partial or unchecked jar on a classpath.
//
// expectedSHA1 is the lowercase 40-character hex digest fetchSHA1Sidecar
// parsed from the artifact's sidecar, or "" when downloadWithVerification
// determined none is required to be checked (the UNVERIFIED path for an
// explicit --url whose sidecar returned 404). When expectedSHA1 is set, the bytes
// actually written must hash to it or the write fails -- this closes the
// hole a bare Content-Length check cannot: a mirror that answers 200 OK and
// then closes the connection early would otherwise pass unnoticed whenever
// Content-Length itself was unknown. When expectedSHA1 is "" AND
// resp.ContentLength is also unknown (-1), there is no signal left at all
// that the bytes written are complete, so that combination is itself a
// per-artifact failure rather than a silent accept -- Maven Central always
// sends Content-Length, so this only bites an odd mirror on an unverified
// --url.
//
// Three checks a prior pass already got right and this pass must not
// regress: a 0-byte body is rejected unconditionally regardless of
// Content-Length or expectedSHA1; n == resp.ContentLength is enforced
// whenever ContentLength >= 0; and the deferred cleanup removes the temp file
// on every error path, because every error return happens before "done =
// true".
//
// Whatever this function returns, remember what it does NOT prove: the sha1
// sidecar is served by the same host as the jar itself, so a match proves
// only that the bytes on disk match what that host currently serves --
// integrity against truncation, corruption or a partial transfer, never
// authenticity against a host that was compromised or malicious to begin
// with.
func writeAtomic(resp *http.Response, dst string, expectedSHA1 string) error {
	body := resp.Body
	defer body.Close()

	dir := filepath.Dir(dst)
	tmp, err := os.CreateTemp(dir, ".libs-download-*.tmp")
	if err != nil {
		return fmt.Errorf("creating temp file: %w", err)
	}
	tmpPath := tmp.Name()
	done := false
	defer func() {
		tmp.Close()
		if !done {
			os.Remove(tmpPath)
		}
	}()

	hasher := sha1.New()
	n, err := io.Copy(tmp, io.TeeReader(io.LimitReader(body, maxArtifactBytes+1), hasher))
	if err != nil {
		return fmt.Errorf("writing %s: %w", filepath.Base(dst), err)
	}
	if n > maxArtifactBytes {
		return fmt.Errorf("%s exceeds the %d byte cap", filepath.Base(dst), maxArtifactBytes)
	}
	if n == 0 {
		return fmt.Errorf("%s: response body was empty", filepath.Base(dst))
	}
	if resp.ContentLength >= 0 && n != resp.ContentLength {
		return fmt.Errorf("%s: wrote %d bytes but Content-Length was %d", filepath.Base(dst), n, resp.ContentLength)
	}
	switch {
	case expectedSHA1 != "":
		if got := hex.EncodeToString(hasher.Sum(nil)); got != expectedSHA1 {
			return fmt.Errorf("%s: sha1 mismatch: computed %s, expected %s", filepath.Base(dst), got, expectedSHA1)
		}
	case resp.ContentLength == -1:
		return fmt.Errorf("%s: no Content-Length and no sha1 digest verified: refusing an unverifiable download", filepath.Base(dst))
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("closing temp file: %w", err)
	}
	if err := os.Rename(tmpPath, dst); err != nil {
		return fmt.Errorf("renaming into place: %w", err)
	}
	done = true
	return nil
}
