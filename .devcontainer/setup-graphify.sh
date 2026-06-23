#!/usr/bin/env bash
set -euo pipefail

# Install the Codex skill in the repository (not in the persistent /root volume)
# and create the initial local graph. Both commands are incremental on later runs.
graphify install --project --platform codex
graphify extract .
