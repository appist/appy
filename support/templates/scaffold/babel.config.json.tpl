{
  "compact": false,
  "plugins": ["@babel/plugin-syntax-dynamic-import"],
  "presets": [
    [
      "@babel/preset-env",
      {
        "targets": {
          "node": "current"
        }
      }
    ],
    "@babel/preset-typescript"
  ]
}
