'use script'

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
