module github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn

go 1.27

toolchain go1.27.3

require (
	github.com/pavlo-v-chernykh/keystore-go/v4 v4.5.0
	gopkg.in/yaml.v3 v3.0.1
	software.sslmate.com/src/go-pkcs12 v0.7.4
)

require (
	golang.org/x/crypto v0.58.0 // indirect
	golang.org/x/mod v0.38.0 // indirect
	golang.org/x/sync v0.22.0 // indirect
	golang.org/x/sys v0.49.0 // indirect
	golang.org/x/telemetry v0.0.0-20260708182218-49f421fb7959 // indirect
	golang.org/x/tools v0.48.0 // indirect
	golang.org/x/vuln v1.6.0 // indirect
)

tool golang.org/x/vuln/cmd/govulncheck
