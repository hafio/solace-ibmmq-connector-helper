package gen

import (
	"encoding/asn1"
	"testing"
)

// pfxShellOf is a structurally valid PKCS#12 envelope (version 3, a data
// AuthSafe) carrying payload as an unreadable body, with no MAC: it passes
// the structure half of checkExistingStore and fails the password half with a
// non-password error. It stands in for a store wherever only its bytes, not
// its contents, are what a test is about.
func pfxShellOf(t testing.TB, payload []byte) []byte {
	t.Helper()
	b, err := asn1.Marshal(existingPKCS12Shell{
		Version: 3,
		AuthSafe: shellContentInfo{
			ContentType: asn1.ObjectIdentifier{1, 2, 840, 113549, 1, 7, 1},
			Content:     asn1.RawValue{Class: asn1.ClassUniversal, Tag: asn1.TagOctetString, Bytes: payload},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	return b
}

// pfxShell is a pfxShellOf of exactly n bytes, for the size limits.
func pfxShell(t testing.TB, n int) []byte {
	t.Helper()
	body := n
	for i := 0; i < 8; i++ {
		b := pfxShellOf(t, make([]byte, body))
		if len(b) == n {
			return b
		}
		body += n - len(b)
	}
	t.Fatalf("no PKCS12 shell of exactly %d bytes", n)
	return nil
}

type existingPKCS12Shell struct {
	Version  int
	AuthSafe shellContentInfo
}

type shellContentInfo struct {
	ContentType asn1.ObjectIdentifier
	Content     asn1.RawValue `asn1:"tag:0,explicit"`
}
