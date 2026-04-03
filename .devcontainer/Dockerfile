FROM mcr.microsoft.com/devcontainers/go:1.25

# Air (live reload Go)
RUN go install github.com/air-verse/air@latest

# Node (LTS)
ARG NODE_VERSION=24
RUN curl -fsSL https://deb.nodesource.com/setup_${NODE_VERSION}.x | bash - \
    && apt-get install -y nodejs \
    && npm install -g npm@latest

RUN apt-get install -y jq sqlite3

# Goreleaser (release automation)
RUN npm install -g @goreleaser/goreleaser