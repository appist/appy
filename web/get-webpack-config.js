const fs = require("fs");
const path = require("path");
const { BundleAnalyzerPlugin } = require("webpack-bundle-analyzer");
const CaseSensitivePathsPlugin = require("case-sensitive-paths-webpack-plugin");
const { CleanWebpackPlugin } = require("clean-webpack-plugin");
const CopyWebpackPlugin = require("copy-webpack-plugin");
const FaviconsWebpackPlugin = require("favicons-webpack-plugin");
const FriendlyErrorsWebpackPlugin = require("friendly-errors-webpack-plugin");
const HtmlWebpackPlugin = require("html-webpack-plugin");
const MiniCssExtractPlugin = require("mini-css-extract-plugin");
const OptimizeCssnanoPlugin = require("@intervolga/optimize-cssnano-plugin");
const PreloadWebpackPlugin = require("preload-webpack-plugin");
const WorkboxWebpackPlugin = require("workbox-webpack-plugin");
const {
  EnvironmentPlugin,
  HashedModuleIdsPlugin,
  NamedChunksPlugin
} = require("webpack");

module.exports = function(pkg) {
  const isProduction = process.env.NODE_ENV === "production" ? true : false;
  const srcDir = "web/src";
  const publicDir = "web/public";
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
    key: `${process.env.HTTP_SSL_CERT_PATH}/key.pem`,
    cert: `${process.env.HTTP_SSL_CERT_PATH}/cert.pem`
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
      ? { port: process.env.HTTP_SSL_PORT || 3443, scheme: "https" }
      : { port: process.env.HTTP_PORT || 3000, scheme: "http" };
  proxyConfig.host = process.env.HTTP_HOST || "0.0.0.0";

  let devServer = {
      historyApiFallback: true,
      https,
      host: proxyConfig.host,
      port: parseInt(proxyConfig.port) + 1,
      overlay: {
        warnings: true,
        errors: true
      }
    },
    proxy = {};

  ssrPaths.map(p => {
    if (p !== "/") {
      proxy[p] = {
        secure: false,
        ws: true,
        target: `${proxyConfig.scheme}://${proxyConfig.host}:${proxyConfig.port}`
      };
    }
  });

  if (Object.keys(proxy).length > 0) {
    devServer = Object.assign({}, devServer, { proxy });
  }

  return {
    mode: isProduction ? "production" : "development",
    devServer,
    devtool: isProduction ? "false" : "source-map",
    entry: {
      app: path.resolve(srcDir, "main.js")
    },
    module: {
      rules: [
        {
          test: /\.css$/,
          use: [MiniCssExtractPlugin.loader, "css-loader", "postcss-loader"]
        },
        {
          test: /\.(m)js$/,
          exclude: /node_modules/,
          use: {
            loader: "babel-loader?cacheDirectory=true"
          }
        },
        {
          test: /\.(png|jpe?g|gif|webp)(\?.*)?$/,
          use: [
            {
              loader: "url-loader",
              options: {
                limit: 4096,
                fallback: {
                  loader: "file-loader",
                  options: {
                    name: "img/[name].[hash:8].[ext]"
                  }
                }
              }
            }
          ]
        },
        {
          test: /\.(svg)(\?.*)?$/,
          use: [
            {
              loader: "file-loader",
              options: {
                name: "img/[name].[hash:8].[ext]"
              }
            }
          ]
        },
        {
          test: /\.(mp4|webm|ogg|mp3|wav|flac|aac)(\?.*)?$/,
          use: [
            {
              loader: "url-loader",
              options: {
                limit: 4096,
                fallback: {
                  loader: "file-loader",
                  options: {
                    name: "media/[name].[hash:8].[ext]"
                  }
                }
              }
            }
          ]
        },
        {
          test: /\.(woff2?|eot|ttf|otf)(\?.*)?$/i,
          use: [
            {
              loader: "url-loader",
              options: {
                limit: 4096,
                fallback: {
                  loader: "file-loader",
                  options: {
                    name: "fonts/[name].[hash:8].[ext]"
                  }
                }
              }
            }
          ]
        },
        {
          test: /\.svelte$/,
          use: [
            {
              loader: "svelte-loader",
              options: {
                emitCss: isProduction,
                hotReload: !isProduction,
                preprocess: require("svelte-preprocess")({})
              }
            }
          ]
        }
      ]
    },
    output: {
      chunkFilename: isProduction
        ? "js/[name].[contenthash:8].js"
        : "js/[name].js",
      filename: isProduction ? "js/[name].[contenthash:8].js" : "js/[name].js",
      path: path.resolve("assets"),
      publicPath: ""
    },
    plugins: [
      new CleanWebpackPlugin(),
      new EnvironmentPlugin({
        NODE_ENV: "development",
        BASE_URL: "/"
      }),
      new CaseSensitivePathsPlugin(),
      new FriendlyErrorsWebpackPlugin({
        additionalTransformers: [],
        additionalFormatters: []
      }),
      ...(isProduction
        ? [
            new MiniCssExtractPlugin({
              filename: "css/[name].[contenthash:8].css",
              chunkFilename: "css/[name].[contenthash:8].css"
            }),
            new OptimizeCssnanoPlugin({
              sourceMap: false,
              cssnanoOptions: {
                preset: [
                  "default",
                  {
                    mergeLonghand: false,
                    cssDeclarationSorter: false
                  }
                ]
              }
            }),
            new HashedModuleIdsPlugin({
              hashDigest: "hex"
            }),
            new NamedChunksPlugin(function() {})
          ]
        : []),
      new HtmlWebpackPlugin({
        title: pkg.name,
        template: path.resolve(publicDir, "index.html"),
        minify: isProduction
          ? {
              removeComments: true,
              collapseWhitespace: true,
              removeAttributeQuotes: true,
              collapseBooleanAttributes: true,
              removeScriptTypeAttributes: true
            }
          : {}
      }),
      new PreloadWebpackPlugin({
        rel: "preload",
        include: "initial",
        fileBlacklist: [/\.map$/, /hot-update\.js$/]
      }),
      new PreloadWebpackPlugin({
        rel: "prefetch",
        include: "asyncChunks"
      }),
      new CopyWebpackPlugin([
        {
          from: path.resolve(publicDir),
          to: path.resolve("assets"),
          toType: "dir",
          ignore: [
            ".DS_Store",
            {
              glob: "index.html",
              matchBase: false
            }
          ]
        }
      ]),
      ...(isProduction
        ? [
            new BundleAnalyzerPlugin({
              analyzerMode: "static",
              analyzerHost: proxyConfig.host,
              analyzerPort: parseInt(proxyConfig.port) + 2,
              openAnalyzer: false
            })
          ]
        : []),
      new FaviconsWebpackPlugin({
        cache: !isProduction,
        favicons: Object.assign({}, pkg.pwa, {
          icons: {
            coast: false,
            firefox: false,
            yandex: false
          }
        }),
        inject: true,
        logo: path.resolve(`${srcDir}/assets/images/logo.png`),
        prefix: "pwa/"
      }),
      new WorkboxWebpackPlugin.GenerateSW({
        skipWaiting: true,
        clientsClaim: true,
        navigateFallback: "/index.html",
        navigateFallbackBlacklist: ssrPaths
          .concat(["/service-worker.js"])
          .map(p => new RegExp(p))
      })
    ],
    resolve: {
      alias: {
        "@": path.resolve(srcDir)
      },
      extensions: [".mjs", ".js", ".json", ".svelte"],
      mainFields: ["svelte", "browser", "module", "main"]
    },
    stats: isProduction
      ? {
          assets: true,
          assetsSort: "!size",
          builtAt: false,
          children: false,
          colors: true,
          modules: false
        }
      : "minimal"
  };
};
