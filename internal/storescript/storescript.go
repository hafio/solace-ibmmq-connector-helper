// Package storescript renders the startup script a docker container runs when
// it carries created TLS stores, and records the command the connector image
// starts with, which the script hands over to.
//
// Docker has no secret store for a binary file: a compose secret read from the
// environment carries text, and an environment variable cannot hold the NUL
// bytes every PKCS12 contains. So compose delivers each created store
// base64-encoded, the script decodes them into an in-memory directory on every
// start -- a restarted container decodes them again -- and then execs the
// image's own command. Like statusscript, it is a pure renderer: no os/exec,
// no filesystem, no globals. The script targets the image's busybox userland
// (sh, base64, mv).
package storescript

import (
	"github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn/internal/spec"
	"github.com/solacecommunity/hafio-solace/connectors/ibmmq/solmq-conn/internal/statusscript"
)

// Filename is the script's name inside the container. The leading dot keeps it
// out of the way of anything that lists the directory, like the status
// script's.
const Filename = ".stores-init"

// ContainerPath is the script's full path inside the container: beside the
// status script, outside every directory a volume mounts over.
const ContainerPath = statusscript.ContainerDir + "/" + Filename

// Shell is what compose's entrypoint runs the script with. A compose configs
// entry lands the script read-only without the execute bit, so it is passed
// to a shell rather than run.
const Shell = "/bin/sh"

// ImageEntrypoint and ImageCmd are the connector image's own ENTRYPOINT and
// CMD, the same in every release checked (2.13.0, 2.14.1, 3.1.0 and 3.2.1, read
// from the registry's image configs): Jib's java launch of the connector's main
// class, and the Spring config location this tool mounts application.yml
// under. Setting compose's entrypoint: also drops the image's CMD, so the
// script is handed both and execs them unchanged; deploy compares them with
// the image it finds locally before relying on them.
var (
	ImageEntrypoint = []string{"java", "-cp", "@/app/jib-classpath-file", "com.solace.connector.jms.JmsConnector"}
	ImageCmd        = []string{"--spring.config.additional-location=optional:file:/app/external/spring/config/,optional:file:/app/external/spring/config/*/"}
)

// Command is what compose's command: hands the script to exec: the image's
// entrypoint, then its command.
func Command() []string {
	out := append([]string(nil), ImageEntrypoint...)
	return append(out, ImageCmd...)
}

// Render returns the script. It decodes every file under
// spec.EncodedStoresMountPath into spec.DefaultStoresMountPath under the same
// name -- the path the rendered application.yml names each store by, on every
// platform -- writing through a temporary name so a store is never read half
// written, with a umask that keeps the decoded stores to the connector's own
// identity. set -e stops the container on a store that does not decode rather
// than starting the connector without it.
func Render() string {
	enc, dec := spec.EncodedStoresMountPath, spec.DefaultStoresMountPath
	return `#!/bin/sh
# solmq-conn-util: decode the created TLS stores compose delivers base64-encoded
# under ` + enc + ` into ` + dec + `,
# an in-memory tmpfs, then start the connector with the image's own command.
# Runs on every start, so a restarted container decodes them again.
set -eu
umask 077
for f in ` + enc + `/*; do
  [ -f "$f" ] || continue
  name=${f##*/}
  base64 -d "$f" > "` + dec + `/$name.tmp"
  mv "` + dec + `/$name.tmp" "` + dec + `/$name"
done
exec "$@"
`
}
