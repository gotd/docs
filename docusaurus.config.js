// @ts-check
// Note: type annotations allow type checking and IDEs autocompletion

const {themes} = require('prism-react-renderer');
const lightCodeTheme = themes.github;
const darkCodeTheme = themes.dracula;

/** @type {import('@docusaurus/types').Config} */
const config = {
  title: 'gotd',
  tagline: 'Telegram client in Go',
  url: 'https://gotd.dev',
  baseUrl: '/',
  onBrokenLinks: 'throw',
  future: {
    v4: true,
  },
  headTags: [
    {
      tagName: 'meta',
      attributes: {
        name: 'algolia-site-verification',
        content: 'F1C0F77133870978',
      },
    },
  ],
  markdown: {
    hooks: {
      onBrokenMarkdownLinks: 'warn',
    },
  },
  organizationName: 'gotd', // Usually your GitHub org/user name.
  projectName: 'td', // Usually your repo name.

  presets: [
    [
      '@docusaurus/preset-classic',
      /** @type {import('@docusaurus/preset-classic').Options} */
      ({
        docs: {
          sidebarPath: require.resolve('./sidebars.js'),
          // Please change this to your repo.
          editUrl: 'https://github.com/gotd/docs/edit/main/docs/',
        },
        // blog: {
        //   showReadingTime: true,
        //   // Please change this to your repo.
        //   editUrl:
        //     'https://github.com/gotd/docs/edit/main/blog/',
        // },
        theme: {
          customCss: require.resolve('./src/css/custom.css'),
        },
      }),
    ],
  ],

  themeConfig:
    /** @type {import('@docusaurus/preset-classic').ThemeConfig} */
    ({
      navbar: {
        title: 'gotd',
        logo: {
          alt: 'gotd logo',
          src: 'https://github.com/gotd.png',
        },
        items: [
          {
            type: 'docSidebar',
            sidebarId: 'tutorialSidebar',
            position: 'left',
            label: 'Tutorial',
          },
          {
            type: 'docSidebar',
            sidebarId: 'botapiSidebar',
            position: 'left',
            label: 'Bot API',
          },
          {
            type: 'docSidebar',
            sidebarId: 'referenceSidebar',
            position: 'left',
            label: 'API Reference',
          },
          // {to: '/blog', label: 'Blog', position: 'left'},
          {
            href: 'https://github.com/gotd/td',
            label: 'GitHub',
            position: 'right',
          },
        ],
      },
      footer: {
        style: 'dark',
        links: [
          {
            title: 'Docs',
            items: [
              {
                label: 'Tutorial',
                to: '/docs/intro',
              },
            ],
          },
          {
            title: 'Community',
            items: [
              {
                label: 'Telegram',
                href: 'https://t.me/gotd_en',
              },
              {
                label: 'Telegram (in russian)',
                href: 'https://t.me/gotd_ru',
              },
            ],
          },
          {
            title: 'More',
            items: [
              // {
              //   label: 'Blog',
              //   to: '/blog',
              // },
              {
                label: 'GitHub',
                href: 'https://github.com/gotd/td',
              },
            ],
          },
        ]
      },
      prism: {
        theme: lightCodeTheme,
        darkTheme: darkCodeTheme,
      },
      // Algolia DocSearch. The apiKey is the public, search-only key — it is
      // safe to commit. Values can be overridden via environment variables.
      algolia: {
        appId: process.env.ALGOLIA_APP_ID || 'X2N6LG2Z0C',
        apiKey: process.env.ALGOLIA_API_KEY || '7211a1ddff7c214708ba57a83ddb74bd',
        indexName: process.env.ALGOLIA_INDEX_NAME || 'gotd',
        contextualSearch: true,
      },
    }),
};

module.exports = config;
