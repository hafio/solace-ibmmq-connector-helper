package libs

import (
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

// --- splitJarBasename ---

func TestSplitJarBasename(t *testing.T) {
	cases := []struct {
		name         string
		in           string
		wantArtifact string
		wantVersion  string
		wantOK       bool
	}{
		{"hyphenated artifact name", "bcprov-jdk18on-1.84.jar", "bcprov-jdk18on", "1.84", true},
		{"dotted artifact name", "jakarta.jms-api-3.1.0.jar", "jakarta.jms-api", "3.1.0", true},
		{"short artifact, date-like version", "json-20250517.jar", "json", "20250517", true},
		{"hyphen inside the version itself", "jcip-annotations-1.0-1.jar", "jcip-annotations", "1.0-1", true},
		{"Final qualifier", "hibernate-validator-8.0.3.Final.jar", "hibernate-validator", "8.0.3.Final", true},
		{"alpha qualifier with hyphen", "opentelemetry-proto-1.5.0-alpha.jar", "opentelemetry-proto", "1.5.0-alpha", true},
		// The classifier is dropped: "linux" and "x86" are not version segments,
		// and keeping them made validateImageVersion reject the whole entry, so
		// the image was recorded as not having a jar it plainly ships.
		{"classifier stripped", "netty-transport-native-epoll-4.1.135.Final-linux-x86_64.jar", "netty-transport-native-epoll", "4.1.135.Final", true},
		{"classifier stripped, no qualifier", "netty-common-4.1.135-linux-aarch_64.jar", "netty-common", "4.1.135", true},
		{"sources classifier", "guava-33.0.0-sources.jar", "guava", "33.0.0", true},
		{"underscore in version", "org.apache.servicemix.bundles.jzlib-1.1.3_2.jar", "org.apache.servicemix.bundles.jzlib", "1.1.3_2", true},
		{"no digit-led hyphen at all", "jrt-fs.jar", "", "", false},
		{"plain hyphenated name and version", "spring-boot-3.5.16.jar", "spring-boot", "3.5.16", true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			gotArtifact, gotVersion, gotOK := splitJarBasename(c.in)
			if gotOK != c.wantOK {
				t.Fatalf("splitJarBasename(%q) ok = %v, want %v", c.in, gotOK, c.wantOK)
			}
			if !gotOK {
				return
			}
			if gotArtifact != c.wantArtifact || gotVersion != c.wantVersion {
				t.Errorf("splitJarBasename(%q) = (%q, %q), want (%q, %q)", c.in, gotArtifact, gotVersion, c.wantArtifact, c.wantVersion)
			}
		})
	}
}

func TestSplitJarBasenameRejectsNonJar(t *testing.T) {
	_, _, ok := splitJarBasename("bcprov-jdk18on-1.84.txt")
	if ok {
		t.Fatal("want ok = false for a non-.jar name")
	}
}

// --- loadImageLibs ---

func writeLibsFile(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "libs.list")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestLoadImageLibsSkipsCommentsAndBlankLines(t *testing.T) {
	path := writeLibsFile(t, "# a header comment\n\n  \nbcprov-jdk18on-1.84.jar\n")
	loaded, err := loadImageLibs(path, embeddedLists[0])
	if err != nil {
		t.Fatalf("loadImageLibs: %v", err)
	}
	libs := loaded.Libs
	if len(libs) != 1 || libs["bcprov-jdk18on"] != "1.84" {
		t.Errorf("libs = %v, want exactly bcprov-jdk18on -> 1.84", libs)
	}
}

func TestLoadImageLibsSkipsUnsplittableLineWithoutFailing(t *testing.T) {
	path := writeLibsFile(t, "jrt-fs.jar\nbcprov-jdk18on-1.84.jar\n")
	loaded, err := loadImageLibs(path, embeddedLists[0])
	if err != nil {
		t.Fatalf("loadImageLibs: %v", err)
	}
	libs := loaded.Libs
	if _, ok := libs["jrt-fs"]; ok {
		t.Errorf("libs contains jrt-fs, want it skipped as unsplittable")
	}
	if libs["bcprov-jdk18on"] != "1.84" {
		t.Errorf("libs[bcprov-jdk18on] = %q, want 1.84", libs["bcprov-jdk18on"])
	}
}

func TestLoadImageLibsToleratesSurroundingWhitespace(t *testing.T) {
	path := writeLibsFile(t, "  bcprov-jdk18on-1.84.jar  \r\n")
	loaded, err := loadImageLibs(path, embeddedLists[0])
	if err != nil {
		t.Fatalf("loadImageLibs: %v", err)
	}
	libs := loaded.Libs
	if libs["bcprov-jdk18on"] != "1.84" {
		t.Errorf("libs[bcprov-jdk18on] = %q, want 1.84", libs["bcprov-jdk18on"])
	}
}

func TestLoadImageLibsBadPathIsError(t *testing.T) {
	_, err := loadImageLibs(filepath.Join(t.TempDir(), "does-not-exist.list"), embeddedLists[0])
	if err == nil {
		t.Fatal("want error for a nonexistent path")
	}
}

func TestLoadImageLibsEmbeddedDefault(t *testing.T) {
	loaded, err := loadImageLibs("", embeddedLists[0])
	if err != nil {
		t.Fatalf("loadImageLibs(\"\"): %v", err)
	}
	libs := loaded.Libs

	cases := []struct {
		artifactName, version string
	}{
		{"bcprov-jdk18on", "1.84"},
		{"jakarta.jms-api", "3.1.0"},
		{"json", "20250517"},
		{"logstash-logback-encoder", "8.0"},
	}
	for _, c := range cases {
		if got := libs[c.artifactName]; got != c.version {
			t.Errorf("embedded libs[%q] = %q, want %q", c.artifactName, got, c.version)
		}
	}

	if _, ok := libs["com.ibm.mq.jakarta.client"]; ok {
		t.Errorf("embedded libs contains com.ibm.mq.jakarta.client, want it absent (that is the licensing carve-out)")
	}
}

// --- imageSatisfies ---

func TestImageSatisfies(t *testing.T) {
	libs := imageLibs{
		"jakarta.jms-api": "3.1.0",
		"bcprov-jdk18on":  "1.84",
	}
	cases := []struct {
		name         string
		a            artifact
		wantProvided bool
		wantHave     string
	}{
		{"image has a newer version", artifact{Coord: Coord{Artifact: "jakarta.jms-api"}, Version: "3.0.0"}, true, "3.1.0"},
		{"image has an equal version", artifact{Coord: Coord{Artifact: "bcprov-jdk18on"}, Version: "1.84"}, true, "1.84"},
		{"image has an older version", artifact{Coord: Coord{Artifact: "bcprov-jdk18on"}, Version: "1.85"}, false, "1.84"},
		{"image does not have it at all", artifact{Coord: Coord{Artifact: "com.ibm.mq.jakarta.client"}, Version: "10.0.0.0"}, false, ""},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			provided, have := imageSatisfies(libs, c.a)
			if provided != c.wantProvided || have != c.wantHave {
				t.Errorf("imageSatisfies(%+v) = (%v, %q), want (%v, %q)", c.a, provided, have, c.wantProvided, c.wantHave)
			}
		})
	}
}

// TestValidateImageVersionQualifiers covers which version shapes the omit list
// may carry. The accepted set grew to include the RELEASE qualifiers, which are
// how most of the shipped image classpath spells a stable version -- rejecting
// them recorded the image as not having jars it plainly ships.
//
// The rejections matter just as much and must not loosen with them: the whole
// reason this gate exists is that compareVersionSegment falls back to
// strings.Compare, so an entry like "jackson-core-9zzzzzzzzzzz.jar" would
// otherwise compare as newer than any real release and silently suppress it.
func TestValidateImageVersionQualifiers(t *testing.T) {
	accepted := []string{
		"4.1.135.Final", "8.0.3.Final", "3.6.3.Final", // netty, hibernate, jboss-logging
		"5.3.31.RELEASE", "1.0.GA", "1.0-sp1", "2.0.SP2",
		"1.84", "3.1.0", "20250517", "1.1.3_2", // plain numerics still fine
		"3.0-rc5", "1.0-SNAPSHOT", "5.0.M1", // pre-release still fine
	}
	for _, v := range accepted {
		if err := validateImageVersion(v); err != nil {
			t.Errorf("validateImageVersion(%q) = %v, want accepted", v, err)
		}
	}

	rejected := []string{
		"9zzzzzzzzzzz",        // the case this gate exists for
		"4.1.135.Final-linux", // a classifier is not a version (splitJarBasename strips it)
		"1.0.totally-made-up", // an unknown word is not assumed orderable
		"latest",              // not a version at all
	}
	for _, v := range rejected {
		if err := validateImageVersion(v); err == nil {
			t.Errorf("validateImageVersion(%q) = nil, want rejected", v)
		}
	}
}

// TestEmbeddedOmitListFullyParses holds every shipped list to the standard
// the code expects of it: every line either is not a jar reference at all, or
// splits and carries a version the comparator can order.
//
// This is what the 15 warnings on every `download jar mq` run were telling us,
// unread. A future capture from a newer image that reintroduces an unorderable
// shape now fails the build instead of printing noise no one acts on.
func TestEmbeddedOmitListFullyParses(t *testing.T) {
	for _, l := range embeddedLists {
		t.Run(l.capturedAt, func(t *testing.T) {
			loaded, err := loadImageLibs("", l)
			if err != nil {
				t.Fatalf("loadImageLibs: %v", err)
			}
			if len(loaded.Rejected) != 0 {
				t.Errorf("the %s list has %d unparseable entries, want none: %v",
					l.name(), len(loaded.Rejected), loaded.Rejected)
			}
			// Sanity: the list did actually load, so a future bug that empties
			// it cannot make the assertion above pass vacuously.
			if len(loaded.Libs) < 50 {
				t.Errorf("the %s list parsed only %d entries, expected the full image classpath", l.name(), len(loaded.Libs))
			}
			for _, want := range []string{"netty-common", "hibernate-validator", "jakarta.jms-api"} {
				if _, ok := loaded.Libs[want]; !ok {
					t.Errorf("the %s list has no entry for %q", l.name(), want)
				}
			}
			if loaded.Provenance != l.name() {
				t.Errorf("Provenance = %q, want the list's own name %q", loaded.Provenance, l.name())
			}
		})
	}
}

// TestEmbeddedListsTable keeps the table and the embedded files in step: the
// rows run oldest first with ranges that never overlap, every row names a file
// that is really embedded, and every embedded list has a row -- so a capture
// dropped into imagelibs/ without one, or a row added before its capture,
// fails the build rather than a download.
func TestEmbeddedListsTable(t *testing.T) {
	if len(embeddedLists) == 0 {
		t.Fatal("no built-in jar list at all")
	}
	rows := map[string]bool{}
	for i, l := range embeddedLists {
		if !releaseTag(l.from) || (l.before != "" && compareVersions(l.from, l.before) >= 0) {
			t.Errorf("%s: range %q..%q is not a release followed by a later one", l.capturedAt, l.from, l.before)
		}
		if !l.covers(l.capturedAt) {
			t.Errorf("%s: the list does not cover the release it was captured from (%s)", l.capturedAt, l.describes())
		}
		if i > 0 {
			if prev := embeddedLists[i-1]; prev.before == "" || compareVersions(l.from, prev.before) < 0 {
				t.Errorf("%s: starts at %s, inside %s's range (%s)", l.capturedAt, l.from, prev.capturedAt, prev.describes())
			}
		}
		if _, err := fs.Stat(embeddedListFiles, l.file()); err != nil {
			t.Errorf("%s: %s is not embedded: %v", l.capturedAt, l.file(), err)
		}
		rows[l.file()] = true
		// A capture identical to the list before it means that list already
		// describes the release, so its range should have been widened rather
		// than a second copy built in.
		if i > 0 {
			prev := embeddedLists[i-1]
			a, aerr := loadImageLibs("", prev)
			b, berr := loadImageLibs("", l)
			if aerr != nil || berr != nil {
				t.Fatalf("loading %s / %s: %v / %v", prev.name(), l.name(), aerr, berr)
			}
			if reflect.DeepEqual(a.Libs, b.Libs) {
				t.Errorf("%s is identical to %s: widen %s's range instead of shipping a second copy", l.name(), prev.name(), prev.name())
			}
		}
	}
	entries, err := fs.ReadDir(embeddedListFiles, "imagelibs")
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range entries {
		if f := "imagelibs/" + e.Name(); !rows[f] {
			t.Errorf("%s is embedded but has no row in embeddedLists", f)
		}
	}
}

// TestEmbeddedListsMatchTheirConnectorLine checks each built-in list against
// the connector line its capture tag names: 2.x ships Spring Boot 3 and
// Jackson 2, 3.x Spring Boot 4 and Jackson 3. A capture saved under the wrong
// tag fails here, before download judges a deployment against the wrong
// generation. A list for a line with no expectation yet fails too, so a new
// line's generation is stated here when its first list is added.
func TestEmbeddedListsMatchTheirConnectorLine(t *testing.T) {
	generation := map[string]map[string]string{
		"2": {"spring-boot": "3", "jackson-databind": "2"},
		"3": {"spring-boot": "4", "jackson-databind": "3"},
	}
	for _, l := range embeddedLists {
		t.Run(l.capturedAt, func(t *testing.T) {
			line, _, _ := strings.Cut(l.capturedAt, ".")
			want, ok := generation[line]
			if !ok {
				t.Fatalf("no generation recorded for connector %s.x; add one to this test", line)
			}
			loaded, err := loadImageLibs("", l)
			if err != nil {
				t.Fatalf("loadImageLibs: %v", err)
			}
			for art, major := range want {
				got := loaded.Libs[art]
				if gotMajor, _, _ := strings.Cut(got, "."); gotMajor != major {
					t.Errorf("%s ships %s %q, want major %s for connector %s.x", l.name(), art, got, major, line)
				}
			}
		})
	}
}

// TestEmbeddedListRange pins both ends of a range -- from included, before
// not -- and how the report words it, with and without a known end.
func TestEmbeddedListRange(t *testing.T) {
	closed := embeddedList{capturedAt: "2.13.0", from: "2.10.0", before: "3.0.0"}
	for tag, want := range map[string]bool{"2.9.9": false, "2.10.0": true, "2.14.1": true, "2.99.0": true, "3.0.0": false, "3.1.0": false} {
		if got := closed.covers(tag); got != want {
			t.Errorf("covers(%q) = %v, want %v", tag, got, want)
		}
	}
	open := embeddedList{capturedAt: "3.1.0", from: "3.1.0"}
	for tag, want := range map[string]bool{"3.0.9": false, "3.1.0": true, "9.0.0": true} {
		if got := open.covers(tag); got != want {
			t.Errorf("open covers(%q) = %v, want %v", tag, got, want)
		}
	}
	if got, want := closed.describes(), "2.10.0 and later, before 3.0.0"; got != want {
		t.Errorf("describes() = %q, want %q", got, want)
	}
	if got, want := open.describes(), "3.1.0 and later"; got != want {
		t.Errorf("open describes() = %q, want %q", got, want)
	}
	if got, want := closed.file(), "imagelibs/solace-pubsub-connector-ibmmq-2.13.0.list"; got != want {
		t.Errorf("file() = %q, want %q", got, want)
	}
}

// TestConnectorRelease covers reading the connector release off an image
// reference: the tag and the major it leads with, for the connector image
// only, and nothing for a reference that names no release.
func TestConnectorRelease(t *testing.T) {
	cases := []struct {
		name, ref, wantTag string
		wantMajor          int
		wantOK             bool
	}{
		{"a 2.x release", "solace/solace-pubsub-connector-ibmmq:2.14.1", "2.14.1", 2, true},
		{"a 3.x release", "solace/solace-pubsub-connector-ibmmq:3.1.0", "3.1.0", 3, true},
		{"a two-digit major", "solace/solace-pubsub-connector-ibmmq:10.2", "10.2", 10, true},
		{"a suffixed tag", "solace/solace-pubsub-connector-ibmmq:2.14.1-ubi", "2.14.1-ubi", 2, true},
		{"a private registry mirror", "registry.internal:5000/team/solace-pubsub-connector-ibmmq:3.1.0", "3.1.0", 3, true},
		{"latest names no release", "solace/solace-pubsub-connector-ibmmq:latest", "", 0, false},
		{"a digest pin", "solace/solace-pubsub-connector-ibmmq@sha256:abc123", "", 0, false},
		{"a different image", "solace/some-other-connector:2.14.1", "", 0, false},
		{"a major too large to be one", "solace/solace-pubsub-connector-ibmmq:99999999999999999999.0", "", 0, false},
		{"nothing declared", "", "", 0, false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			tag, major, ok := connectorRelease(c.ref)
			if tag != c.wantTag || major != c.wantMajor || ok != c.wantOK {
				t.Errorf("connectorRelease(%q) = (%q, %d, %v), want (%q, %d, %v)", c.ref, tag, major, ok, c.wantTag, c.wantMajor, c.wantOK)
			}
		})
	}
}

// TestImageNameTag covers splitting a full image reference into the name and
// tag builtinList compares separately: the name says whether this is the
// connector at all, the tag says which built-in list reaches it.
func TestImageNameTag(t *testing.T) {
	cases := []struct {
		name, ref, wantName, wantTag string
		wantOK                       bool
	}{
		{"docker hub reference", "solace/solace-pubsub-connector-ibmmq:2.14.1", "solace-pubsub-connector-ibmmq", "2.14.1", true},
		{"no namespace", "connector:1.0", "connector", "1.0", true},
		// The tag separator must be found in the LAST path element: a registry
		// host may carry a port, and splitting at the first colon would yield
		// "registry.internal" as the name.
		{"registry with a port", "registry.internal:5000/team/connector:2.14.1", "connector", "2.14.1", true},
		{"no tag at all", "solace/solace-pubsub-connector-ibmmq", "", "", false},
		{"digest pin names no release", "solace/connector@sha256:abc123", "", "", false},
		{"empty", "", "", "", false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			name, tag, ok := imageNameTag(c.ref)
			if ok != c.wantOK || name != c.wantName || tag != c.wantTag {
				t.Errorf("imageNameTag(%q) = (%q, %q, %v), want (%q, %q, %v)",
					c.ref, name, tag, ok, c.wantName, c.wantTag, c.wantOK)
			}
		})
	}
}

// TestBuiltinList covers which built-in list judges the deployed image, and
// when that is reported as not describing it. Each list covers a RANGE of
// releases rather than the single tag it was captured from, so matching that
// exact tag is not what makes it silent -- being inside the range is.
//
// Both directions matter. Silence on an image no list describes is what let a
// 2.13.0 list judge a 2.14.1 deployment unnoticed; a warning on one a list
// does describe is noise on every correct run, and noise is what operators
// learn to skip past.
func TestBuiltinList(t *testing.T) {
	first, newest := embeddedLists[0], embeddedLists[len(embeddedLists)-1]
	byTag := map[string]embeddedList{}
	for _, l := range embeddedLists {
		byTag[l.capturedAt] = l
	}
	silent := []struct {
		name, ref string
		want      embeddedList
	}{
		{"nothing declared", "", newest},
		{"the captured release itself", "solace/solace-pubsub-connector-ibmmq:2.13.0", first},
		{"a newer release the same list covers", "solace/solace-pubsub-connector-ibmmq:2.14.1", first},
		{"the floor itself is covered", "solace/solace-pubsub-connector-ibmmq:" + first.from, first},
		{"a private registry mirror of a covered release", "registry.internal:5000/team/solace-pubsub-connector-ibmmq:2.14.1", first},
		{"the first 3.x capture", "solace/solace-pubsub-connector-ibmmq:3.1.0", byTag["3.1.0"]},
		{"a 3.1 point release stays on the 3.1.0 list", "solace/solace-pubsub-connector-ibmmq:3.1.4", byTag["3.1.0"]},
		{"3.2.0 gets its own list", "solace/solace-pubsub-connector-ibmmq:3.2.0", byTag["3.2.0"]},
		{"past the newest capture stays on it until a release proves otherwise", "solace/solace-pubsub-connector-ibmmq:3.3.0", newest},
	}
	for _, c := range silent {
		t.Run("silent/"+c.name, func(t *testing.T) {
			got, note := builtinList(c.ref)
			if note != "" {
				t.Errorf("builtinList(%q) note = %q, want silence", c.ref, note)
			}
			if !reflect.DeepEqual(got, c.want) {
				t.Errorf("builtinList(%q) = %s, want %s", c.ref, got.name(), c.want.name())
			}
		})
	}

	warns := []struct {
		name, ref, mustName string
		want                embeddedList
	}{
		// Below the floor the classpath is unverified, so the omissions may
		// name jars that image does not ship.
		{"a release below the floor", "solace/solace-pubsub-connector-ibmmq:2.9.9", first.from, first},
		// The ceiling is what catches a line whose classpath moved: connector
		// 3.x is Spring Boot 4 and Jackson 3, which a 2.x capture cannot speak
		// for.
		{"the first release past a list's range", "solace/solace-pubsub-connector-ibmmq:" + first.before, first.name(), first},
		// No 3.0.x release was captured: it falls in the gap between the 2.x
		// list and the first 3.x one, and is judged against the 2.x list with
		// a warning rather than silently.
		{"a 3.0.x release no capture covers", "solace/solace-pubsub-connector-ibmmq:3.0.2", first.name(), first},
		{"a different image entirely", "solace/some-other-connector:2.14.1", embeddedListImage, newest},
		// A digest pin or a tag like latest names no release any list could
		// have been captured under, so it cannot be confirmed to be covered
		// either -- latest used to sort past every number and pass silently.
		{"a digest pin", "solace/solace-pubsub-connector-ibmmq@sha256:abc123", first.from, newest},
		{"no tag at all", "solace/solace-pubsub-connector-ibmmq", first.from, newest},
		{"latest", "solace/solace-pubsub-connector-ibmmq:latest", first.from, newest},
	}
	for _, c := range warns {
		t.Run("warns/"+c.name, func(t *testing.T) {
			got, note := builtinList(c.ref)
			if note == "" {
				t.Fatalf("builtinList(%q) was silent, want a warning", c.ref)
			}
			// The deployed reference and what it was judged against both have
			// to appear, or the operator cannot see what to fix.
			for _, want := range []string{c.ref, c.mustName, "--omit-lib-file"} {
				if !strings.Contains(note, want) {
					t.Errorf("warning %q should name %q", note, want)
				}
			}
			if !reflect.DeepEqual(got, c.want) {
				t.Errorf("builtinList(%q) = %s, want %s", c.ref, got.name(), c.want.name())
			}
		})
	}
}

// TestBuiltinListPicksTheNearestLine runs the pick against a fixed two-line
// table, so the choice between lines stays pinned however the shipped table
// grows: a release gets its own line, one in the gap between lines or past
// the last gets the nearest line captured at or before it, and one older than
// every capture gets the oldest.
func TestBuiltinListPicksTheNearestLine(t *testing.T) {
	orig := embeddedLists
	t.Cleanup(func() { embeddedLists = orig })
	twoX := embeddedList{capturedAt: "2.13.0", from: "2.10.0", before: "3.0.0"}
	threeX := embeddedList{capturedAt: "3.1.0", from: "3.1.0"}
	embeddedLists = []embeddedList{twoX, threeX}

	for _, c := range []struct {
		tag    string
		want   embeddedList
		silent bool
	}{
		{"2.14.1", twoX, true},
		{"3.1.0", threeX, true},
		{"4.0.0", threeX, true},
		{"3.0.5", twoX, false},
		{"2.9.0", twoX, false},
	} {
		got, note := builtinList("solace/solace-pubsub-connector-ibmmq:" + c.tag)
		if !reflect.DeepEqual(got, c.want) || (note == "") != c.silent {
			t.Errorf("builtinList(%s) = %s with note %q, want %s (silent %v)", c.tag, got.name(), note, c.want.name(), c.silent)
		}
	}
	if got, _ := builtinList(""); !reflect.DeepEqual(got, threeX) {
		t.Errorf("nothing declared picked %s, want the newest line", got.name())
	}
}
