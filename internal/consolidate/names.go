package consolidate

import (
	"fmt"
	"strings"

	"github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn/internal/spec"
	"github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn/internal/tls"
)

// Stable secret names for the credential positions that belong to env.yaml as a
// whole rather than to one binder. The two store passwords are re-exported from
// internal/tls, which names them for the api-properties path, so the SSL bundle
// and api-properties can never disagree about one password's name. The two
// leader-election names are the fallback for a management session no workflow
// shares -- see leaderNames in Build.
const (
	TruststorePasswordName = tls.TruststorePasswordName
	KeystorePasswordName   = tls.KeystorePasswordName
	LeaderUsernameName     = spec.GeneratedNamePrefix + "LEADER_ELECTION_CLIENT_USERNAME"
	LeaderPasswordName     = spec.GeneratedNamePrefix + "LEADER_ELECTION_CLIENT_PASSWORD"
)

// secretFn records a credential under a mount name and returns the placeholder
// the rendered config carries ("" for an unset credential). origin is the spec
// position the credential came from, used only to name both sides when two
// credentials collide on one name; secretRef appends "-env" to it itself, so
// callers pass the bare field label.
type secretFn func(stable, origin string, c spec.Cred) string

// leaderNameFn answers the stable names the management session's username and
// password are recorded under. Which pair those are depends on the binders
// already built, so Build supplies the closure rather than buildLeaderElection
// deciding for itself.
type leaderNameFn func(sess spec.Side) (user, pass string)

// storeSecret adapts a secretFn to the narrower callback internal/tls needs, so
// a store password mounted for the SSL bundle and the same password referenced
// from the Solace api-properties resolve to one name and one mounted file. The
// origin is recovered from the name here rather than widening tls's callback:
// the two store passwords are the only credentials that reach it, and each has
// exactly one spec position.
func storeSecret(secretRef secretFn) func(string, spec.Cred) string {
	return func(stable string, c spec.Cred) string {
		return secretRef(stable, storeOrigin(stable), c)
	}
}

// storeOrigin is the spec position behind a store password's fixed mount name.
func storeOrigin(stable string) string {
	if stable == KeystorePasswordName {
		return "tls.keystore.password"
	}
	return "tls.truststore.password"
}

// stableName is the in-container mount name for a binder-scoped credential that
// has no name of its own -- a literal. It carries spec.GeneratedNamePrefix to
// keep every derived name clear of the host variable names `-env` credentials
// are mounted under, which share the same namespace.
func stableName(binder, suffix string) string {
	return spec.GeneratedNamePrefix + stableToken(binder) + "_" + suffix
}

// stableToken folds an arbitrary name into the character set every consumer of a
// stable name accepts at once: an environment-variable identifier
// ([A-Za-z_][A-Za-z0-9_]*), which is also a valid file name, a valid Spring
// property segment, a valid Kubernetes Secret data key ([-._a-zA-Z0-9]+), and a
// valid podman secret name component. Anything outside it folds to '_', runs of
// '_' collapse, and a leading digit is prefixed so the result is never a bare
// number.
func stableToken(s string) string {
	var b strings.Builder
	prevUnderscore := false
	for _, r := range s {
		switch {
		case r >= 'a' && r <= 'z':
			b.WriteRune(r - 32) // upper-snake by convention
			prevUnderscore = false
		case (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9'):
			b.WriteRune(r)
			prevUnderscore = false
		default:
			if !prevUnderscore && b.Len() > 0 {
				b.WriteByte('_')
				prevUnderscore = true
			}
		}
	}
	out := strings.TrimRight(b.String(), "_")
	if out == "" {
		return "X"
	}
	if out[0] >= '0' && out[0] <= '9' {
		return "X" + out
	}
	return out
}

// securityUserPasswordName is the stable secret name for one management user's
// password, keyed by the user's name so adding a user never renumbers another.
func securityUserPasswordName(user string) string {
	return spec.GeneratedNamePrefix + "SECURITY_USER_" + stableToken(user) + "_PASSWORD"
}

// assignBinderNames maps each accumulated connection to its binder name: the
// contributing connection name when there is one, else a generated
// sol-conn-N / mq-conn-N, with a -2/-3 suffix disambiguating any clash between
// two different binders.
func assignBinderNames(accs []*acc) {
	used := map[string]int{}
	assign := func(base string) string {
		if used[base] == 0 {
			used[base] = 1
			return base
		}
		used[base]++
		return fmt.Sprintf("%s-%d", base, used[base])
	}
	var solN, mqN int
	for _, a := range accs {
		var base string
		switch {
		case a.connName != "":
			base = sanitize(a.connName)
		case a.kind == spec.SystemSolace:
			solN++
			base = fmt.Sprintf("sol-conn-%d", solN)
		default:
			mqN++
			base = fmt.Sprintf("mq-conn-%d", mqN)
		}
		a.name = assign(base)
	}
}
