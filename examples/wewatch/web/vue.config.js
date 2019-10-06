const appy = require("@appist/appy");
const pkg = require("./package.json");

const config = appy.getVueConfig(pkg);
console.log(config);
module.exports = config;
