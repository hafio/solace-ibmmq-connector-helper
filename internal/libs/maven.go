package libs

import (
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
)

// This file resolves a Maven Central dependency closure for one seed
// coordinate: fetching maven-metadata.xml and .pom documents through the
// injected Doer, comparing versions, walking a POM's <parent> chain to merge
// properties and dependencyManagement, and applying the compile/runtime,
// non-optional, jar-only, first-seen-wins closure rules. resolveClosure,
// resolveClosureAt and jarURL are the only names the rest of the package
// (Download, in libs.go) needs; everything below is a private helper.
//
// encoding/xml is a first for this codebase (nothing else here parses XML).
// The <properties> element has caller-defined child element names rather
// than a fixed schema, so it is captured with a plain `,any` catch-all struct
// and turned into a map by the child element's local name.

const (
	// maxXMLBytes bounds one maven-metadata.xml or .pom response. Real
	// documents are a few KB; this only guards against a malicious or
	// corrupted response with an unbounded body, never a legitimate one.
	maxXMLBytes = 2 * 1024 * 1024

	// maxArtifacts caps the whole resolved closure. The real sets are 5-6
	// jars (mq) or ~4 jars (syslog); this leaves headroom for a legitimately
	// larger POM tree while still bounding a hostile or cyclic one.
	maxArtifacts = 64

	// maxResolveDepth caps how many dependency-graph levels are expanded.
	// The deepest verified chain is bcpkix -> bcutil -> bcprov (depth 3) or
	// jackson-databind -> jackson-core/annotations (depth 2); 8 is ample
	// margin without allowing an unbounded walk.
	maxResolveDepth = 8

	// maxParentDepth caps how many <parent> links are followed. The deepest
	// verified chain (jackson-databind -> jackson-base) is 2 levels; 10
	// leaves margin without allowing a parent cycle to loop forever.
	maxParentDepth = 10
)

// coordCharsetRe is the conservative charset every group/artifact/version
// string parsed out of XML must satisfy before it is spliced into a URL.
// Maven Central coordinates only ever use these characters in practice
// (com.ibm.mq, com.ibm.mq.jakarta.client, 10.0.0.0, 3.0-rc5); rejecting
// anything else keeps a hostile POM from producing an off-host URL.
var coordCharsetRe = regexp.MustCompile(`^[A-Za-z0-9._-]+$`)

// preReleaseSegmentRe matches one version segment (split on '.', '-', '_')
// that names a pre-release qualifier: SNAPSHOT, alpha/beta/rc/cr/pr/ea/
// preview with an optional trailing number, or M<n> (a milestone). Matching
// is case-insensitive since Maven qualifiers are not case-normalized.
var preReleaseSegmentRe = regexp.MustCompile(`(?i)^(snapshot|alpha[0-9]*|beta[0-9]*|rc[0-9]*|cr[0-9]*|pr[0-9]*|ea[0-9]*|preview[0-9]*|m[0-9]+)$`)

// releaseSegmentRe matches a segment naming a qualifier that means "this IS
// the release": Final (Netty, Hibernate, JBoss), RELEASE (older Spring), GA.
// Maven's own ComparableVersion treats all three as aliases of the EMPTY
// qualifier, so 4.1.135.Final and 4.1.135 are the same version -- which is why
// compareVersions skips such a segment rather than ranking the longer side
// higher, and why compareVersionSegment places it BETWEEN a number and a
// pre-release (see releaseVersus): older than 1.0.1, newer than 1.0-rc1.
//
// It is deliberately separate from preReleaseSegmentRe: these rank the other
// way, and folding them together would make every Netty release look like a
// pre-release to isPreRelease, quietly breaking latest-stable selection.
var releaseSegmentRe = regexp.MustCompile(`(?i)^(final|ga|release)$`)

// servicePackSegmentRe matches SP<n>, which Maven ranks ABOVE the plain
// release (1.0-sp1 > 1.0). As a trailing segment that is already what the
// "an extra non-pre-release segment ranks its side higher" rule does, so this
// exists for the two places that cannot infer it: validateImageVersion, which
// would otherwise reject the segment as unparseable, and releaseVersus, where
// SP is aligned with a release qualifier and the string comparison would
// decide it on letter case alone.
var servicePackSegmentRe = regexp.MustCompile(`(?i)^sp[0-9]*$`)

var versionSegmentSplitRe = regexp.MustCompile(`[.\-_]+`)

var propertyRefRe = regexp.MustCompile(`\$\{([^}]+)\}`)

// artifact is one resolved member of a Maven dependency closure.
type artifact struct {
	Coord
	Version  string
	Fallback bool // true when Version came from latestStable, not the POM chain
}

// mavenMetadataXML mirrors the <metadata> document Maven Central serves at
// <group>/<artifact>/maven-metadata.xml.
type mavenMetadataXML struct {
	XMLName    xml.Name `xml:"metadata"`
	Versioning struct {
		Release  string `xml:"release"`
		Versions struct {
			Version []string `xml:"version"`
		} `xml:"versions"`
	} `xml:"versioning"`
}

// pomParent mirrors a POM's <parent> element.
type pomParent struct {
	GroupID    string `xml:"groupId"`
	ArtifactID string `xml:"artifactId"`
	Version    string `xml:"version"`
}

// pomDependency mirrors one <dependency> entry, whether under <dependencies>
// or <dependencyManagement>.
type pomDependency struct {
	GroupID    string `xml:"groupId"`
	ArtifactID string `xml:"artifactId"`
	Version    string `xml:"version"`
	Scope      string `xml:"scope"`
	Optional   string `xml:"optional"`
	Type       string `xml:"type"`
}

// pomProperty is one <properties> child; its element name is the property
// name, so XMLName is what makes the catch-all below useful.
type pomProperty struct {
	XMLName xml.Name
	Value   string `xml:",chardata"`
}

// pomXML mirrors the subset of a Maven POM this resolver needs.
type pomXML struct {
	XMLName    xml.Name   `xml:"project"`
	Parent     *pomParent `xml:"parent"`
	Properties struct {
		Entries []pomProperty `xml:",any"`
	} `xml:"properties"`
	DependencyManagement struct {
		Dependencies []pomDependency `xml:"dependencies>dependency"`
	} `xml:"dependencyManagement"`
	Dependencies []pomDependency `xml:"dependencies>dependency"`
}

// resolveClosure resolves seed's dependency closure at its latest stable
// release. It is resolveClosureAt with an empty pinned version.
func resolveClosure(d Doer, seed Coord) ([]artifact, error) {
	return resolveClosureAt(d, seed, "")
}

// resolveClosureAt resolves the full dependency closure for seed: seed itself
// at version (or at its latest stable release when version is empty),
// followed by a breadth-first walk of its compile/runtime, non-optional, jar
// dependencies, deduped by group:artifact with first-seen-wins. The seed is
// always result[0]. A pinned version only changes which release the seed
// itself is resolved at -- every dependency version still comes from that
// release's own POM and parent chain, exactly as when version is empty.
// Fallback is never set for the seed: an unpinned seed's version came from
// maven-metadata.xml (already a real, published release), and a pinned
// seed's version was named by the operator, not guessed, so there is nothing
// to report as a fallback either way.
//
// error is returned only for a systemic failure: an invalid seed coordinate,
// an invalid pinned version, the seed's own maven-metadata.xml being
// unreachable (version == "" only -- a pinned version never fetches it), or
// a pinned version that does not name a real release (the seed's own POM
// 404s at that exact version, which a typo'd --version deserves an
// actionable error for rather than the bare 404 libs.Download would
// otherwise surface as a single opaque Failure). Once the seed's version is
// known, an unreachable dependency (or parent) POM only stops that one
// branch from expanding further; the artifact stays in the closure at the
// version its declaring POM already gave it (or its own latest-stable
// fallback), since a real problem with that artifact's jar is caught later
// when libs.Download fetches it and records a per-artifact Failure.
// resolveClosureAt has no per-artifact error channel of its own, so this is
// the only place that distinction can be made.
func resolveClosureAt(d Doer, seed Coord, version string) ([]artifact, error) {
	if err := validateCoordPart(seed.Group); err != nil {
		return nil, fmt.Errorf("seed group %q: %w", seed.Group, err)
	}
	if err := validateCoordPart(seed.Artifact); err != nil {
		return nil, fmt.Errorf("seed artifact %q: %w", seed.Artifact, err)
	}

	pinned := version != ""
	seedVersion := version
	if pinned {
		if err := validateCoordPart(seedVersion); err != nil {
			return nil, fmt.Errorf("pinned version %q for %s:%s: %w", seedVersion, seed.Group, seed.Artifact, err)
		}
	} else {
		v, err := latestStable(d, seed)
		if err != nil {
			return nil, fmt.Errorf("resolving latest stable version for %s:%s: %w", seed.Group, seed.Artifact, err)
		}
		if err := validateCoordPart(v); err != nil {
			return nil, fmt.Errorf("latest stable version %q for %s:%s: %w", v, seed.Group, seed.Artifact, err)
		}
		seedVersion = v
	}

	result := []artifact{{Coord: seed, Version: seedVersion}}
	visited := map[string]bool{coordKey(seed): true}

	type queueItem struct {
		idx   int
		depth int
	}
	queue := []queueItem{{idx: 0, depth: 0}}

	for len(queue) > 0 {
		item := queue[0]
		queue = queue[1:]

		if item.depth >= maxResolveDepth || len(result) >= maxArtifacts {
			continue
		}

		cur := result[item.idx]
		pom, pomErr := fetchPOM(d, cur.Coord, cur.Version)
		if pomErr != nil {
			if pinned && item.idx == 0 {
				return nil, fmt.Errorf("pinned version %q not found for %s:%s: %w", seedVersion, seed.Group, seed.Artifact, pomErr)
			}
			continue
		}

		props, depMgmt, chainErr := resolveParentChain(d, pom)
		if chainErr != nil {
			continue
		}

		for _, dep := range pom.Dependencies {
			if len(result) >= maxArtifacts {
				break
			}
			if !acceptDependency(dep) {
				continue
			}
			if validateCoordPart(dep.GroupID) != nil || validateCoordPart(dep.ArtifactID) != nil {
				continue
			}

			depCoord := Coord{Group: dep.GroupID, Artifact: dep.ArtifactID}
			if visited[coordKey(depCoord)] {
				continue
			}

			depVersion, fallback, ok := resolveDependencyVersion(d, dep, depMgmt, props, cur.Group, cur.Version)
			if !ok {
				continue
			}

			visited[coordKey(depCoord)] = true
			result = append(result, artifact{Coord: depCoord, Version: depVersion, Fallback: fallback})
			queue = append(queue, queueItem{idx: len(result) - 1, depth: item.depth + 1})
		}
	}

	return result, nil
}

// acceptDependency applies the scope/optional/type filter: compile and
// runtime (an absent scope means compile), never optional=true, and jar or
// an absent type (never pom, and never a -sources/-javadoc classifier since
// none is ever requested here).
func acceptDependency(dep pomDependency) bool {
	scope := dep.Scope
	if scope == "" {
		scope = "compile"
	}
	if scope != "compile" && scope != "runtime" {
		return false
	}
	if strings.EqualFold(dep.Optional, "true") {
		return false
	}
	if dep.Type != "" && dep.Type != "jar" {
		return false
	}
	return true
}

// resolveDependencyVersion determines dep's version: its own literal
// version, or the merged dependencyManagement entry when dep declares none,
// with ${property} substitution against props (plus ${project.version} and
// ${project.groupId}, which refer to the POM that declared dep). If no
// version can be determined that way, it falls back to dep's own latest
// stable release and reports fallback=true. ok is false only when even the
// fallback lookup fails -- the dependency's own metadata is unreachable, or
// the version it names falls outside the allowed coordinate charset -- in
// which case the caller drops dep like any other single malformed item.
func resolveDependencyVersion(d Doer, dep pomDependency, depMgmt, props map[string]string, projectGroupID, projectVersion string) (version string, fallback bool, ok bool) {
	raw := dep.Version
	if raw == "" {
		raw = depMgmt[dep.GroupID+":"+dep.ArtifactID]
	}
	if resolved, resolvedOK := resolveProperty(raw, props, projectGroupID, projectVersion); resolvedOK {
		if validateCoordPart(resolved) == nil {
			return resolved, false, true
		}
	}

	latest, err := latestStable(d, Coord{Group: dep.GroupID, Artifact: dep.ArtifactID})
	if err != nil {
		return "", false, false
	}
	if err := validateCoordPart(latest); err != nil {
		return "", false, false
	}
	return latest, true, true
}

// resolveParentChain walks child's <parent> chain (capped at
// maxParentDepth, cycle-guarded), merging <properties> and
// <dependencyManagement> at every level with the nearer (more child-ward)
// level winning. dependencyManagement versions are kept raw (possibly still
// a ${property}); the caller resolves them against the returned props once
// the whole chain has been merged.
func resolveParentChain(d Doer, child *pomXML) (props map[string]string, depMgmt map[string]string, err error) {
	props = map[string]string{}
	depMgmt = map[string]string{}
	visitedParents := map[string]bool{}

	cur := child
	depth := 0
	for {
		for name, value := range extractProperties(cur) {
			if _, exists := props[name]; !exists {
				props[name] = value
			}
		}
		for _, dm := range cur.DependencyManagement.Dependencies {
			k := dm.GroupID + ":" + dm.ArtifactID
			if _, exists := depMgmt[k]; !exists {
				depMgmt[k] = dm.Version
			}
		}

		if cur.Parent == nil {
			return props, depMgmt, nil
		}

		depth++
		if depth > maxParentDepth {
			return props, depMgmt, fmt.Errorf("parent chain exceeds max depth %d", maxParentDepth)
		}

		p := cur.Parent
		if validateCoordPart(p.GroupID) != nil || validateCoordPart(p.ArtifactID) != nil || validateCoordPart(p.Version) != nil {
			return props, depMgmt, fmt.Errorf("parent coordinate %s:%s:%s outside the allowed charset", p.GroupID, p.ArtifactID, p.Version)
		}

		pKey := p.GroupID + ":" + p.ArtifactID + ":" + p.Version
		if visitedParents[pKey] {
			return props, depMgmt, fmt.Errorf("parent cycle detected at %s", pKey)
		}
		visitedParents[pKey] = true

		parentPOM, ferr := fetchPOM(d, Coord{Group: p.GroupID, Artifact: p.ArtifactID}, p.Version)
		if ferr != nil {
			return props, depMgmt, fmt.Errorf("fetching parent pom %s: %w", pKey, ferr)
		}
		cur = parentPOM
	}
}

func extractProperties(p *pomXML) map[string]string {
	m := make(map[string]string, len(p.Properties.Entries))
	for _, e := range p.Properties.Entries {
		m[e.XMLName.Local] = strings.TrimSpace(e.Value)
	}
	return m
}

// resolveProperty substitutes every ${...} reference in raw. ${project.version}
// and ${project.groupId} resolve to the declaring POM's own coordinate;
// anything else is looked up in props. ok is false when raw is empty or a
// reference cannot be resolved, so the caller falls back to latestStable.
func resolveProperty(raw string, props map[string]string, projectGroupID, projectVersion string) (string, bool) {
	if raw == "" {
		return "", false
	}
	if !strings.Contains(raw, "${") {
		return raw, true
	}
	resolved := propertyRefRe.ReplaceAllStringFunc(raw, func(m string) string {
		name := propertyRefRe.FindStringSubmatch(m)[1]
		switch name {
		case "project.version":
			return projectVersion
		case "project.groupId":
			return projectGroupID
		}
		if v, ok := props[name]; ok {
			return v
		}
		return m
	})
	if strings.Contains(resolved, "${") {
		return "", false
	}
	return resolved, true
}

// latestStable picks c's latest stable release from maven-metadata.xml:
// <release> if present and not itself a pre-release (and actually published
// in <versions> -- a metadata document naming a release that is not in its
// own version list is not trusted), else the highest surviving <version> by
// segment-wise comparison.
func latestStable(d Doer, c Coord) (string, error) {
	meta, err := fetchMetadataXML(d, c)
	if err != nil {
		return "", err
	}

	release := meta.Versioning.Release
	if release != "" && !isPreRelease(release) && versionPublished(release, meta.Versioning.Versions.Version) {
		return release, nil
	}

	best := ""
	for _, v := range meta.Versioning.Versions.Version {
		if isPreRelease(v) {
			continue
		}
		if best == "" || compareVersions(v, best) > 0 {
			best = v
		}
	}
	if best == "" {
		return "", fmt.Errorf("no stable version found for %s:%s in maven-metadata.xml", c.Group, c.Artifact)
	}
	return best, nil
}

func versionPublished(v string, list []string) bool {
	for _, x := range list {
		if x == v {
			return true
		}
	}
	return false
}

// isPreRelease reports whether any segment of v (split on '.', '-', '_')
// names a pre-release qualifier: SNAPSHOT, alpha/beta/rc/cr/pr/ea/preview,
// or M<n>, case-insensitive.
func isPreRelease(v string) bool {
	for _, seg := range versionSegmentSplitRe.Split(v, -1) {
		if preReleaseSegmentRe.MatchString(seg) {
			return true
		}
	}
	return false
}

// compareVersions orders two version strings segment-wise (split on '.',
// '-', '_'): numeric comparison when both segments are numeric, string
// comparison otherwise (see compareVersionSegment). A segment present on only
// one side ranks that side higher than the other, with two exceptions:
//
//   - a pre-release qualifier ranks LOWER, so "9.4.0.0-rc1" sorts below
//     "9.4.0.0" even though it has one more segment;
//   - a release qualifier (Final/GA/RELEASE) ranks EQUAL and is skipped, so
//     "4.1.135.Final" and "4.1.135" are the same version -- Maven treats those
//     words as aliases of the empty qualifier, and the image classpath is full
//     of them (Netty, Hibernate, JBoss).
//
// SP<n> is deliberately not special-cased: Maven ranks it above the plain
// release, which is exactly what the default "extra segment ranks higher"
// branch already does.
func compareVersions(a, b string) int {
	if a == b {
		return 0
	}
	as := versionSegmentSplitRe.Split(a, -1)
	bs := versionSegmentSplitRe.Split(b, -1)

	n := len(as)
	if len(bs) > n {
		n = len(bs)
	}
	for i := 0; i < n; i++ {
		haveA := i < len(as)
		haveB := i < len(bs)
		switch {
		case haveA && haveB:
			if c := compareVersionSegment(as[i], bs[i]); c != 0 {
				return c
			}
		case haveA && !haveB:
			if c, decided := rankExtraSegment(as[i]); decided {
				return c
			}
		case !haveA && haveB:
			// Same rule, mirrored: negating it keeps one definition of how an
			// extra segment ranks instead of two that can drift apart.
			if c, decided := rankExtraSegment(bs[i]); decided {
				return -c
			}
		}
	}
	return 0
}

// rankExtraSegment ranks a segment that exists on only one side of a
// comparison, from that side's point of view: -1 it sorts lower, +1 higher.
// decided is false when the segment carries no ordering information at all --
// a release qualifier, which means the same version -- and the caller moves on
// to the next segment rather than returning.
func rankExtraSegment(seg string) (rank int, decided bool) {
	switch {
	case releaseSegmentRe.MatchString(seg):
		return 0, false
	case preReleaseSegmentRe.MatchString(seg):
		return -1, true
	}
	return 1, true
}

// compareVersionSegment orders two aligned segments: numerically when both
// are numbers, otherwise by string.
//
// A release qualifier (Final/GA/RELEASE) is handled before that fallback,
// because it means "no qualifier at all" and so must sort BELOW any number --
// 1.0.Final is 1.0, which is older than 1.0.1. Left to strings.Compare it
// would win by lexicographic accident ("Final" > "1"), making a stale image
// entry look newer than a real release. Two release qualifiers are equal
// however they are spelled, since they all mean the same thing.
func compareVersionSegment(a, b string) int {
	an, aErr := strconv.Atoi(a)
	bn, bErr := strconv.Atoi(b)
	if aErr == nil && bErr == nil {
		switch {
		case an < bn:
			return -1
		case an > bn:
			return 1
		default:
			return 0
		}
	}
	aRel, bRel := releaseSegmentRe.MatchString(a), releaseSegmentRe.MatchString(b)
	switch {
	case aRel && bRel:
		return 0
	case aRel:
		if c, decided := releaseVersus(b); decided {
			return c
		}
	case bRel:
		if c, decided := releaseVersus(a); decided {
			return -c
		}
	}
	return strings.Compare(a, b)
}

// releaseVersus reports how a release qualifier compares against the segment
// it is aligned with, from the release qualifier's point of view.
//
// The qualifier means "no qualifier at all", which puts it BETWEEN a number
// and a pre-release: 1.0.Final is older than 1.0.1, and newer than 1.0-rc1.
// Getting only the first half of that right is the easy mistake -- it makes
// every .Final release sort below its own release candidates.
//
// A service pack is ranked explicitly rather than left to the string
// comparison, which would otherwise decide "sp1" vs "Final" on the accident
// that lowercase sorts after uppercase, and flip on "SP1" vs "final".
//
// decided is false for anything else, leaving the caller's string comparison
// in charge exactly as before.
func releaseVersus(other string) (rank int, decided bool) {
	if _, err := strconv.Atoi(other); err == nil {
		return -1, true
	}
	if servicePackSegmentRe.MatchString(other) {
		return -1, true
	}
	if preReleaseSegmentRe.MatchString(other) {
		return 1, true
	}
	return 0, false
}

// validateCoordPart is the conservative charset gate every group, artifact
// and version string parsed out of untrusted XML must pass before it is
// spliced into a URL or used to build a filesystem-adjacent name. It rejects
// empty strings, anything outside coordCharsetRe, and a ".." sequence (which
// the charset alone does not catch, since '.' is otherwise allowed).
func validateCoordPart(s string) error {
	if s == "" {
		return fmt.Errorf("coordinate part is empty")
	}
	if !coordCharsetRe.MatchString(s) {
		return fmt.Errorf("coordinate part %q contains characters outside the allowed set (letters, digits, '.', '-', '_')", s)
	}
	if strings.Contains(s, "..") {
		return fmt.Errorf("coordinate part %q contains a path-traversal sequence", s)
	}
	return nil
}

func coordKey(c Coord) string {
	return c.Group + ":" + c.Artifact
}

func groupPath(group string) string {
	return strings.ReplaceAll(group, ".", "/")
}

func metadataURL(c Coord) string {
	return fmt.Sprintf("%s/%s/%s/maven-metadata.xml", CentralBase, groupPath(c.Group), c.Artifact)
}

func pomURL(c Coord, version string) string {
	return fmt.Sprintf("%s/%s/%s/%s/%s-%s.pom", CentralBase, groupPath(c.Group), c.Artifact, version, c.Artifact, version)
}

// jarURL builds a's jar download URL on Maven Central. Central serves no
// redirects on the happy path, but any Doer that does follow one is
// responsible for re-validating https on every hop; jarURL only ever
// produces an https URL itself since CentralBase is https.
func jarURL(a artifact) string {
	return fmt.Sprintf("%s/%s/%s/%s/%s-%s.jar", CentralBase, groupPath(a.Group), a.Artifact, a.Version, a.Artifact, a.Version)
}

func fetchMetadataXML(d Doer, c Coord) (*mavenMetadataXML, error) {
	body, err := fetchXML(d, metadataURL(c))
	if err != nil {
		return nil, err
	}
	var meta mavenMetadataXML
	if err := xml.Unmarshal(body, &meta); err != nil {
		return nil, fmt.Errorf("parsing maven-metadata.xml for %s:%s: %w", c.Group, c.Artifact, err)
	}
	return &meta, nil
}

func fetchPOM(d Doer, c Coord, version string) (*pomXML, error) {
	if err := validateCoordPart(version); err != nil {
		return nil, fmt.Errorf("pom version for %s:%s: %w", c.Group, c.Artifact, err)
	}
	body, err := fetchXML(d, pomURL(c, version))
	if err != nil {
		return nil, err
	}
	var p pomXML
	if err := xml.Unmarshal(body, &p); err != nil {
		return nil, fmt.Errorf("parsing pom for %s:%s:%s: %w", c.Group, c.Artifact, version, err)
	}
	return &p, nil
}

// fetchXML issues an https GET through d and returns the body, capped at
// maxXMLBytes read from the wire (never trusting Content-Length).
func fetchXML(d Doer, rawURL string) ([]byte, error) {
	req, err := http.NewRequest(http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, fmt.Errorf("building request for %s: %w", redactURL(rawURL), err)
	}
	resp, err := d.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetching %s: %w", redactURL(rawURL), err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fetching %s: unexpected status %d", redactURL(rawURL), resp.StatusCode)
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, maxXMLBytes+1))
	if err != nil {
		return nil, fmt.Errorf("reading %s: %w", redactURL(rawURL), err)
	}
	if len(body) > maxXMLBytes {
		return nil, fmt.Errorf("response from %s exceeds %d bytes", redactURL(rawURL), maxXMLBytes)
	}
	return body, nil
}

// redactURL strips any userinfo before a URL is used in an error message, so
// a credential never reaches output. Every URL this file builds comes from
// CentralBase plus a validated coordinate, so there is never one to strip in
// practice; this is defensive.
func redactURL(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return rawURL
	}
	u.User = nil
	return u.String()
}
