# Developing

## Installation

On macOS or Linux run the following:

```
curl -sfL https://raw.githubusercontent.com/serverless/event-gateway/master/install.sh | sh
```

On Windows download [binary](https://github.com/serverless/event-gateway/releases).


## Running Locally

Run `event-gateway` in `dev` mode:

```
event-gateway -dev
```

### Register a Function

Register an AWS Lambda function in the Function Discovery.

```
curl --request POST \
  --url http://127.0.0.1:4001/v1/functions \
  --header 'content-type: application/json' \
  --data '{"functionId": "hello", "provider":{"type": "awslambda", "arn": "<Function AWS ARN>", "region": "<Region>", "accessKeyId": "<Access Key ID>", "secretAccessKey": "<Secret Access Key>"}}'
```

### Subscribe to an Event

Once the function is register you can subscribe it to you custom event.

```
curl --request POST \
  --url http://127.0.0.1:4001/v1/subscriptions \
  --header 'content-type: application/json' \
  --data '{"functionId": "hello", "event": "pageVisited"}'
```

### Emit an Event

An event can be emitted using [Events API](#events-api).

```
curl --request POST \
  --url http://127.0.0.1:4000/ \
  --header 'content-type: application/json' \
  --header 'event: pageVisited' \
  --data '{"userId": "123"}'
```

After emitting the event subscribed function is called asynchronously.
