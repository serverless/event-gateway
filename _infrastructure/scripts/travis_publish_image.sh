#! /bin/bash

# Publish Docker image to ECR

set -e

if [ -z "$TRAVIS_PULL_REQUEST" ] || [ "$TRAVIS_PULL_REQUEST" == "false" ]; then
  # Push only if we're testing the master or prod branch
  if [ "$TRAVIS_BRANCH" == "master" ]; then
    # Compile
    GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o bin/$NAME

    # This is needed to login on AWS and push the image on ECR
    pip install --user awscli > /dev/null
    export PATH=$PATH:$HOME/.local/bin
    eval $(aws ecr get-login --region $AWS_DEFAULT_REGION)

    # Build
    docker build -t $NAME .

    # Push latest
    docker tag $NAME:latest "$REMOTE_IMAGE_URL:latest"
    docker push "$REMOTE_IMAGE_URL:latest"
    echo "Pushed $REMOTE_IMAGE_URL:latest"

    # Push tagged with commit ID
    docker tag $NAME:latest "$REMOTE_IMAGE_URL:$TRAVIS_COMMIT"
    docker push "$REMOTE_IMAGE_URL:$TRAVIS_COMMIT"
    echo "Pushed $REMOTE_IMAGE_URL:$TRAVIS_COMMIT"
  else
    echo "Skipping publish because branch is not 'master'"
  fi
else
  echo "Skipping publish because it's a pull request"
fi