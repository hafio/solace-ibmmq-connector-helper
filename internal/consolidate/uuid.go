package consolidate

import (
	"crypto/sha1"
	"encoding/hex"
	"strings"
)

// DurableNamespace is the fixed UUIDv5 namespace for MQ topic-consumer durable
// subscription names. It is pinned so durable names stay stable across releases.
const DurableNamespace = "6ba7f4e2-9c1d-5a3b-8e47-2f9a0c7d13e5"

// durableUnitSep joins the durable-name key fields (ASCII unit separator).
const durableUnitSep = "\x1f"

// DurableName returns the auto-generated durable-subscription-name for an MQ
// topic consumer: "solmq-" + UUIDv5(NS, connName ‖ queueManager ‖ topic ‖ basename).
func DurableName(connName, queueManager, topic, basename string) string {
	ns := mustParseUUID(DurableNamespace)
	key := strings.Join([]string{connName, queueManager, topic, basename}, durableUnitSep)
	return "solmq-" + uuidv5(ns, key)
}

// uuidv5 implements RFC 4122 version 5 (SHA-1) name-based UUIDs.
func uuidv5(ns [16]byte, name string) string {
	h := sha1.New()
	h.Write(ns[:])
	h.Write([]byte(name))
	sum := h.Sum(nil)
	var u [16]byte
	copy(u[:], sum[:16])
	u[6] = (u[6] & 0x0f) | 0x50 // version 5
	u[8] = (u[8] & 0x3f) | 0x80 // RFC 4122 variant
	return hex.EncodeToString(u[0:4]) + "-" +
		hex.EncodeToString(u[4:6]) + "-" +
		hex.EncodeToString(u[6:8]) + "-" +
		hex.EncodeToString(u[8:10]) + "-" +
		hex.EncodeToString(u[10:16])
}

// mustParseUUID parses the canonical hyphenated form into 16 bytes. It panics on
// a malformed constant, which can only be a programming error since the only
// caller passes DurableNamespace.
func mustParseUUID(s string) [16]byte {
	b, err := hex.DecodeString(strings.ReplaceAll(s, "-", ""))
	if err != nil || len(b) != 16 {
		panic("invalid UUID constant: " + s)
	}
	var u [16]byte
	copy(u[:], b)
	return u
}
