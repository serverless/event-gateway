# gateway
⚡️😍💸

## API

Gateway exposes configuration API for all sub-services.

### Function Discovery

#### Models

##### `Function`

- id - `string` - function id/name
- instance - `array` of objects:
  - provider - `string` - compute provider, possible values: `aws-lambda`
  - originId - `string` - provider specific function ID
  - region - `string` - deployment region
  - credentials - `object`:
    - aws_access_key_id - `string`
    - aws_secret_access_key - `string`

#### Methods

##### Register function

`POST /v0/gateway/api/function`

Request:

`Function` model object

Response:

`Function` model object

##### Get function

`GET /v0/gateway/api/function/:name`

Response:

`Function` model object

##### Invoke function

`POST /v0/gateway/api/invoke/:name`

Request:

`object` passed to the function

Response:

`object` with function invocation result

### Endpoints

#### Models

##### `Endpoint`

- id - `string` - endpoint id
- functions - `array` of objects:
  - functionId - `string` - ID of function registered in Function Discovery
  - method - `string` - HTTP method
  - path - `string` - URL path

#### Methods

##### Create endpoint

`POST /v0/gateway/api/endpoint`

Request:

`Endpoint` model object (without `id`)

Response:

`Endpoint` model object

##### Get endpoint

 `GET /v0/gateway/api/endpoint/:id`

 Response:

 `Endpoint` model object

 ##### Call endpoint (public API)

`<method registered in endpoint> /v0/gateway/endpoint/:id/<path registered in endpoint`

Request:

Payload expected by function

Response:

Payload returned by function
