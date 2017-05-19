# gateway
⚡️😍💸

## API

Gateway exposes configuration API for all sub-services.

### Function discovery

#### Models

##### `Function`

- id - `string` - function id/name
- instance - `array` of objects:
  - provider - `string` - compute provider, possible values: `aws-lambda`
  - originId - `string` - provider specific function ID
  - region - `string` - deployment region

#### Methods

##### Register function

`/v0/api/function`

Request:

`Function` model object

Response:

`Function` model object

##### Get function

`/v0/api/function/:name`

Response:

`Function` model object

##### Invoke function

`/v0/api/invoke/:name`

Request:

`object` passed to the function

Response:

`object` with function invocation result