set shell := ["bash", "-eo", "pipefail", "-c"]

version := env_var_or_default("VERSION", "0.1.0")
commit := env_var_or_default("COMMIT", `git rev-parse --short=12 HEAD 2>/dev/null || echo unknown`)
out_dir := env_var_or_default("OUT_DIR", "dist")
mcp2cli_version := "0.1.10"

default:
    @just --list

verify:
    go mod verify
    go test ./...
    go test -race ./...
    go vet ./...

test:
    go test ./...

test-hast:
    go test ./internal/hastrt ./cmd/hast

build-hast: _validate-version
    @just build-archive hast "$(go env GOOS)" "$(go env GOARCH)"
    tar -xzf "{{ out_dir }}/hast_{{ version }}_$(go env GOOS)_$(go env GOARCH).tar.gz" -C "{{ out_dir }}" hast

hast-mcp-check:
    "{{ out_dir }}/hast" ls
    "{{ out_dir }}/hast" get_models --json

release-build: _validate-version
    @just build-archive sicli linux amd64
    @just build-archive sicli linux arm64
    @just build-archive sicli darwin amd64
    @just build-archive sicli darwin arm64
    @just build-archive sicli windows amd64
    @just build-archive sicli windows arm64
    @just build-archive hast linux amd64
    @just build-archive hast linux arm64
    @just build-archive hast darwin amd64
    @just build-archive hast darwin arm64
    @just release-checksums
    @echo "release artifacts in {{ out_dir }}"
    @ls -lh "{{ out_dir }}"

build-archive command goos goarch: _validate-version
    @just _build-archive "{{ command }}" "{{ goos }}" "{{ goarch }}"

release-checksums: _validate-version
    #!/usr/bin/env bash
    set -euo pipefail
    cd "{{ out_dir }}"
    archive_count="$(find . -maxdepth 1 -type f \( -name '*_{{ version }}_*.tar.gz' -o -name '*_{{ version }}_*.zip' \) | wc -l)"
    if [[ "$archive_count" -ne 10 ]]; then
      echo "expected 10 release archives, found ${archive_count}" >&2
      exit 1
    fi
    sha256sum ./*_"{{ version }}"_*.tar.gz ./*_"{{ version }}"_*.zip > checksums.txt

_validate-version:
    @if [[ ! "{{ version }}" =~ ^[0-9]+\.[0-9]+\.[0-9]+([+-][0-9A-Za-z.-]+)?$ ]]; then echo "VERSION must be a semantic version without the leading v: {{ version }}" >&2; exit 64; fi

_build-archive command goos goarch:
    #!/usr/bin/env bash
    set -euo pipefail
    command="{{ command }}"
    goos="{{ goos }}"
    goarch="{{ goarch }}"
    binary="$command"
    archive_base="${command}_{{ version }}_${goos}_${goarch}"
    ldflags="-s -w -buildid="
    mkdir -p "{{ out_dir }}"
    output_dir="$(cd "{{ out_dir }}" && pwd)"

    if [[ "$goos" == "windows" ]]; then
      binary="${binary}.exe"
    fi
    if [[ "$command" == "hast" ]]; then
      ldflags+=" -X main.version={{ version }} -X main.commit={{ commit }}"
    fi

    stage="$(mktemp -d)"
    echo "building ${archive_base}"
    if [[ "$command" == "hast" ]]; then
      case "${goos}/${goarch}" in
        linux/amd64) target=x86_64-unknown-linux-musl; expected=a0e53b0685638b0263415fa891d5884f52419b24579dc513fd2db1c11327735e ;;
        linux/arm64) target=aarch64-unknown-linux-musl; expected=0fd984d7398d1d6db8b583d63a33dbdfaab3c928c59848299addfdfd50121b6c ;;
        darwin/amd64) target=x86_64-apple-darwin; expected=57df130f2b09762690bb856364cde61bb93ec512b46a2c0147b0527d7675833c ;;
        darwin/arm64) target=aarch64-apple-darwin; expected=f292de88842ef767f71b967d13276cd083ca3a70a08af1a9870ee102dca3792d ;;
        *) echo "mcp2cli {{ mcp2cli_version }} is unavailable for ${goos}/${goarch}" >&2; exit 64 ;;
      esac
      runtime_archive="${stage}/mcp2cli.tar.xz"
      curl --fail --silent --show-error --location \
        "https://github.com/mcp2cli/source-code/releases/download/v{{ mcp2cli_version }}/mcp2cli-{{ mcp2cli_version }}-${target}.tar.xz" \
        --output "$runtime_archive"
      printf '%s  %s\n' "$expected" "$runtime_archive" | sha256sum -c -
      tar -xJf "$runtime_archive" -C "$stage" --strip-components=1 \
        "mcp2cli-{{ mcp2cli_version }}-${target}/mcp2cli" \
        "mcp2cli-{{ mcp2cli_version }}-${target}/LICENSE"
      runtime_sha="$(sha256sum "${stage}/mcp2cli" | awk '{print $1}')"
      trap 'rm -f internal/mcp2clirt/runtime.gz' EXIT
      gzip -n -9 -c "${stage}/mcp2cli" > internal/mcp2clirt/runtime.gz
      mv "${stage}/LICENSE" "${stage}/MCP2CLI_LICENSE"
      rm "$runtime_archive"
      rm "${stage}/mcp2cli"
      ldflags+=" -X github.com/fewensa/sicli/internal/mcp2clirt.runtimeSHA256=${runtime_sha}"
      CGO_ENABLED=0 GOOS="$goos" GOARCH="$goarch" \
        go build -tags embedded_runtime -trimpath -buildvcs=false -ldflags "$ldflags" \
        -o "${stage}/${binary}" "./cmd/${command}"
      rm internal/mcp2clirt/runtime.gz
      trap - EXIT
    else
      CGO_ENABLED=0 GOOS="$goos" GOARCH="$goarch" \
        go build -trimpath -buildvcs=false -ldflags "$ldflags" \
        -o "${stage}/${binary}" "./cmd/${command}"
    fi

    cp LICENSE THIRD_PARTY_NOTICES.md "$stage/"
    chmod 0755 "${stage}/${binary}"
    chmod 0644 "${stage}/LICENSE" "${stage}/THIRD_PARTY_NOTICES.md"

    if [[ "$goos" == "windows" ]]; then
      (
        cd "$stage"
        zip -q "${output_dir}/${archive_base}.zip" "$binary" LICENSE THIRD_PARTY_NOTICES.md
      )
    else
      files=("$binary" LICENSE THIRD_PARTY_NOTICES.md)
      if [[ "$command" == "hast" ]]; then files+=(MCP2CLI_LICENSE); fi
      tar -C "$stage" -czf "${output_dir}/${archive_base}.tar.gz" "${files[@]}"
    fi
    rm "${stage}/${binary}"
    rm "${stage}/LICENSE"
    rm "${stage}/THIRD_PARTY_NOTICES.md"
    if [[ "$command" == "hast" ]]; then
      rm "${stage}/MCP2CLI_LICENSE"
    fi
    rmdir "$stage"
