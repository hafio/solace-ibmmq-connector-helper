package spec

import "testing"

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
