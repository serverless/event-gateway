const SDK = require('@serverless/event-gateway-sdk')
const URL = 'tenant-myapp.slsgateway.com'

const eventGateway = new SDK({
  url: URL
})

function createUser (user) {
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
  'username': 'sls-fan',
  'firstname': 'Bill',
  'lastname': 'Jones',
  'company': 'Big Corp, Inc.',
  'email': 'bjones12@bigcorp.com'
}

createUser(user)
