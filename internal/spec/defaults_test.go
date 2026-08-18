package spec

import "testing"

func TestSecurityEffectivelyEnabled(t *testing.T) {
	tests := []struct {
		name string
		sec  Security
		want bool
	}{
		{"absent", Security{}, true},
		{"present and enabled", Security{Present: true, Enabled: true}, true},
		{"present and disabled", Security{Present: true, Enabled: false}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.sec.EffectivelyEnabled(); got != tt.want {
				t.Errorf("EffectivelyEnabled() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLeaderElectionEffectiveMode(t *testing.T) {
	tests := []struct {
		name string
		mode string
		want string
	}{
		{"empty", "", LeaderStandalone},
		{"standalone", LeaderStandalone, LeaderStandalone},
		{"active_active", LeaderActiveActive, LeaderActiveActive},
		{"active_standby", LeaderActiveStby, LeaderActiveStby},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			le := LeaderElection{Mode: tt.mode}
			if got := le.EffectiveMode(); got != tt.want {
				t.Errorf("EffectiveMode() = %q, want %q", got, tt.want)
			}
		})
	}
}
