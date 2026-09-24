package main

// This file holds the one wording of the support statement: solmq-conn-util is
// not a supported Solace product -- Solace Professional Services built it and
// support it, and Solace Support does not. Every surface that carries the
// statement takes it from here or is gated against it:
//
//   - docs/commands.md and docs/abbreviation.md emit supportNoticeMarkdown
//     verbatim (renderCommandsDoc, renderAbbreviationDoc);
//   - README.md, docs/userguide.md and docs/DEVELOPMENT.md are hand-written, so
//     TestSupportNoticeInHandWrittenDocs requires the same block verbatim at
//     the top of each;
//   - solmq-conn-util-generator.html carries it as HTML, which
//     TestSupportNoticeInGeneratorPage compares against this block as plain
//     text;
//   - `version` prints supportNoticeVersion under its first line.
//
// Terminal help deliberately does not carry it.

// supportNoticeMarkdown is the support statement as a GitHub [!WARNING] alert,
// one element per line. The last sentence is load-bearing: without it the
// notice reads as if the Solace PubSub+ Connector for IBM MQ itself -- a Solace
// product -- were unsupported.
var supportNoticeMarkdown = []string{
	"> [!WARNING]",
	"> **Not a supported Solace product.** " + bt + "solmq-conn-util" + bt + " was created by Solace",
	"> Professional Services and is supported only by Solace Professional Services --",
	"> not by Solace Support. For help with this tool, contact your Solace",
	"> Professional Services representative rather than opening a Solace Support",
	"> case. This notice covers this tool only, not the Solace PubSub+ Connector for",
	"> IBM MQ that it configures and deploys.",
}

// supportNoticeVersion is the short form `version` prints under its first line.
// That first line is left exactly as it was, so anything reading the version
// off it (`| head -1`, a field split) is unaffected.
var supportNoticeVersion = []string{
	"Not a supported Solace product. Created by Solace Professional Services and",
	"supported only by Solace Professional Services -- not by Solace Support.",
}
