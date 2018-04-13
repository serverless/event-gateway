# Using Custom Events

## Background

In a microservices architecture, you'll often have events in one service that are relevant to other services. In this example, we'll show you how to emit events from one service and subscribe to event from other services.

We'll use a common example -- the user creation flow. Your users may be created in the Users service, but the event when a user is created is important to multiple services. Perhaps you want to trigger a welcome email in your email service, post a celebratory message to Slack, or insert the user into your marketing team's CRM.

Let's focus on the last use case here. We'll emulate a CRM service that listens to events across our architecture and updates our CRM accordingly. We'll set up a subscription on the `user.created` event that's emitted from our Users service and insert the new user into our CRM.

## Deploying & Testing the CRM Service

First, let's set up our CRM service. I like to start with my function logic first, as my business logic is where I should be focusing. Take a look at `index.js`, which includes the logic:

```javascript
module.exports.handler = (event, context, cb) => {
  const user = event.data;
  console.log('Saving user to CRM: ');
  console.dir(user);

  // Add your CRM logic here
  // saveUserToCRM(user)

  cb(null, { message: 'success' })
}
```

This function is pretty basic as you'll need to implement the logic specific to your CRM. It accepts the event and creates a user object from the event data. Then it saves that user to my CRM. We'll just log the user object here rather than actually save it to a CRM.

Next, let's see how we configure that function to be triggered by our event. In the `serverless.yml`, look at the `functions` block:

```yml
functions:
  addUserToCRM:
    handler: index.handler
    events:
      - eventgateway:
          event: user.created
```

We've registered our function logic from `index.js`, and we've created an `eventgateway` subscription to the `user.created` event. Whenever that event is emitted into our Event Gateway space, this function will be triggered.

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

With your function deployed and subscription configured, it's time to test it out. You can use the plugin command `sls gateway emit` to emit a test event into your Event Gateway space. There's a sample event body in `event.json` to represent a test user.created event:

```bash
$ sls gateway emit --event "user.created" --data "$(cat event.json)"
Event emitted: user.created
Run `serverless logs -f <functionName>` to verify your subscribed function was triggered.
```

Note that we only get a message that an event was emitted. Custom events are processed asynchronously, and multiple functions can be subscribed to a custom event. This decouples producers and consumers, and it means a producer won't be able to tell who receives an event when it publishes it.

You can check your function logs to make sure everything is working properly:

```bash
$ sls logs -f addUserToCRM -t
START RequestId: 7b8800f5-2d34-11e8-8165-171c6f09ec0e Version: $LATEST
2018-03-21 18:20:07.365 (+00:00)	7b8800f5-2d34-11e8-8165-171c6f09ec0e	Saving user to CRM:
{ company: 'Test Corp, Inc.',
  email: 'test@testcorp.com',
  firstname: 'Test',
  lastname: 'Williams',
  username: 'test-user' }
END RequestId: 7b8800f5-2d34-11e8-8165-171c6f09ec0e
REPORT RequestId: 7b8800f5-2d34-11e8-8165-171c6f09ec0e	Duration: 5.77 ms	Billed Duration: 100 ms 	Memory Size: 1024 MB	Max Memory Used: 19 MB
```

As you can see, our function was triggered from our event and logged out the message as desired. Success! 

## Emitting Events in your Services

In the example above, we saw how to set up a consumer of events as well as how to publish test events via the command line. However, most events won't come from the command line -- they'll come from your application code, or webhooks, or database events. 

In this example, let's complete the story by emitting `user.created` events from our Users service. The easiest way to emit an event is to use the [`event-gateway-sdk`](https://github.com/serverless/event-gateway-sdk) for Node.

To use the SDK, you'll create an Event Gateway client configured for your space, then emit an event with the event name and payload to send.

The `emit.js` file shows how you could use the SDK in your application:

```javascript
// emit.js

const SDK = require('@serverless/event-gateway-sdk');
const SPACE = 'examplestest';

const eventGateway = new SDK({
  url: `https://${SPACE}.slsgateway.com`,
  space: `${SPACE}`
})

function createUser(user) {
  // Save your user to database
  // saveUserToDB(user)

  // Then emit your event
  eventGateway
    .emit({
      event: 'user.created',
      data: user
    })
    .then(() => console.log('Emitted user.created event!'))
}

const user = {
  "username": "sls-fan",
  "firstname": "Bill",
  "lastname": "Jones",
  "company": "Big Corp, Inc.",
  "email": "bjones12@bigcorp.com"
}

createUser(user);
```

The `createUser` function is similar to one you would have in your application -- save the new user to your database, then emit a `user.created` event that other services can subscribe to.

You can run this file as a test if you like. You'll need to edit the `SPACE` variable to match the space you created in your `serverless.yml`. Then just execute the file and check your function logs:

```bash
$ node emit.js
Emitted user.created event!

$ sls logs -f addUserToCRM -t
START RequestId: 5a11a5d8-2d36-11e8-ac4f-eb85ef6fadda Version: $LATEST
2018-03-21 18:33:29.796 (+00:00)	5a11a5d8-2d36-11e8-ac4f-eb85ef6fadda	Saving user to CRM:
{ company: 'Big Corp, Inc.',
  email: 'bjones12@bigcorp.com',
  firstname: 'Bill',
  lastname: 'Jones',
  username: 'sls-fan' }
END RequestId: 5a11a5d8-2d36-11e8-ac4f-eb85ef6fadda
REPORT RequestId: 5a11a5d8-2d36-11e8-ac4f-eb85ef6fadda	Duration: 1.64 ms	Billed Duration: 100 ms 	Memory Size: 1024 MB	Max Memory Used: 21 MB
```

Success!
