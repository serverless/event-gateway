# Running Locally

## Installation (via GitHub releases)

On macOS or Linux run the following:

```
curl -sfL https://raw.githubusercontent.com/serverless/event-gateway/master/install.sh | sh
```

On Windows download [binary](https://github.com/serverless/event-gateway/releases).

## Installation (via Docker)

There is a [official Docker image](https://hub.docker.com/r/serverless/event-gateway/).

```
docker pull serverless/event-gateway
```

## Running Locally

Run `event-gateway` in `dev` mode:

```
event-gateway -dev
```

## Running in Docker

```
docker run -p 4000:4000 -p 4001:4001 serverless/event-gateway -dev
```

**Pass in your AWS credentials**

Mounts the `~/.aws` folder from the host to the `/home/.aws` folder inside the container. Event Gateway can then read the credentials from within the container.
```
docker run -p 4000:4000 -p 4001:4001 -e "HOME=/home" -v ~/.aws:/home/.aws event-gateway -dev
```

**Preserve state of etcd**

While testing if you restart the container running Event Gateway and want to preserve the data in etcd, you can specify a data dir with the `-embed-data-dir "/home/data"` flag specifying a destination folder. Then you can mount the folder `~/.event-gateway/data` from your host into the container at `/home/data`. Event Gateway will read the data from there.

```
docker run -p 4000:4000 -p 4001:4001 -v ~/.event-gateway/data:/home/data event-gateway -embed-data-dir "/home/data" -dev
```

## Operations

* [Register a Function](#register-a-function)
* [Subscribe to an Event](#subscribe-to-an-event)
* [Emit an Event](#emit-an-event)

### Register a Function

Register an AWS Lambda function in the Function Discovery.

```
curl --request POST \
  --url http://127.0.0.1:4001/v1/spaces/default/functions \
  --header 'content-type: application/json' \
  --data '{"functionId": "hello", "provider":{"type": "awslambda", "arn": "<Function AWS ARN>", "region": "<Region>", "accessKeyId": "<Access Key ID>", "secretAccessKey": "<Secret Access Key>"}}'
```

### Subscribe to an Event

Once the function is register you can subscribe it to you custom event.

```
curl --request POST \
  --url http://127.0.0.1:4001/v1/spaces/default/subscriptions \
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
