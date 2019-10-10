const appy = require('@appist/appy')
const pkg = require('../package.json')

module.exports = appy.getWebpackConfig(pkg)
