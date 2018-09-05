#!/usr/bin/env bash

set -e
echo "" > coverage.txt

for d in $(go list ./... | grep -v vendor); do
    go test -race -mod=vendor -coverprofile=profile.out -covermode=atomic $d
    if [ -f profile.out ]; then
        cat profile.out >> coverage.txt
        rm profile.out
    fi
done

# include coverage for hosted EG
go test -race -mod=vendor -coverprofile=profile.out -covermode=atomic -tags=hosted ./router
cat profile.out >> coverage.txt
rm profile.out