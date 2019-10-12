const sveltePreprocess = require('svelte-preprocess')

module.exports = {
  coverageDirectory: 'coverage',
  transform: {
    '^.+\\.svelte$': [
      'jest-transform-svelte',
      { preprocess: sveltePreprocess() },
    ],
    '.+\\.(css|styl|less|sass|scss|svg|png|jpg|ttf|woff|woff2)$':
      'jest-transform-stub',
    '^.+\\.jsx?$': 'babel-jest',
  },
  moduleFileExtensions: ['js', 'jsx', 'json', 'svelte'],
  moduleNameMapper: {
    '^@/(.*)$': '<rootDir>/web/src/$1',
  },
}
