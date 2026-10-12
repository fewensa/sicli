# sicli

`sicli` is a Go monorepo for service CLIs.

MCP services use the pinned [mcp2cli](https://github.com/mcp2cli/source-code) runtime
shipped in their release archive. Users install one command per service.

## Structure

- `cmd/` contains release command entrypoints (`sicli`, `hast`).
- `internal/hastrt/` exposes Hast MCP tools, resources and prompts with Bearer auth and write guards.
- `internal/cli/` contains JSON helpers used by the `sicli` registry command.
- `services/` contains service registrations and command-surface metadata.

## Hast AIP MCP (`hast`)

`hast` discovers the tools, resources and prompts available to the selected key at
`https://backend.hast.so/api/public/v1/mcp`, following the
[MCP guide](https://hast.so/zh-CN/docs/mcp). Saved AIP Agents are represented by
Webhooks; `post_webhooks` creates one. Tool availability depends on the key's
Profile and scopes. Agent configuration requires `webhooks:read`; creating,
updating and invoking Agents requires `webhooks:write`.

```sh
hast auth login --mode key
hast auth status
hast context list
hast --context work auth login --mode key
hast ls
hast --help
hast get_models --json
```

Credentials remain at `$XDG_CONFIG_HOME/sicli/hast/config.toml`, mode `0600`.

## Build and verify

```sh
just verify
just test-hast
just build-hast
```

Release binaries are static (`CGO_ENABLED=0`) and built with stripped symbols.
See [THIRD_PARTY_NOTICES.md](THIRD_PARTY_NOTICES.md) for embedded dependency notices.
