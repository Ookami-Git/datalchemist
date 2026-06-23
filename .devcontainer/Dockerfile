FROM mcr.microsoft.com/devcontainers/go:1.25

# Graphify builds a local, queryable codebase graph for AI coding assistants.
# pipx keeps its Python dependencies isolated from the Go/Node development tools.
ENV PIPX_HOME=/opt/pipx \
    PIPX_BIN_DIR=/usr/local/bin

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
    && apt-get install -y --no-install-recommends jq pipx python3-venv sqlite3 \
    && rm -rf /var/lib/apt/lists/*

# The published package is named graphifyy (the CLI remains `graphify`). Pin the
# version so rebuilding the devcontainer is reproducible.
RUN pipx install "graphifyy==0.8.46"

# Goreleaser (release automation)
RUN npm install -g @goreleaser/goreleaser
