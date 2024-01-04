#!/bin/bash

export $(cat .env | xargs)
goreleaser release --release-notes /tmp/release-notes.md --clean

./scripts/post-release.sh
