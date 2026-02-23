#!/usr/bin/env bash
# Source this file to load Atlas environment variables:
#   source ./load-env.sh

set -a
source "$(dirname "${BASH_SOURCE[0]}")/.env"
set +a

echo "Atlas env loaded (ATLAS_IMAP_HOST=$ATLAS_IMAP_HOST)"
