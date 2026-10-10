package libs

import (
	"bufio"
	"bytes"
	"embed"
	"fmt"
	"io/fs"
	"os"
	"strconv"
	"strings"
)

// This file teaches the package what the connector image already ships, so
// Download can omit a Maven-resolved DEPENDENCY the image already provides
// at an equal-or-newer version instead of downloading a duplicate copy onto
// the libs mount. The seed artifact itself -- the IBM MQ client for the mq
// set, the logback encoder for the syslog set -- is NEVER a candidate for
// omission, no matter what an omit list says about it; that invariant is
// enforced by Download (libs.go), not here, but it exists for three reasons
// worth recording next to the mechanism it constrains:
//   - the seed is the jar the operator ran the command to get; if the image
//     already had it there would be no reason to run the command at all.
//   - "download jar syslog" must keep fetching the encoder even against an
//     image that already ships it (2.13.0 ships logstash-logback-encoder
//     8.0, 3.1.0 ships 9.0), because the same command also has to work
//     against an OLDER image that lacks it entirely, and there is no
//     per-image special case here.
//   - it makes a stale or hostile omit list unable to skip the one jar that
//     matters -- a crafted line such as "com.ibm.mq.jakarta.client-99999.jar"
//     would otherwise omit the MQ client itself.
//
// The omit list is a plain list of jar basenames -- one per line, no
// groupId, since that is all "find -name *.jar | xargs basename" can ever
// produce -- and loadImageLibs turns that into an artifact-base-name ->
// version lookup table, plus this load's provenance and the entries whose
// version could not be trusted. imageSatisfies then compares a resolved
// dependency against that
// table using the same compareVersions this package already uses to pick a
// latest-stable release (maven.go); this file never re-implements version
// comparison, and it never trusts a version compareVersions cannot compare
// meaningfully -- see validateImageVersion.
//
// Matching a resolved artifact against the image is therefore always by jar
// artifact base name, never by groupId, because a jar filename does not
// carry one. Two Maven artifacts that happen to share a base name across
// different groupIds are indistinguishable by name alone and can only be
// told apart by version -- tools.jackson.core:jackson-core versus
// com.fasterxml.jackson.core:jackson-core is a live example in this very
// closure. That is a real limitation, not a bug: it is also why Jackson 3
// still downloads correctly despite the collision -- its 3.x version compares
// higher than the image's Jackson 2 2.x copy, so imageSatisfies reports
// not-provided regardless of the groupId mismatch it can not see.
//
// Finally: an omit list is a DECLARATION BY THE OPERATOR about what their
// image contains, not something this package can independently verify. A
// well-formed but wrong entry -- "com.ibm.mq.jakarta.client-99999.jar" against
// an image that does not actually ship version 99999 of anything -- is
// trusted for every non-seed dependency exactly as declared; only the seed
// exemption above and the unparseable-version rejection below narrow that
// trust. A wrong list yields wrong omissions for dependencies by design.

// embeddedListImage is the connector image every built-in jar list was
// captured from.
const embeddedListImage = "solace-pubsub-connector-ibmmq"

// embeddedList is one jar list captured from a connector image and built into
// the binary, with the connector releases it is known to describe. A list is
// named for the single tag its bytes came from, but it describes a RANGE: a
// classpath can stay put across releases (2.13.0's capture and 2.14.1's are
// byte-for-byte identical), while connector 3.x moved to Spring Boot 4 and
// Jackson 3, which no 2.x list can speak for -- and within 3.x, 3.2.0 already
// ships newer versions of the same jars than 3.1.0. Keeping the captured tag
// apart from the range keeps that distinction in the type rather than in a
// comment: the file name stays an honest record of where the bytes came from,
// and the range is what a deployed tag is checked against.
type embeddedList struct {
	capturedAt string // the tag the list was captured from; names its file
	from       string // the earliest release it describes
	before     string // the first release it does not describe; "" while none is known
}

// embeddedLists are the built-in jar lists, oldest first, with ranges that
// never overlap. A range is verified evidence, not a guess: lower a from only
// once a capture from an older release proves that release matches, and set a
// before the moment one proves a release does NOT. Adding a list is a small
// change -- capture it with the command in an existing list's header, save it
// as imagelibs/<image>-<tag>.list, and add its row here -- and only needed
// when the contents actually differ from the newest list's, since an identical
// capture means that list already covers the release and only its range moves.
var embeddedLists = []embeddedList{
	{capturedAt: "2.13.0", from: "2.10.0", before: "3.0.0"},
	// No 3.0.x release has been captured, so the 3.x line starts at 3.1.0: a
	// 3.0.x deployment falls outside every range and is warned about.
	{capturedAt: "3.1.0", from: "3.1.0", before: "3.2.0"},
	{capturedAt: "3.2.0", from: "3.2.0"},
}

//go:embed imagelibs/*.list
var embeddedListFiles embed.FS

// embeddedListFS is where loadImageLibs reads a built-in list from: the
// embedded files, behind fs.FS so a test can substitute a broken one.
var embeddedListFS fs.FS = embeddedListFiles

// name is the list's display name: its file name without the extension, the
// image and the tag it was captured from.
func (l embeddedList) name() string {
	return embeddedListImage + "-" + l.capturedAt
}

// file is the list's path inside embeddedListFS.
func (l embeddedList) file() string {
	return "imagelibs/" + l.name() + ".list"
}

// covers reports whether the release tag names falls inside the list's range.
func (l embeddedList) covers(tag string) bool {
	return compareVersions(tag, l.from) >= 0 && (l.before == "" || compareVersions(tag, l.before) < 0)
}

// describes renders the list's range for the report: "2.10.0 and later,
// before 3.0.0", or "3.1.0 and later" for a line with no known end.
func (l embeddedList) describes() string {
	if l.before == "" {
		return l.from + " and later"
	}
	return l.from + " and later, before " + l.before
}

// maxOmitListLineBytes bounds one line of an omit list. bufio.Scanner's
// default token limit is 64KB; a well-formed line is a few dozen bytes, so
// this only guards against a pathological line (two entries concatenated by
// a missing newline, or a stray non-newline-terminated byte run) causing the
// whole load to abort. Without an explicit buffer that pathological line
// turns into bufio.ErrTooLong -- a systemic failure -- which would break this
// function's own promise that one malformed line is skipped, not fatal.
const maxOmitListLineBytes = 1024 * 1024

// imageLibs maps a jar artifact base name (e.g. "bcprov-jdk18on") to the
// version string the image ships that jar at (e.g. "1.84"). See the package
// doc above this file for why base name, not groupId:artifactId, is the only
// key available.
type imageLibs map[string]string

// loadedImageLibs is everything one loadImageLibs call recovers from an omit
// list: the parsed map imageSatisfies compares against, a human-readable
// Provenance identifying which list produced it (so the CLI can report which
// list is in use), and Rejected -- artifact base name to the reason its
// version could not be trusted (see validateImageVersion), for every line
// recognized as a jar reference but left out of Libs. Bundling all three in
// the loader's return value, rather than a separate accessor call, keeps
// loadImageLibs the single place that ever reads the list.
//
// Rejected is a map rather than a list of ready-made messages because a
// rejection only matters when a closure actually asks about that artifact:
// warning at load time reported every rejected line on every run, so a
// `download jar mq` told the operator about netty and hibernate-validator,
// which no mq closure has ever referenced. The caller looks an artifact up
// here only when it is missing from Libs, and warns then -- at the one moment
// the rejection changed what gets downloaded.
type loadedImageLibs struct {
	Libs       imageLibs
	Provenance string
	Rejected   map[string]string
}

// omitListProvenance names which omit list is in effect: the operator's
// --omit-lib-file path verbatim when one was supplied, or the built-in list's
// own name -- derived from the same capturedAt its file path is, so the two
// can never drift apart into independent copies of the image name.
func omitListProvenance(omitLibFile string, builtin embeddedList) string {
	if omitLibFile != "" {
		return omitLibFile
	}
	return builtin.name()
}

// imageNameTag splits a full image reference into its bare image name and
// tag: "solace/solace-pubsub-connector-ibmmq:2.14.1" becomes
// ("solace-pubsub-connector-ibmmq", "2.14.1").
//
// The tag separator is looked for in the LAST path element, not at the first
// colon in the whole reference, because a registry host may carry a port
// ("registry.internal:5000/team/connector:2.14.1").
//
// ok is false when the reference names no comparable tag -- absent, or pinned
// by digest, which names no release any list could have been captured under.
func imageNameTag(ref string) (string, string, bool) {
	if ref == "" || strings.Contains(ref, "@") {
		return "", "", false
	}
	last := ref
	if i := strings.LastIndexByte(ref, '/'); i >= 0 {
		last = ref[i+1:]
	}
	name, tag, found := strings.Cut(last, ":")
	if !found || name == "" || tag == "" {
		return "", "", false
	}
	return name, tag, true
}

// releaseTag reports whether tag names a release a list's range can be
// compared with: one that starts with a digit. A tag such as "latest" names
// none -- compared as a version it would sort after every number, and be
// judged covered by accident.
func releaseTag(tag string) bool {
	return tag != "" && tag[0] >= '0' && tag[0] <= '9'
}

// connectorRelease reads the connector release off the image reference
// env.yaml deploys: its tag, and the major number the tag leads with. ok is
// false when there is no release to read -- no reference, a digest pin, a
// different image, or a tag such as "latest".
func connectorRelease(ref string) (string, int, bool) {
	name, tag, ok := imageNameTag(ref)
	if !ok || name != embeddedListImage || !releaseTag(tag) {
		return "", 0, false
	}
	end := strings.IndexFunc(tag, func(r rune) bool { return r < '0' || r > '9' })
	if end < 0 {
		end = len(tag)
	}
	major, err := strconv.Atoi(tag[:end])
	if err != nil {
		return "", 0, false
	}
	return tag, major, true
}

// listRanges names what the built-in lists describe, for a mismatch note:
// "2.10.0 and later, before 3.0.0; 3.1.0 and later".
func listRanges() string {
	ranges := make([]string, 0, len(embeddedLists))
	for _, l := range embeddedLists {
		ranges = append(ranges, l.describes())
	}
	return strings.Join(ranges, "; ")
}

// builtinList picks the built-in jar list for the connector image env.yaml
// deploys, with a note when no list is known to describe that image, which
// makes every omission in this run a claim about the wrong classpath. A
// connector release inside a list's range gets that list and no note.
// Anything else gets the nearest list -- the newest captured at or before its
// release, the oldest for a release older than every capture, or the newest
// when there is no release to go by -- and a note naming the reference, what
// it was judged against and the remedy. A differing tag is not itself a
// mismatch: the ranges are what decide.
//
// Nothing declared at all is silent, since download runs fine with no config
// and the report still names the list it used. A reference with no comparable
// release -- no tag, a digest pin, "latest" -- warns rather than staying
// quiet: "we cannot tell whether these omissions apply" is the same operator
// problem as "they do not", and silence is what let a 2.13.0 list judge a
// 2.14.1 deployment unnoticed in the first place.
func builtinList(deployedImage string) (embeddedList, string) {
	newest := embeddedLists[len(embeddedLists)-1]
	if deployedImage == "" {
		return newest, ""
	}
	const remedy = "Pass --omit-lib-file with a list captured from that image, or --include-provided to skip omission entirely"
	name, tag, ok := imageNameTag(deployedImage)
	switch {
	case !ok || !releaseTag(tag):
		return newest, fmt.Sprintf("env.yaml deploys %s, which names no release to check against the built-in %s jar lists (%s) -- omissions are approximate. %s",
			deployedImage, embeddedListImage, listRanges(), remedy)
	case name != embeddedListImage:
		return newest, fmt.Sprintf("env.yaml deploys %s but the built-in jar lists describe %s -- every omission above is a claim about that image, not the one being deployed. %s",
			deployedImage, embeddedListImage, remedy)
	}
	nearest := embeddedLists[0]
	for _, l := range embeddedLists {
		if l.covers(tag) {
			return l, ""
		}
		if compareVersions(tag, l.from) >= 0 {
			nearest = l
		}
	}
	return nearest, fmt.Sprintf("env.yaml deploys %s, which no built-in jar list describes (they cover %s) -- every omission above is judged against %s, so it may name a jar that image does not ship. %s",
		deployedImage, listRanges(), nearest.name(), remedy)
}

// loadImageLibs reads an omit list, one jar basename per line. A blank line
// (after trimming surrounding whitespace) or a line starting with '#' is
// skipped. A line that does not fit the "<artifact>-<version>.jar" shape
// splitJarBasename expects is skipped too, as one malformed item, rather
// than failing the whole load -- this repo's fail-loud-vs-skip rule reserves
// the loud failure for a systemic problem, and one odd line in a 160-entry
// classpath dump is not that. A line that DOES split but carries a version
// validateImageVersion rejects as unparseable is likewise skipped rather
// than failing the load, but it is not silent: it is recorded in
// loadedImageLibs.Rejected, so that if this closure goes on to ask about that
// artifact the caller can report the rejection rather than let the operator
// assume the comparison was trusted (see validateImageVersion for why
// trusting it would be unsafe). A rejection no closure ever consults changed
// nothing and is never reported.
//
// omitLibFile REPLACES the built-in list completely when non-empty -- there
// is no merging with it, so an omit list containing nothing (an empty file)
// omits nothing. An empty omitLibFile loads builtin, the built-in list
// builtinList picked for the image being deployed; its header comment names
// the image it was captured from. A non-empty omitLibFile that cannot be read
// IS a systemic error: the caller named a specific file with --omit-lib-file
// and it does not exist, which is worth failing loud on rather than silently
// falling back to a built-in list.
func loadImageLibs(omitLibFile string, builtin embeddedList) (loadedImageLibs, error) {
	provenance := omitListProvenance(omitLibFile, builtin)
	var content []byte
	if omitLibFile != "" {
		b, err := os.ReadFile(omitLibFile)
		if err != nil {
			return loadedImageLibs{}, fmt.Errorf("reading omit list %q: %w", omitLibFile, err)
		}
		content = b
	} else {
		b, err := fs.ReadFile(embeddedListFS, builtin.file())
		if err != nil {
			return loadedImageLibs{}, fmt.Errorf("reading %q: %w", builtin.file(), err)
		}
		content = b
	}

	result := loadedImageLibs{
		Libs:       imageLibs{},
		Provenance: provenance,
		Rejected:   map[string]string{},
	}
	scanner := bufio.NewScanner(bytes.NewReader(content))
	scanner.Buffer(make([]byte, 0, 64*1024), maxOmitListLineBytes)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		art, version, ok := splitJarBasename(line)
		if !ok {
			continue
		}
		if err := validateImageVersion(version); err != nil {
			result.Rejected[art] = fmt.Sprintf(
				"%s-%s.jar: version %q rejected as unparseable (%v) -- treated as not provided by the image",
				art, version, version, err)
			continue
		}
		result.Libs[art] = version
	}
	if err := scanner.Err(); err != nil {
		return loadedImageLibs{}, fmt.Errorf("reading omit list %q: %w", provenance, err)
	}
	return result, nil
}

// splitJarBasename splits a jar filename (e.g. "bcprov-jdk18on-1.84.jar")
// into its artifact base name and version. A jar filename carries no
// groupId, so the base name this recovers is at best what Maven Central
// would call the artifactId -- never the group.
//
// The rule: strip the ".jar" suffix (case-insensitively), then split at the
// FIRST '-' that is immediately followed by a digit. Maven release versions
// conventionally start with a digit, so the first digit-led hyphen is the
// artifact/version boundary even when the artifact name itself is
// hyphenated ("bcprov-jdk18on-1.84" splits after "jdk18on", not after
// "bcprov") and even when the version goes on to contain more hyphens of
// its own ("jcip-annotations-1.0-1", "opentelemetry-proto-1.5.0-alpha": the
// first hyphen is never followed by a digit, the second is, so that is
// where each splits).
//
// A basename with no digit-led hyphen at all cannot be split; ok is false
// and the caller skips it as one malformed item. jrt-fs.jar is the one real
// example in this image's classpath: "jrt-fs" has a hyphen but it is never
// followed by a digit, so it does not split and is skipped. That is the
// right outcome for it specifically -- it is a JDK-internal file with no
// Maven version, never a jar any resolved closure would request, so leaving
// it out of the map can never cause a false match either way.
func splitJarBasename(name string) (artifactName, version string, ok bool) {
	const suffix = ".jar"
	if len(name) <= len(suffix) || !strings.EqualFold(name[len(name)-len(suffix):], suffix) {
		return "", "", false
	}
	base := name[:len(name)-len(suffix)]

	for i := 0; i+1 < len(base); i++ {
		if base[i] == '-' && base[i+1] >= '0' && base[i+1] <= '9' {
			return base[:i], trimJarClassifier(base[i+1:]), true
		}
	}
	return "", "", false
}

// trimJarClassifier drops a trailing Maven classifier from the version tail
// splitJarBasename recovered. A jar filename is
// "<artifact>-<version>[-<classifier>].jar", and nothing in the name marks
// where the version stops -- so the version is taken to run for as long as its
// hyphen-separated components keep looking like version components (numeric,
// or a qualifier compareVersionSegment can order), and the first component
// that does not is where the classifier begins.
//
// This is what makes netty's native jars usable:
// "netty-transport-native-epoll-4.1.135.Final-linux-x86_64" yields
// "4.1.135.Final" rather than a version with "linux" and "x86" in it, which no
// comparison could order and validateImageVersion would reject outright.
//
// The two cases splitJarBasename's own comment calls out still hold, because
// both of their tails are made only of version-shaped components:
// "jcip-annotations-1.0-1" keeps "1.0-1" (numeric) and
// "opentelemetry-proto-1.5.0-alpha" keeps "1.5.0-alpha" (a qualifier).
//
// Consequence worth naming: dropping the classifier also drops it from the
// match, so a classified image jar can now satisfy a request for the plain
// artifact at that version. That cannot arise today -- jarURL only ever builds
// unclassified "<artifact>-<version>.jar" URLs, so no resolved closure member
// carries a classifier -- and it is the same coarseness the base-name match
// already has (a jar filename carries no groupId either). It would need
// revisiting if closure resolution ever learned about classifiers.
func trimJarClassifier(version string) string {
	parts := strings.Split(version, "-")
	end := 1 // the first component led with a digit, so it is always the version's
	for ; end < len(parts); end++ {
		if !versionComponent(parts[end]) {
			break
		}
	}
	return strings.Join(parts[:end], "-")
}

// versionComponent reports whether one hyphen-separated component of a jar
// filename's tail still belongs to the version. Its dot-separated segments
// must each be numeric or a recognized qualifier -- the same test
// validateImageVersion applies, reused so the two can never disagree about
// what a version looks like.
func versionComponent(part string) bool {
	if part == "" {
		return false
	}
	return validateImageVersion(part) == nil
}

// validateImageVersion reports whether v is a version string compareVersions
// can compare meaningfully. It first runs v through validateCoordPart, the
// same conservative charset gate (maven.go) every XML-sourced coordinate
// part must pass, then additionally requires every '.'/'-'/'_'-separated
// segment (the same split versionSegmentSplitRe uses in compareVersions) to
// be all-numeric or a recognized qualifier -- pre-release
// (preReleaseSegmentRe), release (releaseSegmentRe: the Final/GA/RELEASE the
// image classpath is full of), or a service pack (servicePackSegmentRe) --
// the same shapes compareVersionSegment knows how to order.
//
// This exists because compareVersionSegment falls back to strings.Compare
// whenever either side fails strconv.Atoi, with no signal that the result is
// meaningless rather than a real ordering: compareVersionSegment("9zzz...",
// "3") returns positive purely from lexicographic accident, which would make
// an omit-list entry like "jackson-core-9zzzzzzzzzzz.jar" register as a
// newer version than any real jackson-core release and silently omit it. A
// version that fails this check must be treated by the caller as NOT
// PROVIDED (the safe direction -- the download proceeds), never as
// comparable.
func validateImageVersion(v string) error {
	if err := validateCoordPart(v); err != nil {
		return err
	}
	for _, seg := range versionSegmentSplitRe.Split(v, -1) {
		if seg == "" {
			continue
		}
		if _, err := strconv.Atoi(seg); err == nil {
			continue
		}
		if preReleaseSegmentRe.MatchString(seg) || releaseSegmentRe.MatchString(seg) || servicePackSegmentRe.MatchString(seg) {
			continue
		}
		return fmt.Errorf("segment %q is neither numeric nor a recognized version qualifier", seg)
	}
	return nil
}

// imageSatisfies reports whether libs already provides a at a version at
// least a.Version, using the same compareVersions maven.go already defines
// for picking a latest-stable release. It returns provided=true only when
// the image has a matching base name AND that version compares >= a's; the
// version string returned is the one the image has (for the omission
// message this package's caller renders), and is only meaningful when
// provided is true -- an absent entry returns "".
//
// Every version stored in libs has already passed validateImageVersion in
// loadImageLibs, so imageSatisfies itself never needs to re-validate; a line
// whose version was rejected as unparseable was never added to the map and
// so is already, correctly, absent here.
//
// imageSatisfies makes no exception for a seed artifact -- the seed is never
// a candidate for omission in the first place, and skipping that check is
// the caller's (libs.go's) responsibility, applied before an artifact ever
// reaches this function.
func imageSatisfies(libs imageLibs, a artifact) (bool, string) {
	have, ok := libs[a.Artifact]
	if !ok {
		return false, ""
	}
	return compareVersions(have, a.Version) >= 0, have
}
