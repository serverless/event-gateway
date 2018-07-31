'use script'

module.exports.handler = (event, context, cb) => {
  const user = event.data
  console.log('Saving user to CRM: ')
  console.dir(user)

  // Add your CRM logic here
  // saveUserToCRM(user)

  cb(null, { message: 'success' })
}
