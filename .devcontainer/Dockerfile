FROM mcr.microsoft.com/devcontainers/go:1.25

# Air (live reload Go)
RUN go install github.com/air-verse/air@latest

# Node (LTS) and AI coding CLIs
ARG NODE_VERSION=24
RUN curl -fsSL https://deb.nodesource.com/setup_${NODE_VERSION}.x | bash - \
    && apt-get install -y nodejs \
    && npm install -g npm@latest \
    && npm install -g @openai/codex @anthropic-ai/claude-code

# Antigravity CLI. Keep the executable outside /root: /root is a persistent
# Docker volume so authentication data never lands in the repository.
RUN curl -fsSL https://antigravity.google/cli/install.sh | bash -s -- --dir /usr/local/bin

RUN apt-get update \
    && apt-get install -y --no-install-recommends jq sqlite3 \
    && rm -rf /var/lib/apt/lists/*

# Goreleaser (release automation)
RUN npm install -g @goreleaser/goreleaser
