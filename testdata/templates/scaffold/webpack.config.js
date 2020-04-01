const pkg = require("./package.json"),
  fs = require("fs"),
  path = require("path"),
  { BundleAnalyzerPlugin } = require("webpack-bundle-analyzer"),
  CaseSensitivePathsPlugin = require("case-sensitive-paths-webpack-plugin"),
  { CleanWebpackPlugin } = require("clean-webpack-plugin"),
  CopyWebpackPlugin = require("copy-webpack-plugin"),
  FaviconsWebpackPlugin = require("favicons-webpack-plugin"),
  FriendlyErrorsWebpackPlugin = require("friendly-errors-webpack-plugin"),
  HtmlWebpackPlugin = require("html-webpack-plugin"),
  ManifestPlugin = require("webpack-manifest-plugin"),
  MiniCssExtractPlugin = require("mini-css-extract-plugin"),
  OptimizeCssnanoPlugin = require("@intervolga/optimize-cssnano-plugin"),
  WorkboxWebpackPlugin = require("workbox-webpack-plugin"),
  { EnvironmentPlugin } = require("webpack");

// Indicate if the build is optimised for production deployment.
const isProduction = process.env.NODE_ENV === "production";

// Indicate if the `webpack-dev-server` should be running with HTTPS.
const isSSLEnabled = process.env.HTTP_SSL_ENABLED === "true";

// Indicate the folder that contains Svelte SPA source code.
const srcDir = "web/src";

// Indicate the folder that contains the optimised build assets.
const distDir = "dist";

// Indicate the folder that contains the public assets which are directly copied over to `distDir`.
const publicDir = "web/public";

// Indicate the SSL key/cert file location which will be used by `webpack-dev-server` when `HTTP_SSL_ENABLED` is
// set to `true`. By default, `HTTP_SSL_CERT_PATH` is set to `./tmp/ssl`.
const ssl = {
  key: path.resolve(__dirname, `${process.env.HTTP_SSL_CERT_PATH}/key.pem`),
  cert: path.resolve(__dirname, `${process.env.HTTP_SSL_CERT_PATH}/cert.pem`),
};

// Indicate the HTTPS configuration for `webpack-dev-server` to use when `HTTP_SSL_ENABLED` is set to `true`.
const https = (() => {
  return isSSLEnabled && fs.existsSync(ssl.key) && fs.existsSync(ssl.cert)
    ? {
        key: fs.readFileSync(ssl.key),
        cert: fs.readFileSync(ssl.cert),
      }
    : false;
})();

// Indicate the server-side rendering routes which is set by appy's `start` and `build` commands so that the service
// worker doesn't handle navigation fallback to `/index.html` when the current route matching 1 of these routes.
const ssrRoutes = (() => {
  let routes = [];

  if (process.env.APPY_SSR_ROUTES !== undefined && process.env.APPY_SSR_ROUTES !== "") {
    routes = routes.concat(process.env.APPY_SSR_ROUTES.split(","));
  }

  return routes;
})();

// Configure the `webpack-dev-server` for local development use.
const devServer = {
  historyApiFallback: true,
  https,
  host: process.env.HTTP_HOST || "0.0.0.0",
  port: parseInt(isSSLEnabled ? process.env.HTTP_SSL_PORT || 3443 : process.env.HTTP_PORT || 3000) + 1,
  hot: false,
  overlay: {
    warnings: true,
    errors: true,
  },
};

// Capitalize the application name.
const appName = pkg.name.charAt(0).toUpperCase() + pkg.name.slice(1);

module.exports = {
  ...(isProduction
    ? {
        mode: "production",
        stats: {
          assets: true,
          assetsSort: "!size",
          builtAt: false,
          children: false,
          colors: true,
          modules: true,
        },
      }
    : {
        devServer,
        devtool: "eval-source-map",
        mode: "development",
        stats: "minimal",
      }),
  entry: {
    app: path.resolve(__dirname, srcDir, "main.ts"),
  },
  module: {
    rules: [
      {
        test: /\.css$/,
        use: [isProduction ? MiniCssExtractPlugin.loader : "style-loader", "css-loader", "postcss-loader"],
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
                  name: "images/[name].[contenthash:12].[ext]",
                },
              },
            },
          },
        ],
      },
      {
        test: /\.(svg)(\?.*)?$/,
        use: [
          {
            loader: "file-loader",
            options: {
              name: "images/[name].[contenthash:12].[ext]",
            },
          },
        ],
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
                  name: "medias/[name].[contenthash:12].[ext]",
                },
              },
            },
          },
        ],
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
                  name: "fonts/[name].[contenthash:12].[ext]",
                },
              },
            },
          },
        ],
      },
      {
        test: /\.tsx?$/,
        exclude: /node_modules/,
        use: [
          {
            loader: "babel-loader?cacheDirectory=true",
          },
          {
            loader: "ts-loader",
            options: {
              transpileOnly: true,
              happyPackMode: false,
              appendTsxSuffixTo: ["\\.svelte$"],
            },
          },
        ],
      },
      {
        test: /\.svelte$/,
        use: [
          {
            loader: "babel-loader?cacheDirectory=true",
          },
          {
            loader: "svelte-loader",
            options: {
              emitCss: isProduction,
              hotReload: false,
              preprocess: require("./svelte.config").preprocess,
            },
          },
        ],
      },
    ],
  },
  optimization: {
    runtimeChunk: "single",
    splitChunks: {
      chunks: "all",
      maxInitialRequests: Infinity,
      minSize: 0,
      cacheGroups: {
        vendor: {
          test: /[\\/]node_modules[\\/]/,
          name(module) {
            const packageName = module.context.match(/[\\/]node_modules[\\/](.*?)([\\/]|$)/)[1];

            return `vendor.${packageName.replace("@", "")}`;
          },
        },
      },
    },
  },
  output: {
    ...(isProduction
      ? {
          chunkFilename: "scripts/chunk.[name].[contenthash:12].js",
          filename: "scripts/[name].[contenthash:12].js",
        }
      : {
          chunkFilename: "scripts/chunk.[name].js",
          filename: "scripts/[name].js",
        }),
    path: path.resolve(__dirname, distDir),
    publicPath: "/",
  },
  plugins: [
    new CleanWebpackPlugin(),
    new EnvironmentPlugin({
      NODE_ENV: process.env.NODE_ENV,
      BASE_URL: "/",
      AVAILABLE_LOCALES: (() => fs.readdirSync(`${__dirname}/${srcDir}/locales`).map(fn => fn.replace(".json", "")))(),
    }),
    new CaseSensitivePathsPlugin(),
    new FriendlyErrorsWebpackPlugin({
      additionalTransformers: [],
      additionalFormatters: [],
    }),
    ...(isProduction
      ? [
          new MiniCssExtractPlugin({
            filename: "styles/[name].[contenthash:12].css",
            chunkFilename: "styles/chunk.[name].[contenthash:12].css",
          }),
          new OptimizeCssnanoPlugin({
            sourceMap: true,
            cssnanoOptions: {
              preset: [
                "default",
                {
                  mergeLonghand: false,
                  cssDeclarationSorter: false,
                },
              ],
            },
          }),
        ]
      : []),
    new HtmlWebpackPlugin({
      title: appName,
      scriptLoading: "defer",
      template: path.resolve(__dirname, publicDir, "index.html"),
      minify: isProduction,
    }),
    new CopyWebpackPlugin(
      [
        {
          from: path.resolve(__dirname, publicDir),
          to: path.resolve(__dirname, distDir),
          toType: "dir",
          ignore: [
            {
              glob: "index.html",
              matchBase: false,
            },
          ],
        },
        {
          from: path.resolve(__dirname, "assets"),
          to: path.resolve(
            __dirname,
            isProduction ? `${distDir}/[path][name].[contenthash:12].[ext]` : `${distDir}/[path][name].[ext]`
          ),
        },
      ],
      {
        copyUnmodified: true,
        ignore: [".DS_Store", ".gitkeep", "**/*.{jsx,less,sass,scss,ts,tsx}"],
      }
    ),
    new ManifestPlugin({
      map: function (file) {
        file.name = file.name.replace(/(\.[a-z0-9]{12})(\..*)$/i, "$2");

        return file;
      },
    }),
    ...(isProduction
      ? [
          new BundleAnalyzerPlugin({
            analyzerMode: "disabled",
            analyzerHost: devServer.host,
            analyzerPort: parseInt(devServer.port) + 2,
            openAnalyzer: false,
          }),
          new FaviconsWebpackPlugin({
            cache: false,
            favicons: Object.assign(
              {},
              (() =>
                Object.assign({}, pkg.pwa, {
                  appName: appName,
                  appShortName: appName,
                  appDescription: pkg.description,
                }))(),
              {
                icons: {
                  coast: false,
                  favicons: false,
                  firefox: false,
                  yandex: false,
                },
              }
            ),
            inject: true,
            logo: path.resolve(__dirname, "web/public/brand.png"),
            prefix: "pwa/",
          }),
          new WorkboxWebpackPlugin.GenerateSW({
            skipWaiting: true,
            clientsClaim: true,
            navigateFallback: "/index.html",
            navigateFallbackDenylist: ssrRoutes.concat(["/service-worker.js"]).map(p => new RegExp(p)),
          }),
        ]
      : []),
  ],
  resolve: {
    alias: {
      "@assets": path.resolve(__dirname, "assets"),
      "@": path.resolve(__dirname, srcDir),
      svelte: path.resolve(__dirname, "node_modules", "svelte"),
    },
    extensions: [".mjs", ".js", ".jsx", ".json", ".svelte", ".ts", ".tsx"],
    mainFields: ["svelte", "browser", "module", "main"],
  },
};
