package libs

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
)

// mavenFakeResp is one canned response for a single fixture URL.
type mavenFakeResp struct {
	status int
	body   string
	err    error
}

// mavenFakeDoer is the test double for Doer: it records every URL requested in
// calls and answers from responses, keyed by the exact URL string. A URL
// with no entry answers 404, which is what a real Maven Central miss looks
// like -- this lets an "unreachable POM" test simply omit a fixture rather
// than needing an explicit not-found entry.
type mavenFakeDoer struct {
	responses map[string]mavenFakeResp
	calls     []string
}

func (f *mavenFakeDoer) Do(req *http.Request) (*http.Response, error) {
	u := req.URL.String()
	f.calls = append(f.calls, u)
	r, ok := f.responses[u]
	if !ok {
		return &http.Response{StatusCode: http.StatusNotFound, Status: "404 Not Found", Body: io.NopCloser(strings.NewReader(""))}, nil
	}
	if r.err != nil {
		return nil, r.err
	}
	status := r.status
	if status == 0 {
		status = http.StatusOK
	}
	return &http.Response{StatusCode: status, Status: http.StatusText(status), Body: io.NopCloser(strings.NewReader(r.body))}, nil
}

func metaXML(release string, versions ...string) string {
	var sb strings.Builder
	sb.WriteString("<metadata><versioning><release>" + release + "</release><versions>")
	for _, v := range versions {
		sb.WriteString("<version>" + v + "</version>")
	}
	sb.WriteString("</versions></versioning></metadata>")
	return sb.String()
}

func dep(group, artifact, version, scope, optional, typ string) string {
	s := "<dependency><groupId>" + group + "</groupId><artifactId>" + artifact + "</artifactId>"
	if version != "" {
		s += "<version>" + version + "</version>"
	}
	if scope != "" {
		s += "<scope>" + scope + "</scope>"
	}
	if optional != "" {
		s += "<optional>" + optional + "</optional>"
	}
	if typ != "" {
		s += "<type>" + typ + "</type>"
	}
	s += "</dependency>"
	return s
}

func pomXMLBody(deps ...string) string {
	return "<project><dependencies>" + strings.Join(deps, "") + "</dependencies></project>"
}

// --- version comparator ---

func TestCompareVersions(t *testing.T) {
	cases := []struct {
		name string
		a, b string
		want int
	}{
		{"patch numeric ordering", "9.4.10.0", "9.4.9.0", 1},
		{"avoids lexical misorder", "9.1.0.15", "9.1.0.9", 1},
		{"major numeric ordering", "10.0.0.0", "9.4.0.0", 1},
		{"prerelease suffix ranks lower", "9.4.0.0", "9.4.0.0-rc1", 1},
		// Maven treats Final/GA/RELEASE as aliases of the empty qualifier, so
		// they name the SAME version rather than a later one. The whole image
		// classpath (netty, hibernate, jboss) is spelled this way.
		{"Final means the release itself", "4.1.135.Final", "4.1.135", 0},
		{"RELEASE means the release itself", "5.3.31.RELEASE", "5.3.31", 0},
		{"GA means the release itself", "1.0.GA", "1.0", 0},
		{"release qualifiers are interchangeable", "1.0.Final", "1.0.GA", 0},
		// Aligned, not trailing: left to strings.Compare "Final" beats "1"
		// lexicographically, which would make 1.0.Final look newer than 1.0.1.
		{"release qualifier sorts below a number", "1.0.Final", "1.0.1", -1},
		{"release outranks a prerelease", "1.0.Final", "1.0-rc1", 1},
		{"service pack outranks the release", "1.0-sp1", "1.0", 1},
		{"service pack outranks the same release spelled Final", "1.0-sp1", "1.0.Final", 1},
		{"date-like numeric ordering", "20260814", "20251224", 1},
		{"missing trailing segments rank lower", "9.4", "9.4.0.0", -1},
		{"equal inputs", "9.4.0.0", "9.4.0.0", 0},
		{"empty strings", "", "", 0},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := compareVersions(c.a, c.b)
			if sign(got) != sign(c.want) {
				t.Errorf("compareVersions(%q, %q) = %d, want sign %d", c.a, c.b, got, c.want)
			}
			// Antisymmetry: swapping arguments must flip (or preserve, for 0) the sign.
			if rev := compareVersions(c.b, c.a); sign(rev) != -sign(c.want) {
				t.Errorf("compareVersions(%q, %q) = %d, want sign %d", c.b, c.a, rev, -c.want)
			}
		})
	}
}

func sign(n int) int {
	switch {
	case n < 0:
		return -1
	case n > 0:
		return 1
	default:
		return 0
	}
}

func TestIsPreRelease(t *testing.T) {
	cases := []struct {
		v    string
		want bool
	}{
		{"3.0-rc5", true}, // the verified jackson-annotations <release> case
		{"1.0-SNAPSHOT", true},
		{"1.0.snapshot", true},
		{"5.0.M1", true},
		{"5.0.m2", true},
		{"1.0-alpha1", true},
		{"1.0.beta2", true},
		{"1.0_cr1", true},
		{"1.0-pr3", true},
		{"8-ea", true},
		{"1.0-preview1", true},
		// The one that matters most: if a release qualifier ever counted as a
		// pre-release, every netty and hibernate release would be skipped when
		// picking a latest stable version.
		{"4.1.135.Final", false},
		{"5.3.31.RELEASE", false},
		{"1.0.GA", false},
		{"1.0-sp1", false},
		{"10.0.0.0", false},
		{"20260814", false},
		{"1.85.2", false},
		{"", false},
	}
	for _, c := range cases {
		if got := isPreRelease(c.v); got != c.want {
			t.Errorf("isPreRelease(%q) = %v, want %v", c.v, got, c.want)
		}
	}
}

// --- coordinate charset validation ---

func TestValidateCoordPart(t *testing.T) {
	cases := []struct {
		s       string
		wantErr bool
	}{
		{"com.ibm.mq", false},
		{"com.ibm.mq.jakarta.client", false},
		{"10.0.0.0", false},
		{"3.0-rc5", false},
		{"", true},
		{"../../etc", true},
		{"com..evil", true},
		{"a/b", true},
		{"a\\b", true},
		{"a\x00b", true},
	}
	for _, c := range cases {
		err := validateCoordPart(c.s)
		if (err != nil) != c.wantErr {
			t.Errorf("validateCoordPart(%q) err = %v, wantErr %v", c.s, err, c.wantErr)
		}
	}
}

// --- latestStable ---

func TestLatestStablePrefersRelease(t *testing.T) {
	c := Coord{Group: "com.ibm.mq", Artifact: "com.ibm.mq.allclient"}
	d := &mavenFakeDoer{responses: map[string]mavenFakeResp{
		metadataURL(c): {body: metaXML("10.0.0.0", "9.1.0.9", "9.1.0.15", "9.1.0.18", "10.0.0.0")},
	}}
	got, err := latestStable(d, c)
	if err != nil {
		t.Fatal(err)
	}
	if got != "10.0.0.0" {
		t.Errorf("latestStable = %q, want 10.0.0.0", got)
	}
}

func TestLatestStableSkipsPreReleaseCandidateInRelease(t *testing.T) {
	// The verified jackson-annotations case: <release> names a release
	// candidate, so the highest surviving <version> must be used instead.
	c := Coord{Group: "com.fasterxml.jackson.core", Artifact: "jackson-annotations"}
	d := &mavenFakeDoer{responses: map[string]mavenFakeResp{
		metadataURL(c): {body: metaXML("3.0-rc5", "2.19.0", "3.0.0", "3.0-rc5")},
	}}
	got, err := latestStable(d, c)
	if err != nil {
		t.Fatal(err)
	}
	if got != "3.0.0" {
		t.Errorf("latestStable = %q, want 3.0.0", got)
	}
}

func TestLatestStableAllPreReleaseVersionsIsError(t *testing.T) {
	c := Coord{Group: "g", Artifact: "a"}
	d := &mavenFakeDoer{responses: map[string]mavenFakeResp{
		metadataURL(c): {body: metaXML("1.0-rc1", "1.0-rc1", "1.0-beta1")},
	}}
	if _, err := latestStable(d, c); err == nil {
		t.Error("latestStable = nil error, want an error when every version is a pre-release")
	}
}

func TestLatestStableUnreachableMetadataIsError(t *testing.T) {
	c := Coord{Group: "g", Artifact: "a"}
	d := &mavenFakeDoer{responses: map[string]mavenFakeResp{}}
	if _, err := latestStable(d, c); err == nil {
		t.Error("latestStable = nil error, want an error when metadata is unreachable")
	}
}

// --- resolveClosure: the two verified mq closures ---

func mqFixtures(seed Coord, seedVersion string, jmsDep, jsonVersion string) map[string]mavenFakeResp {
	bcprov := Coord{"org.bouncycastle", "bcprov-jdk18on"}
	bcpkix := Coord{"org.bouncycastle", "bcpkix-jdk18on"}
	bcutil := Coord{"org.bouncycastle", "bcutil-jdk18on"}
	orgJSON := Coord{"org.json", "json"}

	return map[string]mavenFakeResp{
		metadataURL(seed): {body: metaXML(seedVersion, seedVersion)},
		pomURL(seed, seedVersion): {body: pomXMLBody(
			dep("org.bouncycastle", "bcprov-jdk18on", "1.84", "", "", ""),
			dep("org.bouncycastle", "bcpkix-jdk18on", "1.84", "", "", ""),
			dep("org.bouncycastle", "bcutil-jdk18on", "1.84", "", "", ""),
			jmsDep,
			dep("org.json", "json", jsonVersion, "", "", ""),
		)},
		pomURL(bcprov, "1.84"): {body: pomXMLBody()},
		pomURL(bcpkix, "1.84"): {body: pomXMLBody(
			dep("org.bouncycastle", "bcutil-jdk18on", "1.84", "", "", ""),
		)},
		pomURL(bcutil, "1.84"): {body: pomXMLBody(
			dep("org.bouncycastle", "bcprov-jdk18on", "1.84", "", "", ""),
		)},
		pomURL(orgJSON, jsonVersion): {body: pomXMLBody(
			dep("junit", "junit", "4.13.2", "test", "", ""),
		)},
	}
}

func TestResolveClosureMQJakarta(t *testing.T) {
	seed := Coord{Group: "com.ibm.mq", Artifact: "com.ibm.mq.jakarta.client"}
	jakartaJMS := Coord{"jakarta.jms", "jakarta.jms-api"}
	fixtures := mqFixtures(seed, "10.0.0.0", dep("jakarta.jms", "jakarta.jms-api", "3.0.0", "", "", ""), "20251224")
	fixtures[pomURL(jakartaJMS, "3.0.0")] = mavenFakeResp{body: pomXMLBody()}

	d := &mavenFakeDoer{responses: fixtures}
	got, err := resolveClosure(d, seed)
	if err != nil {
		t.Fatal(err)
	}

	want := []artifact{
		{Coord: seed, Version: "10.0.0.0"},
		{Coord: Coord{"org.bouncycastle", "bcprov-jdk18on"}, Version: "1.84"},
		{Coord: Coord{"org.bouncycastle", "bcpkix-jdk18on"}, Version: "1.84"},
		{Coord: Coord{"org.bouncycastle", "bcutil-jdk18on"}, Version: "1.84"},
		{Coord: jakartaJMS, Version: "3.0.0"},
		{Coord: Coord{"org.json", "json"}, Version: "20251224"},
	}
	assertClosure(t, got, want)
}

// TestResolveClosureSyslogResolvesParentProperties reproduces the verified
// logstash-logback-encoder:9.0 -> jackson-databind:3.0.1 case, where
// jackson-databind's own POM declares its jackson-core/jackson-annotations
// dependency versions only as ${jackson.version.core} /
// ${jackson.version.annotations}, defined in its parent jackson-base's
// <properties>. The parent chain must resolve both, not fall back.
func TestResolveClosureSyslogResolvesParentProperties(t *testing.T) {
	seed := Coord{Group: "net.logstash.logback", Artifact: "logstash-logback-encoder"}
	jacksonDatabind := Coord{"tools.jackson.core", "jackson-databind"}
	jacksonBase := Coord{"com.fasterxml.jackson", "jackson-base"}
	jacksonAnnotations := Coord{"com.fasterxml.jackson.core", "jackson-annotations"}
	jacksonCore := Coord{"tools.jackson.core", "jackson-core"}

	d := &mavenFakeDoer{responses: map[string]mavenFakeResp{
		metadataURL(seed): {body: metaXML("9.0", "9.0")},
		pomURL(seed, "9.0"): {body: pomXMLBody(
			dep("ch.qos.logback", "logback-classic", "", "compile", "true", ""),
			dep("ch.qos.logback", "logback-core", "", "provided", "", ""),
			dep("ch.qos.logback", "logback-access-common", "", "compile", "true", ""),
			dep("tools.jackson.core", "jackson-databind", "3.0.1", "compile", "", ""),
			dep("com.fasterxml.jackson.dataformat", "jackson-dataformat-cbor", "", "", "true", ""),
			dep("com.fasterxml.jackson.dataformat", "jackson-dataformat-smile", "", "", "true", ""),
			dep("com.fasterxml.jackson.dataformat", "jackson-dataformat-yaml", "", "", "true", ""),
			dep("com.fasterxml.uuid", "java-uuid-generator", "", "", "true", ""),
		)},
		pomURL(jacksonDatabind, "3.0.1"): {body: `<project>
			<parent><groupId>com.fasterxml.jackson</groupId><artifactId>jackson-base</artifactId><version>3.0.1</version></parent>
			<dependencies>
				<dependency><groupId>com.fasterxml.jackson.core</groupId><artifactId>jackson-annotations</artifactId><version>${jackson.version.annotations}</version></dependency>
				<dependency><groupId>tools.jackson.core</groupId><artifactId>jackson-core</artifactId><version>${jackson.version.core}</version></dependency>
			</dependencies>
		</project>`},
		pomURL(jacksonBase, "3.0.1"): {body: `<project>
			<properties>
				<jackson.version.annotations>3.0.1</jackson.version.annotations>
				<jackson.version.core>3.0.1</jackson.version.core>
			</properties>
		</project>`},
		pomURL(jacksonAnnotations, "3.0.1"): {body: pomXMLBody()},
		pomURL(jacksonCore, "3.0.1"):        {body: pomXMLBody()},
	}}

	got, err := resolveClosure(d, seed)
	if err != nil {
		t.Fatal(err)
	}
	want := []artifact{
		{Coord: seed, Version: "9.0"},
		{Coord: jacksonDatabind, Version: "3.0.1"},
		{Coord: jacksonAnnotations, Version: "3.0.1"},
		{Coord: jacksonCore, Version: "3.0.1"},
	}
	assertClosure(t, got, want)
	for _, a := range got {
		if a.Fallback {
			t.Errorf("artifact %s:%s unexpectedly marked Fallback (parent chain should have resolved it)", a.Group, a.Artifact)
		}
	}
}

func TestResolveClosureAppliesScopeOptionalTypeFilter(t *testing.T) {
	seed := Coord{Group: "test.group", Artifact: "filtertest"}
	compileLeaf := Coord{"test.group", "compile-leaf"}
	runtimeLeaf := Coord{"test.group", "runtime-leaf"}

	d := &mavenFakeDoer{responses: map[string]mavenFakeResp{
		metadataURL(seed): {body: metaXML("1.0", "1.0")},
		pomURL(seed, "1.0"): {body: pomXMLBody(
			dep("test.group", "test-scope-leaf", "1.0", "test", "", ""),
			dep("test.group", "provided-scope-leaf", "1.0", "provided", "", ""),
			dep("test.group", "system-scope-leaf", "1.0", "system", "", ""),
			dep("test.group", "import-scope-leaf", "1.0", "import", "", ""),
			dep("test.group", "optional-leaf", "1.0", "compile", "true", ""),
			dep("test.group", "bom", "1.0", "compile", "", "pom"),
			dep("test.group", "compile-leaf", "1.0", "compile", "", ""),
			dep("test.group", "runtime-leaf", "1.0", "runtime", "", ""),
		)},
		pomURL(compileLeaf, "1.0"): {body: pomXMLBody()},
		pomURL(runtimeLeaf, "1.0"): {body: pomXMLBody()},
	}}

	got, err := resolveClosure(d, seed)
	if err != nil {
		t.Fatal(err)
	}
	want := []artifact{
		{Coord: seed, Version: "1.0"},
		{Coord: compileLeaf, Version: "1.0"},
		{Coord: runtimeLeaf, Version: "1.0"},
	}
	assertClosure(t, got, want)
}

func TestResolveClosureDependencyVersionFromDependencyManagement(t *testing.T) {
	seed := Coord{Group: "test.group", Artifact: "dmtest"}
	parent := Coord{"test.group", "dmparent"}
	leaf := Coord{"test.group", "dmleaf"}

	d := &mavenFakeDoer{responses: map[string]mavenFakeResp{
		metadataURL(seed): {body: metaXML("1.0", "1.0")},
		pomURL(seed, "1.0"): {body: `<project>
			<parent><groupId>test.group</groupId><artifactId>dmparent</artifactId><version>1.0</version></parent>
			<dependencies>
				<dependency><groupId>test.group</groupId><artifactId>dmleaf</artifactId></dependency>
			</dependencies>
		</project>`},
		pomURL(parent, "1.0"): {body: `<project>
			<dependencyManagement>
				<dependencies>
					<dependency><groupId>test.group</groupId><artifactId>dmleaf</artifactId><version>5.5</version></dependency>
				</dependencies>
			</dependencyManagement>
		</project>`},
		pomURL(leaf, "5.5"): {body: pomXMLBody()},
	}}

	got, err := resolveClosure(d, seed)
	if err != nil {
		t.Fatal(err)
	}
	want := []artifact{
		{Coord: seed, Version: "1.0"},
		{Coord: leaf, Version: "5.5"},
	}
	assertClosure(t, got, want)
}

func TestResolveClosurePropertyDefinedTwoParentsUp(t *testing.T) {
	seed := Coord{Group: "test.group", Artifact: "twoparentchild"}
	parent1 := Coord{"test.group", "parent1"}
	parent2 := Coord{"test.group", "parent2"}
	leaf := Coord{"test.group", "twoparentleaf"}

	d := &mavenFakeDoer{responses: map[string]mavenFakeResp{
		metadataURL(seed): {body: metaXML("1.0", "1.0")},
		pomURL(seed, "1.0"): {body: `<project>
			<parent><groupId>test.group</groupId><artifactId>parent1</artifactId><version>1.0</version></parent>
			<dependencies>
				<dependency><groupId>test.group</groupId><artifactId>twoparentleaf</artifactId><version>${some.prop}</version></dependency>
			</dependencies>
		</project>`},
		pomURL(parent1, "1.0"): {body: `<project>
			<parent><groupId>test.group</groupId><artifactId>parent2</artifactId><version>1.0</version></parent>
		</project>`},
		pomURL(parent2, "1.0"): {body: `<project>
			<properties><some.prop>2.5</some.prop></properties>
		</project>`},
		pomURL(leaf, "2.5"): {body: pomXMLBody()},
	}}

	got, err := resolveClosure(d, seed)
	if err != nil {
		t.Fatal(err)
	}
	want := []artifact{
		{Coord: seed, Version: "1.0"},
		{Coord: leaf, Version: "2.5"},
	}
	assertClosure(t, got, want)
}

func TestResolveClosureUndefinedPropertyFallsBackToLatestStable(t *testing.T) {
	seed := Coord{Group: "test.group", Artifact: "undefinedpropchild"}
	leaf := Coord{"test.group", "undefinedpropleaf"}

	d := &mavenFakeDoer{responses: map[string]mavenFakeResp{
		metadataURL(seed): {body: metaXML("1.0", "1.0")},
		pomURL(seed, "1.0"): {body: `<project>
			<dependencies>
				<dependency><groupId>test.group</groupId><artifactId>undefinedpropleaf</artifactId><version>${nowhere.defined}</version></dependency>
			</dependencies>
		</project>`},
		metadataURL(leaf):   {body: metaXML("9.9", "9.9")},
		pomURL(leaf, "9.9"): {body: pomXMLBody()},
	}}

	got, err := resolveClosure(d, seed)
	if err != nil {
		t.Fatal(err)
	}
	want := []artifact{
		{Coord: seed, Version: "1.0"},
		{Coord: leaf, Version: "9.9", Fallback: true},
	}
	assertClosure(t, got, want)
}

func TestResolveClosureDependencyCycleTerminates(t *testing.T) {
	seed := Coord{Group: "test.group", Artifact: "cycleseed"}
	x := Coord{"test.group", "cyclex"}
	y := Coord{"test.group", "cycley"}

	d := &mavenFakeDoer{responses: map[string]mavenFakeResp{
		metadataURL(seed): {body: metaXML("1.0", "1.0")},
		pomURL(seed, "1.0"): {body: pomXMLBody(
			dep("test.group", "cyclex", "1.0", "", "", ""),
		)},
		pomURL(x, "1.0"): {body: pomXMLBody(
			dep("test.group", "cycley", "1.0", "", "", ""),
		)},
		pomURL(y, "1.0"): {body: pomXMLBody(
			dep("test.group", "cyclex", "1.0", "", "", ""), // back-edge to x
		)},
	}}

	got, err := resolveClosure(d, seed)
	if err != nil {
		t.Fatal(err)
	}
	want := []artifact{
		{Coord: seed, Version: "1.0"},
		{Coord: x, Version: "1.0"},
		{Coord: y, Version: "1.0"},
	}
	assertClosure(t, got, want)
}

func TestResolveClosureArtifactCountCap(t *testing.T) {
	seed := Coord{Group: "test.group", Artifact: "capseed"}
	var deps []string
	for i := 0; i < maxArtifacts+10; i++ {
		deps = append(deps, dep("test.group", fmt.Sprintf("capleaf%d", i), "1.0", "", "", ""))
	}

	d := &mavenFakeDoer{responses: map[string]mavenFakeResp{
		metadataURL(seed):   {body: metaXML("1.0", "1.0")},
		pomURL(seed, "1.0"): {body: pomXMLBody(deps...)},
		// Deliberately no fixtures for the capleaf* dependency POMs: each is a
		// leaf as far as this test cares, and an unreachable dependency POM
		// is not an error (see resolveClosure's doc comment).
	}}

	got, err := resolveClosure(d, seed)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != maxArtifacts {
		t.Errorf("len(got) = %d, want exactly maxArtifacts (%d)", len(got), maxArtifacts)
	}
}

func TestResolveClosureHostileGroupIDIsDropped(t *testing.T) {
	seed := Coord{Group: "test.group", Artifact: "hostileseed"}
	goodLeaf := Coord{"test.group", "goodleaf"}

	d := &mavenFakeDoer{responses: map[string]mavenFakeResp{
		metadataURL(seed): {body: metaXML("1.0", "1.0")},
		pomURL(seed, "1.0"): {body: pomXMLBody(
			dep("../../etc", "passwd", "1.0", "", "", ""),
			dep("test.group", "goodleaf", "1.0", "", "", ""),
		)},
	}}

	got, err := resolveClosure(d, seed)
	if err != nil {
		t.Fatal(err)
	}
	want := []artifact{
		{Coord: seed, Version: "1.0"},
		{Coord: goodLeaf, Version: "1.0"},
	}
	assertClosure(t, got, want)

	for _, u := range d.calls {
		if strings.Contains(u, "..") {
			t.Errorf("a hostile coordinate reached the HTTP layer: %s", u)
		}
	}
}

func TestResolveClosureUnreachableSeedMetadataIsError(t *testing.T) {
	seed := Coord{Group: "test.group", Artifact: "unreachableseed"}
	d := &mavenFakeDoer{responses: map[string]mavenFakeResp{}}
	if _, err := resolveClosure(d, seed); err == nil {
		t.Error("resolveClosure = nil error, want an error when the seed metadata is unreachable")
	}
}

// TestResolveClosureUnreachableDependencyPomKeepsArtifact documents the
// chosen behavior for an unreachable non-seed dependency POM: resolveClosure
// has no per-artifact failure channel, so the artifact stays in the closure
// at the version its declaring POM already gave it, and simply is not
// expanded further. A real problem with the artifact's jar is left for
// libs.Download to report as a per-artifact Failure when it actually fetches
// the jar.
func TestResolveClosureUnreachableDependencyPomKeepsArtifact(t *testing.T) {
	seed := Coord{Group: "test.group", Artifact: "unreachabledepseed"}
	leaf := Coord{"test.group", "unreachableleaf"}

	d := &mavenFakeDoer{responses: map[string]mavenFakeResp{
		metadataURL(seed): {body: metaXML("1.0", "1.0")},
		pomURL(seed, "1.0"): {body: pomXMLBody(
			dep("test.group", "unreachableleaf", "9.9", "", "", ""),
		)},
		// No fixture for pomURL(leaf, "9.9"): its POM 404s.
	}}

	got, err := resolveClosure(d, seed)
	if err != nil {
		t.Fatal(err)
	}
	want := []artifact{
		{Coord: seed, Version: "1.0"},
		{Coord: leaf, Version: "9.9"},
	}
	assertClosure(t, got, want)
}

// TestResolveClosureDependencyVersionUsesProjectVersionProperty exercises
// resolveProperty's ${project.version} case: a dependency version that
// refers back to the declaring POM's own version.
func TestResolveClosureDependencyVersionUsesProjectVersionProperty(t *testing.T) {
	seed := Coord{Group: "test.group", Artifact: "projectversionchild"}
	leaf := Coord{"test.group", "projectversionleaf"}

	d := &mavenFakeDoer{responses: map[string]mavenFakeResp{
		metadataURL(seed): {body: metaXML("2.3", "2.3")},
		pomURL(seed, "2.3"): {body: pomXMLBody(
			dep("test.group", "projectversionleaf", "${project.version}", "", "", ""),
		)},
		pomURL(leaf, "2.3"): {body: pomXMLBody()},
	}}

	got, err := resolveClosure(d, seed)
	if err != nil {
		t.Fatal(err)
	}
	want := []artifact{
		{Coord: seed, Version: "2.3"},
		{Coord: leaf, Version: "2.3"},
	}
	assertClosure(t, got, want)
}

// TestResolveClosureDependencyVersionUsesProjectGroupIdProperty exercises
// resolveProperty's ${project.groupId} case. No real POM would declare a
// dependency version this way; the fixture only needs to prove the
// substitution branch fires and its result reaches the closure unchanged.
func TestResolveClosureDependencyVersionUsesProjectGroupIdProperty(t *testing.T) {
	seed := Coord{Group: "test.group", Artifact: "projectgroupidchild"}
	leaf := Coord{"test.group", "projectgroupidleaf"}

	d := &mavenFakeDoer{responses: map[string]mavenFakeResp{
		metadataURL(seed): {body: metaXML("1.0", "1.0")},
		pomURL(seed, "1.0"): {body: pomXMLBody(
			dep("test.group", "projectgroupidleaf", "${project.groupId}", "", "", ""),
		)},
		pomURL(leaf, "test.group"): {body: pomXMLBody()},
	}}

	got, err := resolveClosure(d, seed)
	if err != nil {
		t.Fatal(err)
	}
	want := []artifact{
		{Coord: seed, Version: "1.0"},
		{Coord: leaf, Version: "test.group"},
	}
	assertClosure(t, got, want)
}

// TestResolveClosureDependencyUnresolvableVersionAndUnreachableMetadataIsDropped
// exercises resolveDependencyVersion's ok=false path: a dependency with no
// version anywhere in the chain, whose own maven-metadata.xml is also
// unreachable, is dropped like any other malformed item rather than
// surfacing an error.
func TestResolveClosureDependencyUnresolvableVersionAndUnreachableMetadataIsDropped(t *testing.T) {
	seed := Coord{Group: "test.group", Artifact: "unresolvablechild"}

	d := &mavenFakeDoer{responses: map[string]mavenFakeResp{
		metadataURL(seed): {body: metaXML("1.0", "1.0")},
		pomURL(seed, "1.0"): {body: pomXMLBody(
			dep("test.group", "unresolvableleaf", "", "", "", ""),
		)},
		// Deliberately no fixture for metadataURL of the leaf: its own
		// latest-stable fallback lookup 404s too, so ok is false.
	}}

	got, err := resolveClosure(d, seed)
	if err != nil {
		t.Fatal(err)
	}
	want := []artifact{
		{Coord: seed, Version: "1.0"},
	}
	assertClosure(t, got, want)
}

// --- resolveClosureAt: pinned seed version ---

func TestResolveClosureAtPinnedVersionNeverFetchesMetadata(t *testing.T) {
	seed := Coord{Group: "com.ibm.mq", Artifact: "com.ibm.mq.jakarta.client"}
	d := &mavenFakeDoer{responses: map[string]mavenFakeResp{
		pomURL(seed, "9.4.2.0"): {body: pomXMLBody()},
		// Deliberately no fixture for metadataURL(seed): a pinned version
		// must never consult maven-metadata.xml.
	}}

	got, err := resolveClosureAt(d, seed, "9.4.2.0")
	if err != nil {
		t.Fatal(err)
	}
	want := []artifact{{Coord: seed, Version: "9.4.2.0"}}
	assertClosure(t, got, want)
	if got[0].Fallback {
		t.Error("pinned seed unexpectedly marked Fallback")
	}
	for _, u := range d.calls {
		if strings.Contains(u, "maven-metadata.xml") {
			t.Errorf("pinned version fetched maven-metadata.xml: %s", u)
		}
	}
}

// TestResolveClosureAtPinnedVersionResolvesDependencyClosure pins the seed to
// the verified com.ibm.mq.jakarta.client:9.4.2.0 release and confirms its
// dependency (jakarta.jms-api:3.0.0, per that release's own POM) still
// resolves through the normal parent/dependency chain.
func TestResolveClosureAtPinnedVersionResolvesDependencyClosure(t *testing.T) {
	seed := Coord{Group: "com.ibm.mq", Artifact: "com.ibm.mq.jakarta.client"}
	jakartaJMS := Coord{"jakarta.jms", "jakarta.jms-api"}
	d := &mavenFakeDoer{responses: map[string]mavenFakeResp{
		pomURL(seed, "9.4.2.0"): {body: pomXMLBody(
			dep("jakarta.jms", "jakarta.jms-api", "3.0.0", "", "", ""),
		)},
		pomURL(jakartaJMS, "3.0.0"): {body: pomXMLBody()},
	}}

	got, err := resolveClosureAt(d, seed, "9.4.2.0")
	if err != nil {
		t.Fatal(err)
	}
	want := []artifact{
		{Coord: seed, Version: "9.4.2.0"},
		{Coord: jakartaJMS, Version: "3.0.0"},
	}
	assertClosure(t, got, want)
}

func TestResolveClosureAtPinnedVersionNotFoundIsActionableError(t *testing.T) {
	seed := Coord{Group: "com.ibm.mq", Artifact: "com.ibm.mq.jakarta.client"}
	d := &mavenFakeDoer{responses: map[string]mavenFakeResp{}}

	_, err := resolveClosureAt(d, seed, "0.0.0-does-not-exist")
	if err == nil {
		t.Fatal("resolveClosureAt = nil error, want an error for a nonexistent pinned version")
	}
	if !strings.Contains(err.Error(), "0.0.0-does-not-exist") || !strings.Contains(err.Error(), seed.Artifact) {
		t.Errorf("err = %v, want it to name the pinned version and the artifact", err)
	}
}

func TestResolveClosureAtPinnedVersionInvalidCharsetIsError(t *testing.T) {
	seed := Coord{Group: "com.ibm.mq", Artifact: "com.ibm.mq.jakarta.client"}
	d := &mavenFakeDoer{responses: map[string]mavenFakeResp{}}

	_, err := resolveClosureAt(d, seed, "../../etc")
	if err == nil {
		t.Fatal("resolveClosureAt = nil error, want an error for a pinned version outside the allowed charset")
	}
	for _, u := range d.calls {
		if strings.Contains(u, "..") {
			t.Errorf("a hostile pinned version reached the HTTP layer: %s", u)
		}
	}
}

// --- parent chain ---

func TestResolveParentChainDetectsCycle(t *testing.T) {
	a := Coord{"t", "a"}
	b := Coord{"t", "b"}
	d := &mavenFakeDoer{responses: map[string]mavenFakeResp{
		pomURL(b, "1.0"): {body: `<project><parent><groupId>t</groupId><artifactId>a</artifactId><version>1.0</version></parent></project>`},
		pomURL(a, "1.0"): {body: `<project><parent><groupId>t</groupId><artifactId>b</artifactId><version>1.0</version></parent></project>`},
	}}

	child := &pomXML{Parent: &pomParent{GroupID: "t", ArtifactID: "b", Version: "1.0"}}
	if _, _, err := resolveParentChain(d, child); err == nil {
		t.Error("resolveParentChain = nil error, want a cycle error")
	}
}

func TestResolveParentChainExceedsMaxDepth(t *testing.T) {
	responses := map[string]mavenFakeResp{}
	// A straight chain of maxParentDepth+2 distinct parents, none repeated,
	// so only the depth cap (never the cycle guard) can stop it.
	n := maxParentDepth + 2
	for i := 0; i < n; i++ {
		me := Coord{"t", fmt.Sprintf("p%d", i)}
		if i+1 < n {
			next := fmt.Sprintf("p%d", i+1)
			responses[pomURL(me, "1.0")] = mavenFakeResp{body: fmt.Sprintf(
				`<project><parent><groupId>t</groupId><artifactId>%s</artifactId><version>1.0</version></parent></project>`, next)}
		} else {
			responses[pomURL(me, "1.0")] = mavenFakeResp{body: `<project></project>`}
		}
	}
	d := &mavenFakeDoer{responses: responses}

	child := &pomXML{Parent: &pomParent{GroupID: "t", ArtifactID: "p0", Version: "1.0"}}
	if _, _, err := resolveParentChain(d, child); err == nil {
		t.Error("resolveParentChain = nil error, want a max-depth error")
	}
}

// --- jarURL ---

func TestJarURL(t *testing.T) {
	a := artifact{Coord: Coord{Group: "com.ibm.mq", Artifact: "com.ibm.mq.jakarta.client"}, Version: "10.0.0.0"}
	want := CentralBase + "/com/ibm/mq/com.ibm.mq.jakarta.client/10.0.0.0/com.ibm.mq.jakarta.client-10.0.0.0.jar"
	if got := jarURL(a); got != want {
		t.Errorf("jarURL = %q, want %q", got, want)
	}
}

// --- shared assertion helper ---

func assertClosure(t *testing.T, got, want []artifact) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("closure has %d artifacts, want %d\ngot:  %+v\nwant: %+v", len(got), len(want), got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("artifact[%d] = %+v, want %+v", i, got[i], want[i])
		}
	}
}
