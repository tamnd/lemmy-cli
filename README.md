# lemmy

A command line for lemmy.

`lemmy` is a single pure-Go binary. It reads public lemmy data
over plain HTTPS, shapes it into clean records, and prints output that pipes
into the rest of your tools. No API key, nothing to run alongside it.

The same package is also a [resource-URI driver](#use-it-as-a-resource-uri-driver),
so a host program like [ant](https://github.com/tamnd/ant) can address
lemmy as `lemmy://` URIs.

## Install

```bash
go install github.com/tamnd/lemmy-cli/cmd/lemmy@latest
```

Or grab a prebuilt binary from the [releases](https://github.com/tamnd/lemmy-cli/releases), or run
the container image:

```bash
docker run --rm ghcr.io/tamnd/lemmy:latest --help
```

## Usage

```bash
lemmy page <path>                      # fetch one page as a record
lemmy page <path> -o json              # as JSON, ready for jq
lemmy page <path> --template '{{.Body}}'  # just the readable body text
lemmy links <path>                     # the pages it links to, one per line
lemmy --help                           # the whole command tree
```

Every command shares one output contract: `-o table|json|jsonl|csv|tsv|url|raw`,
`--fields` to pick columns, `--template` for a custom line, and `-n` to limit.
The default adapts to where output goes (a table on a terminal, JSONL in a
pipe), so the same command reads well by hand and parses cleanly downstream.

This is a fresh scaffold. It ships one example resource type, `page`, wired end
to end. Model the real lemmy records in `lemmy/` and declare their
operations in `lemmy/domain.go`; each one becomes a command, an HTTP
route, and an MCP tool at once.

## Serve it

The same operations are available over HTTP and as an MCP tool set for agents,
with no extra code:

```bash
lemmy serve --addr :7777    # GET /v1/page/<path>  returns NDJSON
lemmy mcp                   # speak MCP over stdio
```

## Use it as a resource-URI driver

`lemmy` registers a `lemmy` domain the way a program registers a
database driver with `database/sql`. A host enables it with one blank import:

```go
import _ "github.com/tamnd/lemmy-cli/lemmy"
```

Then [ant](https://github.com/tamnd/ant) (or any program that links the package)
dereferences `lemmy://` URIs without knowing anything about lemmy:

```bash
ant get lemmy://page/<path>   # fetch the record
ant cat lemmy://page/<path>   # just the body text
ant ls  lemmy://page/<path>   # the pages it links to, each addressable
ant url lemmy://page/<path>   # the live https URL
```

## Development

```
cmd/lemmy/   thin main: hands cli.NewApp to kit.Run
cli/                 assembles the kit App from the lemmy domain
lemmy/                the library: HTTP client, data models, and domain.go (the driver)
docs/                tago documentation site
```

```bash
make build      # ./bin/lemmy
make test       # go test ./...
make vet        # go vet ./...
```

## Releasing

Push a version tag and GitHub Actions runs GoReleaser, which builds the
archives, Linux packages, the multi-arch GHCR image, checksums, SBOMs, and a
cosign signature:

```bash
git tag v0.1.0
git push --tags
```

The Homebrew and Scoop steps self-disable until their tokens exist, so the first
release works with no extra secrets.

## License

Apache-2.0. See [LICENSE](LICENSE).
