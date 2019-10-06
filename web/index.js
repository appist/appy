const fs = require("fs");
const path = require("path");
const webpackErrorOverlayPlugin = require("error-overlay-webpack-plugin");
const webpackFaviconsPlugin = require("favicons-webpack-plugin");
const webpackWorkboxPlugin = require("workbox-webpack-plugin");
const srcDir = "./src";

function getVueConfig(pkg) {
  const ssrPaths = (() => {
    let paths = ["/service-worker.js"];

    if (process.env.APPY_SSR_PATHS && process.env.APPY_SSR_PATHS !== "") {
      paths = paths.concat(process.env.APPY_SSR_PATHS.split(","));
    }

    return paths;
  })();

  const https = (() => {
    return process.env.HTTP_SSL_ENABLED &&
      process.env.HTTP_SSL_ENABLED === "true"
      ? {
          key: fs.readFileSync(
            path.resolve(
              process.cwd(),
              `../${process.env.HTTP_SSL_CERT_PATH}/key.pem`
            )
          ),
          cert: fs.readFileSync(
            path.resolve(
              process.cwd(),
              `../${process.env.HTTP_SSL_CERT_PATH}/cert.pem`
            )
          )
        }
      : {};
  })();

  const proxyConfig =
    process.env.HTTP_SSL_ENABLED && process.env.HTTP_SSL_ENABLED === "true"
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
      devServer: {
        contentBase: path.resolve(process.cwd(), "public"),
        historyApiFallback: true,
        http2: true,
        https,
        hot: true,
        host: process.env.HTTP_HOST,
        port: parseInt(proxyConfig.port) + 1,
        overlay: {
          warnings: true,
          errors: true
        },
        index: "",
        proxy
      },
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
          navigateFallbackBlacklist: ssrPaths.map(p => new RegExp(p))
        })
      ]
    },

    outputDir: path.resolve(process.cwd(), "../assets"),

    pluginOptions: {
      i18n: {
        locale: "en",
        fallbackLocale: "en",
        localeDir: "locales",
        enableInSFC: true
      },
      webpackBundleAnalyzer: {
        openAnalyzer:
          process.env.BUNDLE_ANALYZER && process.env.BUNDLE_ANALYZER === "true"
      }
    }
  };
}

module.exports = {
  getVueConfig
};
