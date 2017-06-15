# Gateway

## API

Gateway exposes configuration RESTful HTTP API.

### Functions discovery

#### Register function

`POST /api/function`

Request:

- `functionId` - `string` - required, function name
- `type` - `string` - required, function type, possible values: `aws-lambda`, `gcloud-function`, `azure-function`, `openwhisk-action`, `group`, `http`. `group` type may not use another group as a backing function.
- `properties` - object:
  - for `aws-lambda`:
    - `arn` - `string` - AWS ARN identifier
    - `region` - `string` - region name
    - `version` - `string` - a specific version ID
    - `accesKeyID` - `string` - AWS API key ID
    - `secretAccesKey` - `string` - AWS API key
  - for `gcloud-functions`:
    - `name` - `string` - function name
    - `region` - `string` - region name
    - `serviceAccountKey` - `json` - Google Service Account key
  - for `azure-functions`:
    - `name` - `string` - function name
    - `appName` - `string` - azure app name
    - `azureFunctionsAdminKey` - `string` - Azure API key
  - for `openwhisk-action`:
    - `name` - `string` - action name
    - `namespace` - `string` - OpenWhisk namespace
    - `apiHost` - `string` - OpenWhisk platform endpoint, e.g. openwhisk.ng.bluemix.net
    - `auth` - `string` - OpenWhisk authentication key, e.g. xxxxxx:yyyyy
    - `apiGwAccessToken` - `string` - OpenWhisk optional API gateway access token
  - for `group`:
    - `functions` - `array` of `object` - backing functions
      - `functionId` - `string` - function ID
      - `weight` - `number` - proportion of requests destined to this function, defaulting to 1
  - for `http`:
    - `url` - `string` - the URL of an http or https remote endpoint

Response:

- `functionId` - `string` - function name
- `type` - `string` - required. function type, possible values: `aws-lambda`, `gcloud-function`, `azure-function`, `openwhisk-action`, `group`, `http`.
- `properties` - `object` - specific to `type`

#### Change configuration of group function

`PUT /api/function/<function ID>/functions`

Allows changing configuration of group function

Request:

- `functions` - `array` of `object` - backing functions
  - `functionId` - `string` - function ID
  - `weight` - `number` - proportion of requests destined to this function, defaulting to 1

Response:

- `functions` - `array` of `object` - backing functions
  - `functionId` - `string` - function ID
  - `weight` - `number` - proportion of requests destined to this function, defaulting to 1

#### Deregister function

`DELETE /api/function/<function id>`

Notes:
* used to delete all types of functions, including groups
* fails if the function ID is currently in-use by an endpoint or topic

### Endpoints

#### Create endpoint

`POST /api/endpoint`

Request:

- `functionId` - `string` - ID of backing function or function group
- `method` - `string` - HTTP method
- `path` - `string` - URL path

Response:

- `endpointId` - `string` - a short UUID that represents this endpoint mapping
- `functionId` - `string` - function ID
- `method` - `string` - HTTP method
- `path` - `string` - URL path

#### Delete endpoint

`DELETE /api/endpoint/<endpoint ID>`

#### Get endpoints

`GET /api/endpoint`

Response:

- `endpoints` - `array` of `object`
	- `id` - `string` - endpoint ID, which is method + path, e.g. `GET-/homepage`
	- `functionId` - `string` - function ID
	- `method` - HTTP method
	- `path` - URL path

### Pub/Sub

#### Create topic

`POST /api/topic`

Request:

- `id` - `string` - name of topic

Response:

- `id` - `string` - name of topic

#### Delete topic

`DELETE /api/topic/<topic id>`

#### Get topics

`GET /api/topic`

Response:

- `topics` - `array` of `object` - topics
  - `id` - `string` - topic name

#### Add subscription

`POST /api/topic/<topic id>/subscription`

Request:

- `functionId` - ID of function or function group to receive events from the topic

Response:

- `subscriptionId` - `string` - subscription ID, which is topic + function ID, e.g. `newusers-/userProcessGroup`
- `functionId` - ID of function or function group

#### Delete subscription

`DELETE /api/topic/<topic id>/subscription/<subscription id>`

#### Get subscriptions

`GET /api/topic/<topic id>/subscription`

Response:

- `subscriptions` - `array` of `object` - backing functions
  - `subscriptionId` - `string` - subscription ID
  - `functionId` - ID of function or function group

#### Add publisher

`POST /api/topic/<topic id>/publisher`

Request:

- `functionId` - ID of function or function group to publish events to the topic
- `type` - either `input` or `output`

Response:

- `publisherId` - `string` - publisher ID, which is topic + function ID, e.g. `newusers-/userCreateGroup`
- `functionId` - ID of function or function group to publish events to the topic
- `type` - either `input` or `output`

#### Delete publisher

`DELETE /api/topic/<topic id>/publisher/<publisher id>`

#### Get Publishers

`GET /api/topic/<topic id>/publisher`

Response:

- `publishers` - `array` of `object` - backing functions
  - `publisherId` - `string` - publisher ID
  - `functionId` - ID of function or function group
	- `type` - either `input` or `output`

#### Publish message to the topic

`POST /api/topic/<topic id>/publish`

Request: arbitrary payload