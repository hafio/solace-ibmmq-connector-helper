package libs

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

// call is one recorded Doer.Do invocation.
type call struct {
	method string
	url    string
}

// response is what fakeDoer returns for a given call. contentLength is the
// value reported on http.Response.ContentLength; left at zero it defaults to
// len(body) (an honest server), so only a test deliberately exercising a
// Content-Length mismatch or an unknown length (-1) needs to set it
// explicitly.
type response struct {
	status        int
	body          string
	header        http.Header
	err           error
	contentLength int64
}

// fakeDoer fakes the Doer boundary. calls records every invocation for later
// assertion. resp is the default response; respByCall overrides it by call
// index, so a later call can fail (or redirect) while earlier ones succeed.
// byURL, when non-nil, overrides both: it answers by the exact request URL
// instead of call order (the shape a multi-fetch Maven resolution -- or a
// jar-plus-sidecar fetch -- needs, where call order isn't the interesting
// thing being pinned), and any URL with no entry answers 404, matching a real
// Maven Central miss; this is also what makes an unfixtured ".sha1" URL
// behave as a real sidecar 404 without a test needing to spell that out.
// sha1For, checked only when byURL is nil, maps a jar URL (no ".sha1" suffix)
// to the correct lowercase hex sha1 of that jar's body, auto-serving it as a
// 200 OK sidecar so a test that only cares about the jar itself (via resp or
// respByCall) does not need to hand-write a sidecar fixture too.
type fakeDoer struct {
	calls      []call
	resp       response
	respByCall map[int]response
	byURL      map[string]response
	sha1For    map[string]string
}

func (f *fakeDoer) Do(req *http.Request) (*http.Response, error) {
	idx := len(f.calls)
	u := req.URL.String()
	f.calls = append(f.calls, call{method: req.Method, url: u})

	var r response
	switch {
	case f.byURL != nil:
		if o, ok := f.byURL[u]; ok {
			r = o
		} else {
			r = response{status: http.StatusNotFound}
		}
	case f.sha1For != nil && strings.HasSuffix(u, sha1SidecarSuffix):
		jarU := strings.TrimSuffix(u, sha1SidecarSuffix)
		if digest, ok := f.sha1For[jarU]; ok {
			r = response{status: http.StatusOK, body: digest}
		} else {
			r = response{status: http.StatusNotFound}
		}
	default:
		r = f.resp
		if o, ok := f.respByCall[idx]; ok {
			r = o
		}
	}
	if r.err != nil {
		return nil, r.err
	}
	h := r.header
	if h == nil {
		h = http.Header{}
	}
	status := r.status
	if status == 0 {
		status = http.StatusOK
	}
	cl := r.contentLength
	if cl == 0 {
		cl = int64(len(r.body))
	}
	return &http.Response{
		StatusCode:    status,
		Status:        fmt.Sprintf("%d %s", status, http.StatusText(status)),
		Header:        h,
		Body:          io.NopCloser(strings.NewReader(r.body)),
		ContentLength: cl,
	}, nil
}

// sha1Hex returns the lowercase hex sha1 digest of s, for building sidecar
// fixtures that match a fixture jar body exactly.
func sha1Hex(s string) string {
	sum := sha1.Sum([]byte(s))
	return hex.EncodeToString(sum[:])
}

func TestDownloadURLHappyPath(t *testing.T) {
	dir := t.TempDir()
	body := "jar-bytes"
	urlStr := "https://repo1.maven.org/maven2/a/b/thing.jar"
	fd := &fakeDoer{resp: response{status: http.StatusOK, body: body}, sha1For: map[string]string{urlStr: sha1Hex(body)}}

	rep, err := Download(Input{Dir: dir, URLs: []string{urlStr}, HTTP: fd})
	if err != nil {
		t.Fatalf("Download: %v", err)
	}
	want := []string{filepath.Join(dir, "thing.jar")}
	if !reflect.DeepEqual(rep.Written, want) {
		t.Errorf("Written = %v, want %v", rep.Written, want)
	}
	if len(rep.Failed) != 0 || len(rep.Skipped) != 0 {
		t.Errorf("Failed/Skipped = %v/%v, want empty", rep.Failed, rep.Skipped)
	}
	if len(rep.Unverified) != 0 {
		t.Errorf("Unverified = %v, want none: sha1 matched", rep.Unverified)
	}
	b, err := os.ReadFile(want[0])
	if err != nil {
		t.Fatal(err)
	}
	if string(b) != body {
		t.Errorf("content = %q, want %q", b, body)
	}
}

func TestDownloadRejectsNonHTTPSURL(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "libs")
	fd := &fakeDoer{}

	rep, err := Download(Input{Dir: dir, URLs: []string{"http://repo1.maven.org/maven2/a.jar"}, HTTP: fd})
	if err == nil {
		t.Fatal("want error for non-https url")
	}
	if !reflect.DeepEqual(rep, Report{}) {
		t.Errorf("Report = %+v, want zero value", rep)
	}
	if len(fd.calls) != 0 {
		t.Errorf("calls = %v, want none", fd.calls)
	}
	if _, statErr := os.Stat(dir); statErr == nil {
		t.Errorf("dir %q was created, want untouched", dir)
	}
}

func TestDownloadRedirectToHTTPRejected(t *testing.T) {
	dir := t.TempDir()
	urlStr := "https://repo1.maven.org/maven2/a.jar"
	fd := &fakeDoer{byURL: map[string]response{
		urlStr: {status: http.StatusFound, header: http.Header{"Location": {"http://evil.example/a.jar"}}},
		// urlStr+".sha1" deliberately unfixtured: byURL defaults it to 404,
		// which is UNVERIFIED for this --url download -- the redirect-to-http
		// rejection below must fire regardless of verification state.
	}}

	rep, err := Download(Input{Dir: dir, URLs: []string{urlStr}, HTTP: fd})
	if err != nil {
		t.Fatalf("Download: %v", err)
	}
	if len(rep.Written) != 0 {
		t.Errorf("Written = %v, want none", rep.Written)
	}
	if len(rep.Failed) != 1 {
		t.Fatalf("Failed = %v, want 1 entry", rep.Failed)
	}
	entries, _ := os.ReadDir(dir)
	if len(entries) != 0 {
		t.Errorf("dir has entries %v, want none", entries)
	}
}

func TestDownloadRedirectToHTTPSFollowed(t *testing.T) {
	dir := t.TempDir()
	urlStr := "https://repo1.maven.org/maven2/a.jar"
	finalURL := "https://repo1.maven.org/maven2/final.jar"
	body := "final-bytes"
	fd := &fakeDoer{byURL: map[string]response{
		urlStr:           {status: http.StatusFound, header: http.Header{"Location": {finalURL}}},
		finalURL:         {status: http.StatusOK, body: body},
		urlStr + ".sha1": {status: http.StatusOK, body: sha1Hex(body)},
	}}

	rep, err := Download(Input{Dir: dir, URLs: []string{urlStr}, HTTP: fd})
	if err != nil {
		t.Fatalf("Download: %v", err)
	}
	// The filename comes from the originally requested URL, not the redirect target.
	want := []string{filepath.Join(dir, "a.jar")}
	if !reflect.DeepEqual(rep.Written, want) {
		t.Errorf("Written = %v, want %v", rep.Written, want)
	}
	// 3 calls: the sha1 sidecar (fetched against the original URL, never the
	// redirect target), the original jar request, then the followed redirect.
	if len(fd.calls) != 3 {
		t.Fatalf("calls = %v, want 3", fd.calls)
	}
	if fd.calls[0].url != urlStr+".sha1" {
		t.Errorf("first call url = %q, want the sha1 sidecar", fd.calls[0].url)
	}
	if fd.calls[2].url != finalURL {
		t.Errorf("third call url = %q", fd.calls[2].url)
	}
	b, err := os.ReadFile(want[0])
	if err != nil {
		t.Fatal(err)
	}
	if string(b) != body {
		t.Errorf("content = %q", b)
	}
}

func TestDownloadSkipsExistingWithoutForce(t *testing.T) {
	dir := t.TempDir()
	dst := filepath.Join(dir, "a.jar")
	if err := os.WriteFile(dst, []byte("old"), 0o644); err != nil {
		t.Fatal(err)
	}
	fd := &fakeDoer{resp: response{status: http.StatusOK, body: "new"}}

	rep, err := Download(Input{Dir: dir, URLs: []string{"https://repo1.maven.org/maven2/a.jar"}, HTTP: fd})
	if err != nil {
		t.Fatalf("Download: %v", err)
	}
	if !reflect.DeepEqual(rep.Skipped, []string{dst}) {
		t.Errorf("Skipped = %v, want [%s]", rep.Skipped, dst)
	}
	if len(fd.calls) != 0 {
		t.Errorf("calls = %v, want none (skip before request)", fd.calls)
	}
	b, _ := os.ReadFile(dst)
	if string(b) != "old" {
		t.Errorf("content = %q, want unchanged", b)
	}
}

func TestDownloadOverwritesWithForce(t *testing.T) {
	dir := t.TempDir()
	dst := filepath.Join(dir, "a.jar")
	if err := os.WriteFile(dst, []byte("old"), 0o644); err != nil {
		t.Fatal(err)
	}
	body := "new"
	urlStr := "https://repo1.maven.org/maven2/a.jar"
	fd := &fakeDoer{resp: response{status: http.StatusOK, body: body}, sha1For: map[string]string{urlStr: sha1Hex(body)}}

	rep, err := Download(Input{Dir: dir, URLs: []string{urlStr}, Force: true, HTTP: fd})
	if err != nil {
		t.Fatalf("Download: %v", err)
	}
	if !reflect.DeepEqual(rep.Written, []string{dst}) {
		t.Errorf("Written = %v, want [%s]", rep.Written, dst)
	}
	b, _ := os.ReadFile(dst)
	if string(b) != "new" {
		t.Errorf("content = %q, want overwritten", b)
	}
}

func TestDownloadPerArtifactFailureDoesNotBlockOthers(t *testing.T) {
	dir := t.TempDir()
	missingURL := "https://repo1.maven.org/maven2/missing.jar"
	presentURL := "https://repo1.maven.org/maven2/present.jar"
	body := "ok-bytes"
	fd := &fakeDoer{byURL: map[string]response{
		missingURL:           {status: http.StatusNotFound},
		presentURL:           {status: http.StatusOK, body: body},
		presentURL + ".sha1": {status: http.StatusOK, body: sha1Hex(body)},
		// missing.jar.sha1 deliberately unfixtured: defaults to 404
		// (UNVERIFIED), but missing.jar itself still 404s on its own fetch.
	}}

	rep, err := Download(Input{
		Dir:  dir,
		URLs: []string{missingURL, presentURL},
		HTTP: fd,
	})
	if err != nil {
		t.Fatalf("Download: %v", err)
	}
	wantWritten := []string{filepath.Join(dir, "present.jar")}
	if !reflect.DeepEqual(rep.Written, wantWritten) {
		t.Errorf("Written = %v, want %v", rep.Written, wantWritten)
	}
	if len(rep.Failed) != 1 || rep.Failed[0].Name != "missing.jar" {
		t.Errorf("Failed = %v, want one entry for missing.jar", rep.Failed)
	}
}

func TestDownloadUncreatableDirIsSystemic(t *testing.T) {
	parent := t.TempDir()
	blocker := filepath.Join(parent, "blocker")
	if err := os.WriteFile(blocker, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	dir := filepath.Join(blocker, "sub") // blocker is a file, not a directory

	rep, err := Download(Input{Dir: dir, Set: SetMQ, HTTP: &fakeDoer{}})
	if err == nil {
		t.Fatal("want error for uncreatable dir")
	}
	if !reflect.DeepEqual(rep, Report{}) {
		t.Errorf("Report = %+v, want zero value", rep)
	}
}

func TestDownloadUnknownSetIsSystemic(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "libs")

	rep, err := Download(Input{Dir: dir, Set: "bogus", HTTP: &fakeDoer{}})
	if err == nil {
		t.Fatal("want error for unknown set")
	}
	if !strings.Contains(err.Error(), "bogus") || !strings.Contains(err.Error(), SetMQ) || !strings.Contains(err.Error(), SetSyslog) {
		t.Errorf("err = %v, want offending value and valid list", err)
	}
	if !reflect.DeepEqual(rep, Report{}) {
		t.Errorf("Report = %+v, want zero value", rep)
	}
	if _, statErr := os.Stat(dir); statErr == nil {
		t.Errorf("dir %q was created, want untouched", dir)
	}
}

// TestResolveSeedPerSet pins each set's seed coordinate. The mq seed is the
// Jakarta build and there is no longer a choice: the connector image is a
// Jakarta stack, so the javax build (com.ibm.mq.allclient) could only ever
// produce a classpath that fails at run time.
func TestResolveSeedPerSet(t *testing.T) {
	got, err := resolveSeed(SetMQ)
	if err != nil {
		t.Fatal(err)
	}
	if want := (Coord{Group: "com.ibm.mq", Artifact: "com.ibm.mq.jakarta.client"}); got != want {
		t.Errorf("resolveSeed(mq) = %+v, want %+v", got, want)
	}
	if got.Artifact == "com.ibm.mq.allclient" {
		t.Error("the javax build must never be seeded: it cannot satisfy a jakarta.jms binder")
	}

	got, err = resolveSeed(SetSyslog)
	if err != nil {
		t.Fatal(err)
	}
	want := Coord{Group: "net.logstash.logback", Artifact: "logstash-logback-encoder"}
	if got != want {
		t.Errorf("resolveSeed(syslog) = %+v, want %+v", got, want)
	}
}

func TestValidateFilenameShapeRejections(t *testing.T) {
	cases := []string{
		"",
		".",
		"..",
		"a/b.jar",
		"a\\b.jar",
		"/abs.jar",
		"readme.txt",
		"bad\x01name.jar",
	}
	for _, name := range cases {
		if err := validateFilenameShape(name); err == nil {
			t.Errorf("validateFilenameShape(%q) = nil, want error", name)
		}
	}
}

func TestValidateFilenameShapeAccepts(t *testing.T) {
	if err := validateFilenameShape("thing-1.0.0.jar"); err != nil {
		t.Errorf("validateFilenameShape valid name: %v", err)
	}
}

func TestFilenameFromEscapedPathRejectsEscapedTraversal(t *testing.T) {
	// %2e%2e%2f decodes to "../", which -- because decoding happens only
	// after the segment is isolated -- must be caught by the separator check
	// on the decoded value, not waved through because the raw segment looked
	// like a clean name.
	if _, err := filenameFromEscapedPath("/lib/%2e%2e%2fescape.jar"); err == nil {
		t.Fatal("want error for escaped traversal segment")
	}
}

func TestFilenameFromEscapedPathAcceptsPlainName(t *testing.T) {
	name, err := filenameFromEscapedPath("/a/b/thing.jar")
	if err != nil {
		t.Fatal(err)
	}
	if name != "thing.jar" {
		t.Errorf("name = %q, want thing.jar", name)
	}
}

func TestDownloadByteCapTripLeavesNoTempFile(t *testing.T) {
	orig := maxArtifactBytes
	maxArtifactBytes = 16
	defer func() { maxArtifactBytes = orig }()

	dir := t.TempDir()
	urlStr := "https://repo1.maven.org/maven2/big.jar"
	fd := &fakeDoer{byURL: map[string]response{
		urlStr: {status: http.StatusOK, body: strings.Repeat("x", 100)},
		// urlStr+".sha1" unfixtured: defaults to 404 (UNVERIFIED), which is
		// enough to exercise the cap trip -- the cap check fires regardless
		// of whether a digest was verified.
	}}

	rep, err := Download(Input{Dir: dir, URLs: []string{urlStr}, HTTP: fd})
	if err != nil {
		t.Fatalf("Download: %v", err)
	}
	if len(rep.Failed) != 1 {
		t.Fatalf("Failed = %v, want 1 entry", rep.Failed)
	}
	if len(rep.Written) != 0 {
		t.Errorf("Written = %v, want none", rep.Written)
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 0 {
		t.Errorf("dir has leftover entries %v, want none (no temp file)", entries)
	}
}

func TestDownloadUserinfoNotInOutput(t *testing.T) {
	dir := t.TempDir()
	fd := &fakeDoer{resp: response{err: fmt.Errorf("connection reset")}}

	rep, err := Download(Input{Dir: dir, URLs: []string{"https://user:s3cr3t@repo1.maven.org/maven2/a.jar"}, HTTP: fd})
	if err != nil {
		t.Fatalf("Download: %v", err)
	}
	if len(rep.Failed) != 1 {
		t.Fatalf("Failed = %v, want 1 entry", rep.Failed)
	}
	if strings.Contains(rep.Failed[0].Name, "s3cr3t") || strings.Contains(rep.Failed[0].Name, "user") {
		t.Errorf("Failure.Name leaked userinfo: %q", rep.Failed[0].Name)
	}
	if strings.Contains(rep.Failed[0].Err.Error(), "s3cr3t") || strings.Contains(rep.Failed[0].Err.Error(), "user:") {
		t.Errorf("Failure.Err leaked userinfo: %v", rep.Failed[0].Err)
	}
}

func TestDownloadTooManyRedirectsFails(t *testing.T) {
	dir := t.TempDir()
	byURL := map[string]response{}
	for i := 0; i <= maxRedirectHops+1; i++ {
		from := fmt.Sprintf("https://repo1.maven.org/maven2/hop%d.jar", i)
		to := fmt.Sprintf("https://repo1.maven.org/maven2/hop%d.jar", i+1)
		byURL[from] = response{status: http.StatusFound, header: http.Header{"Location": {to}}}
	}
	fd := &fakeDoer{byURL: byURL}

	rep, err := Download(Input{Dir: dir, URLs: []string{"https://repo1.maven.org/maven2/hop0.jar"}, HTTP: fd})
	if err != nil {
		t.Fatalf("Download: %v", err)
	}
	if len(rep.Written) != 0 {
		t.Errorf("Written = %v, want none", rep.Written)
	}
	if len(rep.Failed) != 1 {
		t.Fatalf("Failed = %v, want 1 entry", rep.Failed)
	}
	if !strings.Contains(rep.Failed[0].Err.Error(), "too many redirects") {
		t.Errorf("Err = %v, want a too-many-redirects failure", rep.Failed[0].Err)
	}
}

func TestDownloadEmptyBodyLeavesNoTempFile(t *testing.T) {
	dir := t.TempDir()
	urlStr := "https://repo1.maven.org/maven2/empty.jar"
	fd := &fakeDoer{byURL: map[string]response{
		urlStr: {status: http.StatusOK, body: ""},
	}}

	rep, err := Download(Input{Dir: dir, URLs: []string{urlStr}, HTTP: fd})
	if err != nil {
		t.Fatalf("Download: %v", err)
	}
	if len(rep.Written) != 0 {
		t.Errorf("Written = %v, want none", rep.Written)
	}
	if len(rep.Failed) != 1 {
		t.Fatalf("Failed = %v, want 1 entry", rep.Failed)
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 0 {
		t.Errorf("dir has leftover entries %v, want none (no temp file, no empty jar)", entries)
	}
}

func TestDownloadContentLengthMismatchLeavesNoTempFile(t *testing.T) {
	dir := t.TempDir()
	urlStr := "https://repo1.maven.org/maven2/mismatch.jar"
	fd := &fakeDoer{byURL: map[string]response{
		urlStr: {status: http.StatusOK, body: "short-body", contentLength: 999},
	}}

	rep, err := Download(Input{Dir: dir, URLs: []string{urlStr}, HTTP: fd})
	if err != nil {
		t.Fatalf("Download: %v", err)
	}
	if len(rep.Written) != 0 {
		t.Errorf("Written = %v, want none", rep.Written)
	}
	if len(rep.Failed) != 1 {
		t.Fatalf("Failed = %v, want 1 entry", rep.Failed)
	}
	if !strings.Contains(rep.Failed[0].Err.Error(), "Content-Length") {
		t.Errorf("Err = %v, want a Content-Length mismatch failure", rep.Failed[0].Err)
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 0 {
		t.Errorf("dir has leftover entries %v, want none (no temp file)", entries)
	}
}

// --- sha1 sidecar verification ---

func TestDownloadSHA1MatchSucceeds(t *testing.T) {
	dir := t.TempDir()
	body := "jar-bytes"
	urlStr := "https://repo1.maven.org/maven2/thing.jar"
	fd := &fakeDoer{byURL: map[string]response{
		urlStr:           {status: http.StatusOK, body: body},
		urlStr + ".sha1": {status: http.StatusOK, body: sha1Hex(body)},
	}}

	rep, err := Download(Input{Dir: dir, URLs: []string{urlStr}, HTTP: fd})
	if err != nil {
		t.Fatalf("Download: %v", err)
	}
	want := []string{filepath.Join(dir, "thing.jar")}
	if !reflect.DeepEqual(rep.Written, want) {
		t.Errorf("Written = %v, want %v", rep.Written, want)
	}
	if len(rep.Unverified) != 0 {
		t.Errorf("Unverified = %v, want none: a matching sha1 was verified", rep.Unverified)
	}
}

func TestDownloadSHA1MismatchFailsAndLeavesNoFile(t *testing.T) {
	dir := t.TempDir()
	body := "jar-bytes"
	urlStr := "https://repo1.maven.org/maven2/thing.jar"
	fd := &fakeDoer{byURL: map[string]response{
		urlStr:           {status: http.StatusOK, body: body},
		urlStr + ".sha1": {status: http.StatusOK, body: strings.Repeat("a", 40)}, // wrong digest
	}}

	rep, err := Download(Input{Dir: dir, URLs: []string{urlStr}, HTTP: fd})
	if err != nil {
		t.Fatalf("Download: %v", err)
	}
	if len(rep.Written) != 0 {
		t.Errorf("Written = %v, want none: sha1 mismatch", rep.Written)
	}
	if len(rep.Failed) != 1 || !strings.Contains(rep.Failed[0].Err.Error(), "sha1") {
		t.Errorf("Failed = %v, want one sha1 mismatch failure", rep.Failed)
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 0 {
		t.Errorf("dir has leftover entries %v, want none", entries)
	}
}

func TestDownloadMalformedSHA1SidecarRejected(t *testing.T) {
	cases := []struct {
		name string
		body string
	}{
		{"too short", "abc123"},
		{"non-hex", strings.Repeat("z", 40)},
		{"empty", ""},
		{"html body", "<html><body>error</body></html>"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			dir := t.TempDir()
			body := "jar-bytes"
			urlStr := "https://repo1.maven.org/maven2/thing.jar"
			fd := &fakeDoer{byURL: map[string]response{
				urlStr:           {status: http.StatusOK, body: body},
				urlStr + ".sha1": {status: http.StatusOK, body: c.body},
			}}

			rep, err := Download(Input{Dir: dir, URLs: []string{urlStr}, HTTP: fd})
			if err != nil {
				t.Fatalf("Download: %v", err)
			}
			if len(rep.Written) != 0 {
				t.Errorf("Written = %v, want none: malformed sidecar", rep.Written)
			}
			if len(rep.Failed) != 1 {
				t.Fatalf("Failed = %v, want 1 entry", rep.Failed)
			}
		})
	}
}

func TestDownloadURLUnverifiedOn404SidecarStillWrites(t *testing.T) {
	dir := t.TempDir()
	body := "jar-bytes"
	urlStr := "https://repo1.maven.org/maven2/thing.jar"
	fd := &fakeDoer{byURL: map[string]response{
		urlStr: {status: http.StatusOK, body: body},
		// urlStr+".sha1" deliberately unfixtured: defaults to 404.
	}}

	rep, err := Download(Input{Dir: dir, URLs: []string{urlStr}, HTTP: fd})
	if err != nil {
		t.Fatalf("Download: %v", err)
	}
	want := []string{filepath.Join(dir, "thing.jar")}
	if !reflect.DeepEqual(rep.Written, want) {
		t.Errorf("Written = %v, want %v: a 404 sidecar on --url must still write", rep.Written, want)
	}
	if len(rep.Unverified) != 1 {
		t.Fatalf("Unverified = %v, want 1 entry", rep.Unverified)
	}
	if len(rep.Failed) != 0 {
		t.Errorf("Failed = %v, want none", rep.Failed)
	}
}

func TestDownloadMavenResolved404SidecarFailsArtifact(t *testing.T) {
	dir := t.TempDir()
	seed := Coord{Group: "net.logstash.logback", Artifact: "logstash-logback-encoder"}
	body := "jar-bytes"
	fixtures := syslogFixtures("9.0")
	fixtures[jarURL(artifact{Coord: seed, Version: "9.0"})] = response{status: http.StatusOK, body: body}
	// Deliberately no ".sha1" fixture: byURL defaults it to 404, and a
	// Maven-resolved artifact must never be written unverified.
	fd := &fakeDoer{byURL: fixtures}

	rep, err := Download(Input{Dir: dir, Set: SetSyslog, Version: "9.0", HTTP: fd})
	if err != nil {
		t.Fatalf("Download: %v", err)
	}
	if len(rep.Written) != 0 {
		t.Errorf("Written = %v, want none: sha1 sidecar missing for a Maven-resolved artifact", rep.Written)
	}
	if len(rep.Failed) != 1 {
		t.Fatalf("Failed = %v, want 1 entry", rep.Failed)
	}
	if len(rep.Unverified) != 0 {
		t.Errorf("Unverified = %v, want none: a Maven-resolved artifact is never UNVERIFIED, only failed", rep.Unverified)
	}
}

func TestDownloadContentLengthUnknownWithGoodSHA1Succeeds(t *testing.T) {
	dir := t.TempDir()
	body := "jar-bytes"
	urlStr := "https://repo1.maven.org/maven2/thing.jar"
	fd := &fakeDoer{byURL: map[string]response{
		urlStr:           {status: http.StatusOK, body: body, contentLength: -1},
		urlStr + ".sha1": {status: http.StatusOK, body: sha1Hex(body)},
	}}

	rep, err := Download(Input{Dir: dir, URLs: []string{urlStr}, HTTP: fd})
	if err != nil {
		t.Fatalf("Download: %v", err)
	}
	want := []string{filepath.Join(dir, "thing.jar")}
	if !reflect.DeepEqual(rep.Written, want) {
		t.Errorf("Written = %v, want %v: a verified sha1 must succeed despite unknown Content-Length", rep.Written, want)
	}
}

func TestDownloadContentLengthUnknownWithNoDigestFails(t *testing.T) {
	dir := t.TempDir()
	body := "jar-bytes"
	urlStr := "https://repo1.maven.org/maven2/thing.jar"
	fd := &fakeDoer{byURL: map[string]response{
		urlStr: {status: http.StatusOK, body: body, contentLength: -1},
		// urlStr+".sha1" unfixtured: 404 -> unverified, so there is neither a
		// truthful Content-Length nor a verified digest.
	}}

	rep, err := Download(Input{Dir: dir, URLs: []string{urlStr}, HTTP: fd})
	if err != nil {
		t.Fatalf("Download: %v", err)
	}
	if len(rep.Written) != 0 {
		t.Errorf("Written = %v, want none: no Content-Length and no sha1 verified", rep.Written)
	}
	if len(rep.Failed) != 1 {
		t.Fatalf("Failed = %v, want 1 entry", rep.Failed)
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 0 {
		t.Errorf("dir has leftover entries %v, want none", entries)
	}
}

// --- Version pinning, image-aware omission ---

// syslogFixtures returns the byURL fixtures for resolving the syslog seed,
// pinned at version, down to a closure of exactly the seed itself (no
// dependencies) -- enough to exercise Download's omission step without
// dragging in the full dependency-walk fixtures maven_test.go builds for
// resolveClosure itself.
func syslogFixtures(version string) map[string]response {
	seed := Coord{Group: "net.logstash.logback", Artifact: "logstash-logback-encoder"}
	return map[string]response{
		metadataURL(seed):     {body: metaXML(version, version)},
		pomURL(seed, version): {body: pomXMLBody()},
	}
}

// syslogFixturesWithDependency is syslogFixtures plus exactly one non-seed
// dependency, at depVersion -- just enough to let a test exercise omission
// against a real DEPENDENCY rather than the seed itself (the seed is never a
// candidate for omission; see TestDownloadNeverOmitsSeedEvenWhenImageClaimsHugeVersion).
func syslogFixturesWithDependency(seedVersion, depVersion string) (fixtures map[string]response, depCoord Coord) {
	seed := Coord{Group: "net.logstash.logback", Artifact: "logstash-logback-encoder"}
	depCoord = Coord{Group: "org.example", Artifact: "helper-lib"}
	fixtures = map[string]response{
		metadataURL(seed):            {body: metaXML(seedVersion, seedVersion)},
		pomURL(seed, seedVersion):    {body: pomXMLBody(dep("org.example", "helper-lib", depVersion, "", "", ""))},
		pomURL(depCoord, depVersion): {body: pomXMLBody()},
	}
	return fixtures, depCoord
}

func writeOmitListFile(t *testing.T, lines ...string) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), "omit-list.txt")
	if err := os.WriteFile(p, []byte(strings.Join(lines, "\n")+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	return p
}

// writeRawOmitListFile writes content verbatim, unlike writeOmitListFile
// (which always appends a trailing newline to its joined lines) -- this is
// the only way to produce a genuinely 0-byte omit list file for a test.
func writeRawOmitListFile(t *testing.T, content string) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), "omit-list.txt")
	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	return p
}

func TestDownloadOmitsDependencyImageProvidesAtNewerVersion(t *testing.T) {
	dir := t.TempDir()
	seed := Coord{Group: "net.logstash.logback", Artifact: "logstash-logback-encoder"}
	fixtures, depCoord := syslogFixturesWithDependency("9.0", "2.0")
	seedBody := "seed-bytes"
	seedJarURL := jarURL(artifact{Coord: seed, Version: "9.0"})
	fixtures[seedJarURL] = response{status: http.StatusOK, body: seedBody}
	fixtures[seedJarURL+".sha1"] = response{status: http.StatusOK, body: sha1Hex(seedBody)}
	fd := &fakeDoer{byURL: fixtures}
	imgPath := writeOmitListFile(t, "helper-lib-2.1.jar")

	rep, err := Download(Input{Dir: dir, Set: SetSyslog, Version: "9.0", OmitLibFile: imgPath, HTTP: fd})
	if err != nil {
		t.Fatalf("Download: %v", err)
	}
	wantWritten := []string{filepath.Join(dir, "logstash-logback-encoder-9.0.jar")}
	if !reflect.DeepEqual(rep.Written, wantWritten) {
		t.Errorf("Written = %v, want %v: only the seed -- the dependency must be omitted", rep.Written, wantWritten)
	}
	wantOmitted := []string{"helper-lib-2.0.jar: the image has 2.1"}
	if !reflect.DeepEqual(rep.Omitted, wantOmitted) {
		t.Errorf("Omitted = %v, want %v", rep.Omitted, wantOmitted)
	}
	depJarURLWanted := jarURL(artifact{Coord: depCoord, Version: "2.0"})
	for _, c := range fd.calls {
		if c.url == depJarURLWanted {
			t.Errorf("dependency jar url %q was requested despite being omitted", depJarURLWanted)
		}
	}
}

// TestDownloadEmptyOmitListFileOmitsNothing and
// TestDownloadCommentsOnlyOmitListFileOmitsNothing pin the documented
// semantics that --omit-lib-file REPLACES the embedded default rather than
// merging with it: a supplied file that (after parsing) yields no entries at
// all must omit NOTHING, so the whole resolved closure -- seed and every
// dependency -- downloads, exactly as --include-provided would produce. Both
// exercise the same fixture (a seed plus one real, non-seed dependency) so
// there is something a non-empty list COULD have omitted, proving the empty
// list's silence is what let it through, not the absence of any candidate.
func TestDownloadEmptyOmitListFileOmitsNothing(t *testing.T) {
	dir := t.TempDir()
	seed := Coord{Group: "net.logstash.logback", Artifact: "logstash-logback-encoder"}
	fixtures, depCoord := syslogFixturesWithDependency("9.0", "2.0")
	seedBody := "seed-bytes"
	depBody := "dep-bytes"
	seedJarURL := jarURL(artifact{Coord: seed, Version: "9.0"})
	depJarURL := jarURL(artifact{Coord: depCoord, Version: "2.0"})
	fixtures[seedJarURL] = response{status: http.StatusOK, body: seedBody}
	fixtures[seedJarURL+".sha1"] = response{status: http.StatusOK, body: sha1Hex(seedBody)}
	fixtures[depJarURL] = response{status: http.StatusOK, body: depBody}
	fixtures[depJarURL+".sha1"] = response{status: http.StatusOK, body: sha1Hex(depBody)}
	fd := &fakeDoer{byURL: fixtures}
	imgPath := writeRawOmitListFile(t, "")

	rep, err := Download(Input{Dir: dir, Set: SetSyslog, Version: "9.0", OmitLibFile: imgPath, HTTP: fd})
	if err != nil {
		t.Fatalf("Download: %v", err)
	}
	wantWritten := []string{
		filepath.Join(dir, "helper-lib-2.0.jar"),
		filepath.Join(dir, "logstash-logback-encoder-9.0.jar"),
	}
	if !reflect.DeepEqual(rep.Written, wantWritten) {
		t.Errorf("Written = %v, want %v: an empty omit list must omit nothing", rep.Written, wantWritten)
	}
	if len(rep.Omitted) != 0 {
		t.Errorf("Omitted = %v, want none: an empty omit list replaces the default with nothing", rep.Omitted)
	}
	if rep.OmitListProvenance != imgPath {
		t.Errorf("OmitListProvenance = %q, want %q: a supplied-but-empty file is still a supplied file", rep.OmitListProvenance, imgPath)
	}
}

func TestDownloadCommentsOnlyOmitListFileOmitsNothing(t *testing.T) {
	dir := t.TempDir()
	seed := Coord{Group: "net.logstash.logback", Artifact: "logstash-logback-encoder"}
	fixtures, depCoord := syslogFixturesWithDependency("9.0", "2.0")
	seedBody := "seed-bytes"
	depBody := "dep-bytes"
	seedJarURL := jarURL(artifact{Coord: seed, Version: "9.0"})
	depJarURL := jarURL(artifact{Coord: depCoord, Version: "2.0"})
	fixtures[seedJarURL] = response{status: http.StatusOK, body: seedBody}
	fixtures[seedJarURL+".sha1"] = response{status: http.StatusOK, body: sha1Hex(seedBody)}
	fixtures[depJarURL] = response{status: http.StatusOK, body: depBody}
	fixtures[depJarURL+".sha1"] = response{status: http.StatusOK, body: sha1Hex(depBody)}
	fd := &fakeDoer{byURL: fixtures}
	imgPath := writeOmitListFile(t, "# nothing provided by this image", "", "   ")

	rep, err := Download(Input{Dir: dir, Set: SetSyslog, Version: "9.0", OmitLibFile: imgPath, HTTP: fd})
	if err != nil {
		t.Fatalf("Download: %v", err)
	}
	wantWritten := []string{
		filepath.Join(dir, "helper-lib-2.0.jar"),
		filepath.Join(dir, "logstash-logback-encoder-9.0.jar"),
	}
	if !reflect.DeepEqual(rep.Written, wantWritten) {
		t.Errorf("Written = %v, want %v: a comments-and-blank-lines-only omit list must omit nothing", rep.Written, wantWritten)
	}
	if len(rep.Omitted) != 0 {
		t.Errorf("Omitted = %v, want none", rep.Omitted)
	}
	if rep.OmitListProvenance != imgPath {
		t.Errorf("OmitListProvenance = %q, want %q: a supplied-but-empty file is still a supplied file", rep.OmitListProvenance, imgPath)
	}
}

// TestDownloadNeverOmitsSeedEvenWhenImageClaimsHugeVersion pins the PoC from
// the pass-2 review: an omit list entry naming the seed's own artifact at an
// absurd version ("com.ibm.mq.jakarta.client-99999.jar") would satisfy
// imageSatisfies for any ordinary dependency, but the seed is never a
// candidate for omission at all -- Download identifies it by Coord equality,
// not by trusting the omit list.
func TestDownloadNeverOmitsSeedEvenWhenImageClaimsHugeVersion(t *testing.T) {
	dir := t.TempDir()
	seed := Coord{Group: "com.ibm.mq", Artifact: "com.ibm.mq.jakarta.client"}
	body := "mq-client-bytes"
	seedJarURL := jarURL(artifact{Coord: seed, Version: "9.4.2.0"})
	fixtures := map[string]response{
		pomURL(seed, "9.4.2.0"): {body: pomXMLBody()},
		seedJarURL:              {status: http.StatusOK, body: body},
		seedJarURL + ".sha1":    {status: http.StatusOK, body: sha1Hex(body)},
	}
	fd := &fakeDoer{byURL: fixtures}
	imgPath := writeOmitListFile(t, "com.ibm.mq.jakarta.client-99999.jar")

	rep, err := Download(Input{Dir: dir, Set: SetMQ, Version: "9.4.2.0", OmitLibFile: imgPath, HTTP: fd})
	if err != nil {
		t.Fatalf("Download: %v", err)
	}
	want := []string{filepath.Join(dir, "com.ibm.mq.jakarta.client-9.4.2.0.jar")}
	if !reflect.DeepEqual(rep.Written, want) {
		t.Errorf("Written = %v, want %v: the seed must never be omitted", rep.Written, want)
	}
	if len(rep.Omitted) != 0 {
		t.Errorf("Omitted = %v, want none: the seed is not a candidate for omission", rep.Omitted)
	}
}

func TestDownloadDownloadsArtifactImageHasOlderVersion(t *testing.T) {
	dir := t.TempDir()
	seed := Coord{Group: "net.logstash.logback", Artifact: "logstash-logback-encoder"}
	fixtures := syslogFixtures("9.0")
	body := "jar-bytes"
	seedJarURL := jarURL(artifact{Coord: seed, Version: "9.0"})
	fixtures[seedJarURL] = response{status: http.StatusOK, body: body}
	fixtures[seedJarURL+".sha1"] = response{status: http.StatusOK, body: sha1Hex(body)}
	fd := &fakeDoer{byURL: fixtures}
	imgPath := writeOmitListFile(t, "logstash-logback-encoder-8.0.jar")

	rep, err := Download(Input{Dir: dir, Set: SetSyslog, Version: "9.0", OmitLibFile: imgPath, HTTP: fd})
	if err != nil {
		t.Fatalf("Download: %v", err)
	}
	want := []string{filepath.Join(dir, "logstash-logback-encoder-9.0.jar")}
	if !reflect.DeepEqual(rep.Written, want) {
		t.Errorf("Written = %v, want %v: the image's version is older, so it must still download", rep.Written, want)
	}
	if len(rep.Omitted) != 0 {
		t.Errorf("Omitted = %v, want none", rep.Omitted)
	}
}

func TestDownloadDownloadsArtifactAbsentFromImage(t *testing.T) {
	dir := t.TempDir()
	seed := Coord{Group: "net.logstash.logback", Artifact: "logstash-logback-encoder"}
	fixtures := syslogFixtures("9.0")
	body := "jar-bytes"
	seedJarURL := jarURL(artifact{Coord: seed, Version: "9.0"})
	fixtures[seedJarURL] = response{status: http.StatusOK, body: body}
	fixtures[seedJarURL+".sha1"] = response{status: http.StatusOK, body: sha1Hex(body)}
	fd := &fakeDoer{byURL: fixtures}
	imgPath := writeOmitListFile(t, "unrelated-jar-1.0.jar")

	rep, err := Download(Input{Dir: dir, Set: SetSyslog, Version: "9.0", OmitLibFile: imgPath, HTTP: fd})
	if err != nil {
		t.Fatalf("Download: %v", err)
	}
	want := []string{filepath.Join(dir, "logstash-logback-encoder-9.0.jar")}
	if !reflect.DeepEqual(rep.Written, want) {
		t.Errorf("Written = %v, want %v: absent from the image, so it must download", rep.Written, want)
	}
}

func TestDownloadIncludeProvidedDownloadsEverything(t *testing.T) {
	dir := t.TempDir()
	seed := Coord{Group: "net.logstash.logback", Artifact: "logstash-logback-encoder"}
	fixtures := syslogFixtures("9.0")
	body := "jar-bytes"
	seedJarURL := jarURL(artifact{Coord: seed, Version: "9.0"})
	fixtures[seedJarURL] = response{status: http.StatusOK, body: body}
	fixtures[seedJarURL+".sha1"] = response{status: http.StatusOK, body: sha1Hex(body)}
	fd := &fakeDoer{byURL: fixtures}
	// This would satisfy the omission rule if it were applied.
	imgPath := writeOmitListFile(t, "logstash-logback-encoder-9.9.jar")

	rep, err := Download(Input{Dir: dir, Set: SetSyslog, Version: "9.0", OmitLibFile: imgPath, IncludeProvided: true, HTTP: fd})
	if err != nil {
		t.Fatalf("Download: %v", err)
	}
	want := []string{filepath.Join(dir, "logstash-logback-encoder-9.0.jar")}
	if !reflect.DeepEqual(rep.Written, want) {
		t.Errorf("Written = %v, want %v: --include-provided must download the whole closure", rep.Written, want)
	}
	if len(rep.Omitted) != 0 {
		t.Errorf("Omitted = %v, want none with --include-provided", rep.Omitted)
	}
}

func TestDownloadURLNeverOmittedEvenWhenImageHasIt(t *testing.T) {
	dir := t.TempDir()
	body := "jar-bytes"
	urlStr := "https://repo1.maven.org/maven2/a/b/thing-1.0.0.jar"
	fd := &fakeDoer{resp: response{status: http.StatusOK, body: body}, sha1For: map[string]string{urlStr: sha1Hex(body)}}
	imgPath := writeOmitListFile(t, "thing-1.0.0.jar")

	rep, err := Download(Input{
		Dir:         dir,
		URLs:        []string{urlStr},
		OmitLibFile: imgPath,
		HTTP:        fd,
	})
	if err != nil {
		t.Fatalf("Download: %v", err)
	}
	want := []string{filepath.Join(dir, "thing-1.0.0.jar")}
	if !reflect.DeepEqual(rep.Written, want) {
		t.Errorf("Written = %v, want %v: --url must never be omitted", rep.Written, want)
	}
	if len(rep.Omitted) != 0 {
		t.Errorf("Omitted = %v, want none for --url", rep.Omitted)
	}
}

func TestDownloadBadOmitLibFilePathIsSystemic(t *testing.T) {
	dir := t.TempDir()
	fd := &fakeDoer{byURL: syslogFixtures("9.0")}
	badPath := filepath.Join(t.TempDir(), "does-not-exist.txt")

	rep, err := Download(Input{Dir: dir, Set: SetSyslog, Version: "9.0", OmitLibFile: badPath, HTTP: fd})
	if err == nil {
		t.Fatal("want error for an unreadable --omit-lib-file")
	}
	if !strings.Contains(err.Error(), badPath) {
		t.Errorf("err = %v, want it to name %q", err, badPath)
	}
	if !reflect.DeepEqual(rep, Report{}) {
		t.Errorf("Report = %+v, want zero value", rep)
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 0 {
		t.Errorf("dir has entries %v, want nothing written", entries)
	}
}

// TestDownloadEmbeddedDefaultListLoadFailureIsSystemic exercises Download's
// "loading embedded default omit list" branch: a single line far longer than
// loadImageLibs's own enlarged scanner buffer cannot be split into a token at
// all, so loadImageLibs's own "skip one malformed line" contract cannot save
// it -- that is bufio.Scanner's hard token-size ceiling, not a bug in
// loadImageLibs. This is deliberate, not restructured away: corruption this
// severe in DATA COMPILED INTO THE BINARY is a programming/build error, not
// an operator error, and Download must still surface it as the systemic
// error its own doc comment promises rather than silently falling through.
// embeddedDefaultList is a package-level var (declared in image.go) for
// exactly this kind of test injection.
func TestDownloadEmbeddedDefaultListLoadFailureIsSystemic(t *testing.T) {
	orig := embeddedDefaultList
	embeddedDefaultList = bytes.Repeat([]byte("x"), maxOmitListLineBytes*2)
	defer func() { embeddedDefaultList = orig }()

	dir := t.TempDir()
	fd := &fakeDoer{byURL: syslogFixtures("9.0")}

	rep, err := Download(Input{Dir: dir, Set: SetSyslog, Version: "9.0", HTTP: fd})
	if err == nil {
		t.Fatal("want systemic error when the embedded default list fails to parse")
	}
	if !strings.Contains(err.Error(), "embedded default omit list") {
		t.Errorf("err = %v, want it to name the embedded default", err)
	}
	if !reflect.DeepEqual(rep, Report{}) {
		t.Errorf("Report = %+v, want zero value", rep)
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 0 {
		t.Errorf("dir has entries %v, want nothing written", entries)
	}
}

// TestDownloadReportsOmitListProvenance covers the provenance line, and that
// a rejected omit-list entry NO closure artifact asks about produces no
// warning at all.
//
// This is the noise fix: warnings used to be emitted per rejected line while
// the list was being read, so `download jar mq` reported netty and
// hibernate-validator on every run -- artifacts no mq closure has ever
// referenced. A rejection the download never consulted changed nothing.
func TestDownloadReportsOmitListProvenance(t *testing.T) {
	dir := t.TempDir()
	seed := Coord{Group: "net.logstash.logback", Artifact: "logstash-logback-encoder"}
	fixtures := syslogFixtures("9.0")
	body := "jar-bytes"
	seedJarURL := jarURL(artifact{Coord: seed, Version: "9.0"})
	fixtures[seedJarURL] = response{status: http.StatusOK, body: body}
	fixtures[seedJarURL+".sha1"] = response{status: http.StatusOK, body: sha1Hex(body)}
	fd := &fakeDoer{byURL: fixtures}
	// "not-a-jar-line" fails splitJarBasename's shape entirely and is skipped.
	// "thing-9zzzzzzzzzz.jar" splits fine but its version fails
	// validateImageVersion -- and nothing in this closure is named "thing",
	// so the rejection never gets consulted.
	imgPath := writeOmitListFile(t, "not-a-jar-line", "thing-9zzzzzzzzzz.jar")

	rep, err := Download(Input{Dir: dir, Set: SetSyslog, Version: "9.0", OmitLibFile: imgPath, HTTP: fd})
	if err != nil {
		t.Fatalf("Download: %v", err)
	}
	if rep.OmitListProvenance != imgPath {
		t.Errorf("OmitListProvenance = %q, want %q", rep.OmitListProvenance, imgPath)
	}
	if len(rep.OmitListWarnings) != 0 {
		t.Errorf("OmitListWarnings = %v, want none: no closure artifact is named \"thing\"", rep.OmitListWarnings)
	}
}

// TestDownloadWarnsWhenRejectedEntryAffectedThisClosure is the other half: the
// rejected entry names an artifact this closure DOES resolve, so the rejection
// is why that jar is being downloaded, and the operator is told.
func TestDownloadWarnsWhenRejectedEntryAffectedThisClosure(t *testing.T) {
	dir := t.TempDir()
	seed := Coord{Group: "net.logstash.logback", Artifact: "logstash-logback-encoder"}
	fixtures, depCoord := syslogFixturesWithDependency("9.0", "2.0")
	for _, a := range []artifact{{Coord: seed, Version: "9.0"}, {Coord: depCoord, Version: "2.0"}} {
		body := a.Artifact + "-bytes"
		u := jarURL(a)
		fixtures[u] = response{status: http.StatusOK, body: body}
		fixtures[u+".sha1"] = response{status: http.StatusOK, body: sha1Hex(body)}
	}
	fd := &fakeDoer{byURL: fixtures}
	imgPath := writeOmitListFile(t, "helper-lib-9zzzzzzzzzz.jar")

	rep, err := Download(Input{Dir: dir, Set: SetSyslog, Version: "9.0", OmitLibFile: imgPath, HTTP: fd})
	if err != nil {
		t.Fatalf("Download: %v", err)
	}
	if len(rep.OmitListWarnings) != 1 {
		t.Fatalf("OmitListWarnings = %v, want exactly 1 -- helper-lib is in this closure", rep.OmitListWarnings)
	}
	if !strings.Contains(rep.OmitListWarnings[0], "helper-lib") {
		t.Errorf("warning %q should name the artifact it affected", rep.OmitListWarnings[0])
	}
	// The point of the warning: the jar was downloaded rather than omitted,
	// because its image entry could not be compared.
	if len(rep.Omitted) != 0 {
		t.Errorf("Omitted = %v, want none -- the entry was unusable", rep.Omitted)
	}
}

func TestDownloadVersionPinsSeed(t *testing.T) {
	dir := t.TempDir()
	seed := Coord{Group: "net.logstash.logback", Artifact: "logstash-logback-encoder"}
	fixtures := syslogFixtures("7.4")
	body := "jar-bytes"
	seedJarURL := jarURL(artifact{Coord: seed, Version: "7.4"})
	fixtures[seedJarURL] = response{status: http.StatusOK, body: body}
	fixtures[seedJarURL+".sha1"] = response{status: http.StatusOK, body: sha1Hex(body)}
	fd := &fakeDoer{byURL: fixtures}
	imgPath := writeOmitListFile(t, "unrelated-1.0.jar")

	rep, err := Download(Input{Dir: dir, Set: SetSyslog, Version: "7.4", OmitLibFile: imgPath, HTTP: fd})
	if err != nil {
		t.Fatalf("Download: %v", err)
	}
	want := []string{filepath.Join(dir, "logstash-logback-encoder-7.4.jar")}
	if !reflect.DeepEqual(rep.Written, want) {
		t.Errorf("Written = %v, want %v: --version must pin the seed release", rep.Written, want)
	}
}

func TestDownloadVersionRejectsPathEscape(t *testing.T) {
	fd := &fakeDoer{}

	rep, err := Download(Input{Dir: t.TempDir(), Set: SetSyslog, Version: "../../evil", HTTP: fd})
	if err == nil {
		t.Fatal("want error for a --version containing a path escape")
	}
	if !reflect.DeepEqual(rep, Report{}) {
		t.Errorf("Report = %+v, want zero value", rep)
	}
	if len(fd.calls) != 0 {
		t.Errorf("calls = %v, want none: rejected before any network access", fd.calls)
	}
}

func TestDownloadSetPathAlwaysResolvesEvenWhenFilesExist(t *testing.T) {
	// Unlike the --url path (TestDownloadSkipsExistingWithoutForce), the
	// Set-based path cannot know its target filenames -- and so cannot check
	// the destination for existing files -- until the closure is resolved:
	// a filename is derived from the resolved artifact's version, not from
	// the seed coordinate alone. This is a deliberate choice (see the comment
	// in Download), pinned here: even with every plausible file already on
	// disk, Download must still contact Maven Central and surface a
	// resolution failure rather than quietly reporting everything skipped.
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "irrelevant.jar"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	fd := &fakeDoer{resp: response{err: fmt.Errorf("network unreachable")}}

	_, err := Download(Input{Dir: dir, Set: SetMQ, HTTP: fd})
	if err == nil {
		t.Fatal("want a systemic resolution error, since resolution happens before any existence check")
	}
	if len(fd.calls) == 0 {
		t.Error("calls = none, want at least one: resolution must be attempted despite pre-existing files")
	}
}

func TestSetNames(t *testing.T) {
	if want := []string{SetMQ, SetSyslog}; !reflect.DeepEqual(SetNames(), want) {
		t.Errorf("SetNames() = %v, want %v", SetNames(), want)
	}
}

// TestDownloadImageMismatchReported covers the end-to-end wiring of the
// deployed-image check, including the cases that must stay quiet.
//
// The embedded list describes every release from EmbeddedListMinVersion
// onwards, not just the tag it was captured from, so "deployed tag != captured
// tag" is NOT what makes this warn -- only an image the list cannot speak for
// does.
//
// --omit-lib-file suppresses it entirely: the operator named a list, and
// second-guessing that would contradict the same rule which makes an explicit
// --url immune to omission.
func TestDownloadImageMismatchReported(t *testing.T) {
	// Deliberately below EmbeddedListMinVersion, so the check has something
	// real to find. Every suppression case below uses this same reference --
	// a covered image would make them pass whether suppression works or not.
	const uncovered = "solace/solace-pubsub-connector-ibmmq:2.9.0"

	run := func(t *testing.T, deployed, omitFile string) Report {
		t.Helper()
		seed := Coord{Group: "net.logstash.logback", Artifact: "logstash-logback-encoder"}
		fixtures := syslogFixtures("9.0")
		body := "jar-bytes"
		u := jarURL(artifact{Coord: seed, Version: "9.0"})
		fixtures[u] = response{status: http.StatusOK, body: body}
		fixtures[u+".sha1"] = response{status: http.StatusOK, body: sha1Hex(body)}
		rep, err := Download(Input{
			Dir: t.TempDir(), Set: SetSyslog, Version: "9.0",
			OmitLibFile: omitFile, DeployedImage: deployed, HTTP: &fakeDoer{byURL: fixtures},
		})
		if err != nil {
			t.Fatalf("Download: %v", err)
		}
		return rep
	}

	t.Run("an image the list cannot speak for warns", func(t *testing.T) {
		rep := run(t, uncovered, "")
		if rep.OmitListImageMismatch == "" {
			t.Fatal("want a mismatch warning: 2.9.0 predates the embedded list's floor")
		}
		if !strings.Contains(rep.OmitListImageMismatch, "2.9.0") {
			t.Errorf("warning %q should name the deployed image", rep.OmitListImageMismatch)
		}
	})

	// Both ends of the covered range, because the bug this replaced was a
	// warning on every 2.14.1 run -- an image the list describes perfectly.
	for _, tag := range []string{"2.13.0", "2.14.1"} {
		t.Run("a covered release is silent/"+tag, func(t *testing.T) {
			rep := run(t, "solace/solace-pubsub-connector-ibmmq:"+tag, "")
			if rep.OmitListImageMismatch != "" {
				t.Errorf("%s is described by the embedded list, want silence, got %q", tag, rep.OmitListImageMismatch)
			}
		})
	}

	t.Run("no declared image is silent", func(t *testing.T) {
		if rep := run(t, "", ""); rep.OmitListImageMismatch != "" {
			t.Errorf("no declared image should be silent, got %q", rep.OmitListImageMismatch)
		}
	})

	t.Run("--omit-lib-file suppresses the check", func(t *testing.T) {
		own := writeOmitListFile(t, "helper-lib-1.0.jar")
		rep := run(t, uncovered, own)
		if rep.OmitListImageMismatch != "" {
			t.Errorf("a named list must not be second-guessed, got %q", rep.OmitListImageMismatch)
		}
	})
}
