package consolidate

import "testing"

// Fixed regression vector for the DurableName UUIDv5 algorithm. These inputs are a
// stable, self-contained pin (not tied to any shipped example): the value MUST NOT
// change, or existing MQ-topic durable subscriptions would be silently orphaned.
func TestDurableNameGolden(t *testing.T) {
	got := DurableName("mqhost.internal(1414)", "QM1", "MQTOPIC/EVENTS", "20-events.yaml")
	want := "solmq-3631c883-c0c4-5bc8-985e-ea2842831ad6"
	if got != want {
		t.Fatalf("durable name mismatch:\n got %q\nwant %q", got, want)
	}
}

func TestDurableNameDeterministic(t *testing.T) {
	a := DurableName("h(1414)", "QM", "T", "f.yaml")
	b := DurableName("h(1414)", "QM", "T", "f.yaml")
	if a != b {
		t.Fatalf("not deterministic: %q vs %q", a, b)
	}
	// Renaming the file must change the durable name (documented tradeoff).
	if c := DurableName("h(1414)", "QM", "T", "g.yaml"); a == c {
		t.Fatalf("expected file name to affect durable name, both %q", a)
	}
}
