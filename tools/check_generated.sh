#!/bin/bash

go generate ./...
go mod vendor
go mod tidy
if [[ $(git status --porcelain | wc -l) -ne 0 ]]; then
    echo "Generating code produced a non-empty diff"
    echo "Try running 'go generate ./... && go mod vendor && go mod tidy' then re-pushing."
    git status --porcelain
    git diff | cat
    exit 1;
fi