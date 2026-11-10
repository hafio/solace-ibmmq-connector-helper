package main

import "github.com/hafio/go-deploy-utility/notice"

// This file holds the one wording of the support statement: solmq-conn-util is
// not a supported Solace product -- Solace Professional Services built it and
// support it, and Solace Support does not. The wording is the shared module's
// (github.com/hafio/go-deploy-utility/notice); this tool supplies only its name
// and what the notice does not cover. Every surface that carries the statement
// takes it from here or is gated against it:
//
//   - docs/commands.md and docs/abbreviation.md emit supportNotice.Lines()
//     verbatim (renderCommandsDoc, renderAbbreviationDoc);
//   - README.md and the hand-written docs under docs/ (getting-started,
//     configuration, operations, troubleshooting, parity, DESIGN,
//     DEVELOPMENT) are listed in supportNoticeDocs, so
//     TestSupportNoticeInHandWrittenDocs requires the same block verbatim at
//     the top of each;
//   - solmq-conn-util-generator.html carries it as HTML, which
//     TestSupportNoticeInGeneratorPage compares against supportNotice.Plain();
//   - `version` prints notice.Short() under its first line.
//
// Terminal help deliberately does not carry it.

// supportNotice is the support statement as a GitHub [!WARNING] alert. The
// covers clause is load-bearing: without it the notice reads as if the Solace
// PubSub+ Connector for IBM MQ itself -- a Solace product -- were unsupported.
var supportNotice = mustNotice("the Solace PubSub+ Connector for IBM MQ that it configures and deploys")

// mustNotice builds the tool's notice. A rejected clause is a wiring bug in the
// declaration above, so it panics at start-up rather than printing no notice.
func mustNotice(covers string) notice.Notice {
	n, err := notice.New("solmq-conn-util", "tool", covers)
	if err != nil {
		panic("support.go: the support notice declaration is invalid: " + err.Error())
	}
	return n
}
