#!/bin/bash
HASH=`git log --pretty=format:'%h' -n 1`
HASHURL="s3://eg-binaries/${HASH}/"
MASTERURL="s3://eg-binaries/master/"

echo $HASHURL
aws s3 cp \
  dist/ $HASHURL \
  --acl public-read \
  --exclude "*.txt" \
  --exclude "config.yaml" \
  --recursive

aws s3 sync \
  $HASHURL $MASTERURL \
  --acl public-read
