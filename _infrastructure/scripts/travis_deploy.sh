#! /bin/bash

# Deploy published Docker image to ECS

set -e

if [ -z "$TRAVIS_PULL_REQUEST" ] || [ "$TRAVIS_PULL_REQUEST" == "false" ]; then
  # Deploy only if we're testing the master or prod branch
  if [ "$TRAVIS_BRANCH" == "master" ]; then
    curl -o ecs-cli https://s3.amazonaws.com/amazon-ecs-cli/ecs-cli-linux-amd64-v0.5.0
    chmod +x $TRAVIS_BUILD_DIR/ecs-cli

    $TRAVIS_BUILD_DIR/ecs-cli configure --cluster backend
    $TRAVIS_BUILD_DIR/ecs-cli compose service up --target-group-arn arn:aws:elasticloadbalancing:us-east-1:377024778620:targetgroup/gateway/1c51fea8fd2329be --container-name gateway --container-port 8080 --role gateway-dev-us-east-1-ecs-service
  else
    echo "Skipping deploy because branch is not 'master' not 'prod'"
  fi
else
  echo "Skipping deploy because it's a pull request"
fi