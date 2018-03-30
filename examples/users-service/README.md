# Using Event Gateway as an API Gateway

## Background

With the explosion of FaaS and Serverless architectures, applications are getting split up into smaller and smaller pieces. Rather than have a single bundle of code that handles all of your HTTP endpoints, you can create individual functions that handle a very specific task -- create a User, update my Shopping Cart, delete my Tweet.

This separation makes it easier for teams to move independently and ship more quickly, but it also presents a management problem. How do you make it easy for clients to access these hundreds of isolated functions?

Using the Event Gateway as an API Gateway solves this problem. You can have collections of functions split into services via the [Serverless Framework](https://github.com/serverless/serverless). These services can be deployed independently to the same domain in a way that's easy for clients to access.

In this example, we'll deploy an example Users service with a few different HTTP endpoints. We'll see how we can interact with those endpoints in a familiar, RESTful way.


## Deploying & Testing the Users Service

For our example Users service, let's say we want endpoints to create a new user, to get a specific user, and to delete a single user. We'll implement each of these endpoints in separate functions. Let's take a look at `index.js` where they're implemented:

```javascript
const faker = require('faker')

module.exports.get = (event, context, cb) => {
  console.log(event)

  cb(null, {
    headers: {
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({
      id: event.data.params.id,
      name: faker.name.findName(),
      email: faker.internet.email()
    })
  })
}

module.exports.post = (event, context, cb) => {
  console.log(event)
  cb(null, {
    headers: {
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({
      id: faker.random.number(),
      name: faker.name.findName(),
      email: faker.internet.email()
    })
  })
}

module.exports.delete = (event, context, cb) => {
  console.log(event)
  cb(null, {
    headers: {
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({
      message: `Deleted user with id ${event.data.params.id}`
    })
  })
}
```

We're exporting three functions -- `get`, `post`, and `delete` -- to handle our three endpoints. The logic is using the `faker` library to send out pretend data, but you would use some sort of data store in your actual application.

Next, let's see how we configure the functions to be triggered by our Event Gateway. In the `serverless.yml`, look at the `functions` block:

```yml
functions:
  getUser:
    handler: index.get
    events:
      - eventgateway:
          event: http
          method: GET
          path: /users/:id
  createUser:
    handler: index.post
    events:
      - eventgateway:
          event: http
          path: /users
          method: POST
  deleteUser:
    handler: index.delete
    events:
      - eventgateway:
          event: http
          path: /users/:id
          method: DELETE
```

For each function we've registered an `eventgateway` event of type `http`, indicating that it's an API Gateway event. We include the path and HTTP method to match on as well.

Finally, there are a few other relevant configuration sections of `serverless.yml`:

```yml
custom:
  eventgateway:
    space: <your space>
    apiKey: <your API key>

plugins:
  - "@serverless/serverless-event-gateway-plugin"
```

This `serverless-event-gateway-plugin` adds additional functionality to register functions and subscriptions with the Event Gateway. The `eventgateway` section of the `custom` block configures the plugin with your space and api key. You may choose any space you want, as long as it hasn't already been claimed, and it will be available at `https://<space>.slsgateway.com`.

Once you've entered your space and API key, install your dependencies and deploy:

```bash
$ npm install
$ sls deploy
```

Now let's give it a try! Using `curl` or in your browser, navigate to the `getUser` endpoint at `https://<space>.slsgateway.com/users/15`:

```bash
$ $ curl -X GET https://examplestest.slsgateway.com/users/15 | jq "."
{
  "id": "10",
  "name": "Ariel Dach",
  "email": "Aubree22@gmail.com"
}
```

It worked! You can try your `POST` and `DELETE` endpoints as well.

You can reuse this same space across multiple services using different endpoint paths to allow for agility across your team.
