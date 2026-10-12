# sicli

`sicli` is a Go monorepo for service CLIs.

MCP services use the pinned [mcp2cli](https://github.com/mcp2cli/source-code) runtime
shipped in their release archive. Users install one command per service.

## Structure

- `cmd/` contains release command entrypoints (`sicli`, `hast`, `paseo-bridge`).
- `internal/hastrt/` exposes Hast MCP tools, resources and prompts with Bearer auth and write guards.
- `internal/paseobridgert/` binds Paseo Bridge OpenAPI 3.1, scoped Bearer auth, E2EE, and send safety to Restish.
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

## Paseo Bridge (`paseo-bridge`)

Commands are generated from `<base-url>/api/v1/openapi.json`; the default origin
is `https://paseo-bridge.kahub.in`.

```sh
paseo-bridge auth login --mode token
paseo-bridge auth status
paseo-bridge status list-host-status -o json
paseo-bridge agents list-agents --host HOST -o json
PASEO_BRIDGE_WRITE_MODE=write paseo-bridge send send-message HOST AGENT 'prompt: hello'
```

Config is stored at `$XDG_CONFIG_HOME/sicli/paseo-bridge/config.toml`, mode `0600`.
Legacy `~/.config/aicli/paseo-bridge/` trees are migrated automatically on first run.

## Build and verify

```sh
just verify
just test-hast
just test-paseo-bridge
just build-hast
```

Release binaries are static (`CGO_ENABLED=0`) and built with stripped symbols.
See [THIRD_PARTY_NOTICES.md](THIRD_PARTY_NOTICES.md) for embedded dependency notices.
