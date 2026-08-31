package libs

import (
	"os"
	"path/filepath"
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
	loaded, err := loadImageLibs(path)
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
	loaded, err := loadImageLibs(path)
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
	loaded, err := loadImageLibs(path)
	if err != nil {
		t.Fatalf("loadImageLibs: %v", err)
	}
	libs := loaded.Libs
	if libs["bcprov-jdk18on"] != "1.84" {
		t.Errorf("libs[bcprov-jdk18on] = %q, want 1.84", libs["bcprov-jdk18on"])
	}
}

func TestLoadImageLibsBadPathIsError(t *testing.T) {
	_, err := loadImageLibs(filepath.Join(t.TempDir(), "does-not-exist.list"))
	if err == nil {
		t.Fatal("want error for a nonexistent path")
	}
}

func TestLoadImageLibsEmbeddedDefault(t *testing.T) {
	loaded, err := loadImageLibs("")
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

// TestEmbeddedOmitListFullyParses holds the shipped list to the standard the
// code expects of it: every line either is not a jar reference at all, or
// splits and carries a version the comparator can order.
//
// This is what the 15 warnings on every `download jar mq` run were telling us,
// unread. A future capture from a newer image that reintroduces an unorderable
// shape now fails the build instead of printing noise no one acts on.
func TestEmbeddedOmitListFullyParses(t *testing.T) {
	loaded, err := loadImageLibs("")
	if err != nil {
		t.Fatalf("loadImageLibs: %v", err)
	}
	if len(loaded.Rejected) != 0 {
		t.Errorf("the embedded omit list has %d unparseable entries, want none: %v",
			len(loaded.Rejected), loaded.Rejected)
	}
	// Sanity: the list did actually load, so a future bug that empties it
	// cannot make the assertion above pass vacuously.
	if len(loaded.Libs) < 50 {
		t.Errorf("embedded list parsed only %d entries, expected the full image classpath", len(loaded.Libs))
	}
	for _, want := range []string{"netty-common", "hibernate-validator", "jakarta.jms-api"} {
		if _, ok := loaded.Libs[want]; !ok {
			t.Errorf("embedded list has no entry for %q", want)
		}
	}
}

// TestImageNameTag covers splitting a full image reference into the name and
// tag imageMismatchNote compares separately: the name says whether this is the
// connector at all, the tag says whether the embedded list reaches it.
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

// TestImageMismatchNote covers when the embedded jar list is reported as not
// describing the deployed image. The list covers a RANGE of releases rather
// than the single tag it was captured from, so matching that exact tag is not
// what makes it silent -- being at or above EmbeddedListMinVersion is.
//
// Both directions matter. Silence on an image the list does not describe is
// what let a 2.13.0 list judge a 2.14.1 deployment unnoticed; a warning on one
// it does describe is noise on every correct run, and noise is what operators
// learn to skip past.
func TestImageMismatchNote(t *testing.T) {
	silent := []struct{ name, ref string }{
		{"nothing declared", ""},
		{"the captured release itself", "solace/solace-pubsub-connector-ibmmq:2.13.0"},
		{"a newer release the same list covers", "solace/solace-pubsub-connector-ibmmq:2.14.1"},
		{"the floor itself is covered", "solace/solace-pubsub-connector-ibmmq:" + EmbeddedListMinVersion},
		// No ceiling: the floor is the only bound, so a release beyond every
		// capture stays silent. Raising the floor is how a release that
		// diverges gets caught -- see the list header.
		{"a release past every capture", "solace/solace-pubsub-connector-ibmmq:3.0.0"},
		{"a private registry mirror of a covered release", "registry.internal:5000/team/solace-pubsub-connector-ibmmq:2.14.1"},
	}
	for _, c := range silent {
		t.Run("silent/"+c.name, func(t *testing.T) {
			if got := imageMismatchNote(c.ref); got != "" {
				t.Errorf("imageMismatchNote(%q) = %q, want silence", c.ref, got)
			}
		})
	}

	warns := []struct{ name, ref, mustName string }{
		// Below the floor the classpath is unverified, so the omissions may
		// name jars that image does not ship.
		{"a release below the floor", "solace/solace-pubsub-connector-ibmmq:2.9.9", EmbeddedListMinVersion},
		{"a different image entirely", "solace/some-other-connector:2.14.1", embeddedListImage},
		// A digest pin names no release any list could have been captured
		// under, so it cannot be confirmed to be covered either.
		{"a digest pin", "solace/solace-pubsub-connector-ibmmq@sha256:abc123", EmbeddedListMinVersion},
		{"no tag at all", "solace/solace-pubsub-connector-ibmmq", EmbeddedListMinVersion},
	}
	for _, c := range warns {
		t.Run("warns/"+c.name, func(t *testing.T) {
			got := imageMismatchNote(c.ref)
			if got == "" {
				t.Fatalf("imageMismatchNote(%q) was silent, want a warning", c.ref)
			}
			// The deployed reference and what it was judged against both have
			// to appear, or the operator cannot see what to fix.
			for _, want := range []string{c.ref, c.mustName} {
				if !strings.Contains(got, want) {
					t.Errorf("warning %q should name %q", got, want)
				}
			}
			if !strings.Contains(got, "--omit-lib-file") {
				t.Errorf("warning %q should name the remedy", got)
			}
		})
	}
}
