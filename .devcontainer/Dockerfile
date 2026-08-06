FROM mcr.microsoft.com/devcontainers/go:1.25

# Air (live reload Go)
RUN go install github.com/air-verse/air@latest

# Node (LTS) and AI coding CLIs.
# pnpm is provided through corepack, which pins it to the version declared in
# web/package.json ("packageManager") and verifies its integrity on download.
# That download is done here, at build time: otherwise corepack fetches pnpm on
# first use and blocks on an interactive confirmation prompt.
# Keep PNPM_VERSION in sync with "packageManager" in web/package.json.
#
# Since npm 12, dependency install scripts are blocked unless the package is
# named in --allow-scripts (same intent as `allowBuilds` in pnpm-workspace.yaml).
# claude-code needs its postinstall: it links the platform-native binary
# shipped as an optional dependency into bin/. Without it the CLI is installed
# but exits with "claude native binary not installed". Keep this list minimal.
ARG NODE_VERSION=24
ARG PNPM_VERSION=11.20.0
ENV COREPACK_ENABLE_DOWNLOAD_PROMPT=0
RUN curl -fsSL https://deb.nodesource.com/setup_${NODE_VERSION}.x | bash - \
    && apt-get install -y nodejs \
    && corepack enable pnpm \
    && corepack prepare pnpm@${PNPM_VERSION} --activate \
    && npm install -g npm@latest \
    && npm install -g --allow-scripts=@anthropic-ai/claude-code \
    @openai/codex @anthropic-ai/claude-code

# Antigravity CLI. Keep the executable outside /root: /root is a persistent
# Docker volume so authentication data never lands in the repository.
RUN curl -fsSL https://antigravity.google/cli/install.sh | bash -s -- --dir /usr/local/bin

RUN apt-get update \
    && apt-get install -y --no-install-recommends jq sqlite3 \
    && rm -rf /var/lib/apt/lists/*

# Goreleaser (release automation).
# Its postinstall downloads the platform binary; without --allow-scripts the
# download is deferred to the first `goreleaser` call in every new container.
RUN npm install -g --allow-scripts=@goreleaser/goreleaser @goreleaser/goreleaser
