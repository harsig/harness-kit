# harness-kit

Shared Go packages for [plain-harness](https://github.com/harsig/plain-harness), [scripts-mcp](https://github.com/harsig/scripts-mcp) and [skills-mcp](https://github.com/harsig/skills-mcp). It holds only code that more than one of them needs; anything used by a single tool stays in that tool.

| Package | Purpose |
| --- | --- |
| `gitsource` | A git source (`url`, `ref`) from a config file: its clone folder, `Require` (is it cloned? never uses the network) and `Sync` (clone or fetch, then check out the ref). |
| `yamlconfig` | Strict YAML loading (unknown keys are an error) and `ResolvePath` (relative to the config file, `~` expanded). |
| `text` | `Truncate`: cut text for a model at a byte limit, on a character boundary, with a marker. |

## Conventions shared by the tools

- Git sources are cloned by `<tool> --setup` into `.<tool>/repos/<name>-<hash>/` next to the config file. The hash covers URL and `ref`, so a changed `ref` needs `--setup` again. A normal run never uses the network.
- `ref` is a tag, branch or commit; empty means the remote's default branch.
- Relative paths in a config file are resolved from the file's directory.

## Development

```sh
go vet ./... && go test ./...
```

Until releases are tagged, the tools require a pseudo-version of `main`. After pushing a change here, update a tool with:

```sh
cd src && go get github.com/harsig/harness-kit@main && go mod tidy
```

To work on the kit and a tool at the same time, use a Go workspace in the folder that holds all repositories (it is not part of any repository and is not committed):

```sh
go work init ./harness-kit ./plain-harness/src ./scripts-mcp/src ./skills-mcp/src
```
