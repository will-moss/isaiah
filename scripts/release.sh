#!/bin/bash

export $(cat .env | xargs)
docker context use default
goreleaser release --release-notes /tmp/release-notes.md --clean

./scripts/post-release.sh
