# Event Types

Event Registry is a single source of truth about events occuring in the space. Every event emitted to a space has to have event
type registered beforehand. Event type can be registered using [Event Types Configuration API](./api.md#event-types).

## Authorization

Event Type can define authorizer function that will be called before calling a subscribed function. Authorizer function is a function registered in Event Gateway beforehand.

### Invocation Payload

The authorizer function is invoked with a special payload. The function has access to the whole request object because different parts of the request can be required for running authorization logic (e.g. API key can be stored in different headers or query params). The invocation payload is a JSON object with the following structure:

- `event` - `object` - event received and parsed by EG
- `request` - `object` - original HTTP request to the EG, this field is exactly the same as HTTP event, including body, which in case of CloudEvent will be exactly the same as event field

### Invocation Result

The authorize function is expected to return authorization response JSON object with the following structure:

- `authorization` - `object` - object containing authorization data, required if authorization is successful. Fields:
  - `principalId` - `string` - required if authorization is successful, the principal user identification associated with the token sent in the request
  - `context` - `map[string]string` - arbitrary data that will be accessible by the downstream function
- `error` - `object` - error information, required if authorization is unsuccessful. Fields:
  - `message` - `string` - authorization error message

If the authorization is successful `error` field has to be null or not defined otherwise Event Gateway treats authorization process as unsuccessful and `403 Forbidden` error is returned to the client assuming that there was sync subscription defined. `authorization` object is attached to CloudEvent Extensions field under `eventgateway.authorization` key.

If Event Gateway will not be able to parse the response from authorizer function or there will be error during invocation authorization process is considered as unsuccessful.