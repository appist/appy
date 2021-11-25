/**
 * @type {import('vitepress').UserConfig}
 */
export default {
  base: "/appy/",
  title: "appy",
  description: "",
  head: [["link", { rel: "icon", type: "image/png", href: "/appy/logo.png" }]],
  themeConfig: {
    repo: "appist/appy",
    logo: "/logo.png",
    docsDir: "docs",
    docsBranch: "main",
    editLinks: true,
    editLinkText: "Suggest changes to this page",
    nav: [{ text: "Guides", link: "/guides/" }],
    sidebar: {
      "/guides/": getGuidesSidebar(),
    },
  },
};

function getGuidesSidebar() {
  return [
    {
      text: "Introduction",
      children: [
        { text: "What is appy?", link: "/guides/" },
        { text: "Getting Started", link: "/guides/getting-started" },
      ],
    },
    {
      text: "Advanced",
      children: [],
    },
  ];
}
