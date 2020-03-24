const sveltePreprocess = require("svelte-preprocess");

const scssAliases = aliases => {
  return url => {
    for (const [alias, aliasPath] of Object.entries(aliases)) {
      if (url.indexOf(alias) === 0) {
        return {
          file: url.replace(alias, aliasPath),
        };
      }
    }

    return url;
  };
};

module.exports = {
  preprocess: sveltePreprocess({
    scss: {
      data: `
        @import "bootstrap/scss/functions";
        @import "bootstrap/scss/mixins";
        @import "@assets/styles/theme/variables";
        @import "bootstrap/scss/variables";
      `,
      importer: [
        scssAliases({
          "@assets": `${process.cwd()}/assets`,
          "@": process.cwd(),
        }),
      ],
    },
    typescript: {
      transpileOnly: true,
    },
  }),
};
