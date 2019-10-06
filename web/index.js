const fs = require("fs");
const path = require("path");
const webpackErrorOverlayPlugin = require("error-overlay-webpack-plugin");
const webpackFaviconsPlugin = require("favicons-webpack-plugin");
const webpackWorkboxPlugin = require("workbox-webpack-plugin");

function getVueConfig(pkg) {
  const srcDir = "./src";
  const ssrPaths = (() => {
    let paths = [];

    if (
      process.env.APPY_SSR_PATHS !== undefined &&
      process.env.APPY_SSR_PATHS !== ""
    ) {
      paths = paths.concat(process.env.APPY_SSR_PATHS.split(","));
    }

    return paths;
  })();

  const ssl = {
    key: `../${process.env.HTTP_SSL_CERT_PATH}/key.pem`,
    cert: `../${process.env.HTTP_SSL_CERT_PATH}/cert.pem`
  };
  const https = (() => {
    return process.env.HTTP_SSL_ENABLED !== undefined &&
      process.env.HTTP_SSL_ENABLED === "true" &&
      fs.existsSync(ssl.key) &&
      fs.existsSync(ssl.cert)
      ? {
          key: fs.readFileSync(path.resolve(process.cwd(), ssl.key)),
          cert: fs.readFileSync(path.resolve(process.cwd(), ssl.cert))
        }
      : false;
  })();

  const proxyConfig =
    process.env.HTTP_SSL_ENABLED !== undefined &&
    process.env.HTTP_SSL_ENABLED === "true"
      ? { port: process.env.HTTP_SSL_PORT, scheme: "https" }
      : { port: process.env.HTTP_PORT, scheme: "http" };

  let proxy = {};
  ssrPaths.map(p => {
    proxy[p] = {
      secure: false,
      target: `${proxyConfig.scheme}://${process.env.HTTP_HOST}:${proxyConfig.port}`
    };
  });

  return {
    css: {
      loaderOptions: {
        sass: {
          prependData: `@import '~@/main.scss';`
        }
      }
    },

    configureWebpack: {
      devtool: "cheap-module-source-map",
      plugins: [
        new webpackErrorOverlayPlugin(),
        new webpackFaviconsPlugin({
          cache: process.env.NODE_ENV === "production" ? false : true,
          favicons: Object.assign({}, pkg.pwa, {
            icons: {
              coast: false,
              firefox: false,
              yandex: false
            }
          }),
          inject: true,
          logo: `${srcDir}/assets/logo.svg`,
          prefix: "pwa/"
        }),
        new webpackWorkboxPlugin.GenerateSW({
          skipWaiting: true,
          clientsClaim: true,
          navigateFallback: "/index.html",
          navigateFallbackBlacklist: ssrPaths
            .concat(["/service-worker.js"])
            .map(p => new RegExp(p))
        })
      ]
    },

    devServer: {
      https,
      host: process.env.HTTP_HOST,
      port: parseInt(proxyConfig.port) + 1,
      overlay: {
        warnings: true,
        errors: true
      },
      proxy
    },

    outputDir: path.resolve(process.cwd(), "../assets"),

    pluginOptions: {
      i18n: {
        locale: "en",
        fallbackLocale: "en",
        localeDir: "locales",
        enableInSFC: false
      },
      webpackBundleAnalyzer: {
        openAnalyzer: false
      }
    }
  };
}

module.exports = {
  getVueConfig
};
