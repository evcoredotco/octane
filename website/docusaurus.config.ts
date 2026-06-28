import type { Config } from '@docusaurus/types';
import type * as Preset from '@docusaurus/preset-classic';

const config: Config = {
  title: 'OCTANE',
  tagline: 'Wire-level OCPP 1.6J conformance testing for CSMS teams',
  favicon: 'img/favicon.svg',

  url: 'https://octane.dev',
  baseUrl: '/',

  organizationName: 'evcoreco',
  projectName: 'octane',

  onBrokenLinks: 'throw',

  markdown: {
    hooks: {
      onBrokenMarkdownLinks: 'warn',
    },
  },

  i18n: {
    defaultLocale: 'en',
    locales: ['en'],
  },

  headTags: [
    {
      tagName: 'link',
      attributes: { rel: 'preconnect', href: 'https://fonts.googleapis.com' },
    },
    {
      tagName: 'link',
      attributes: {
        rel: 'preconnect',
        href: 'https://fonts.gstatic.com',
        crossorigin: 'anonymous',
      },
    },
  ],

  stylesheets: [
    'https://fonts.googleapis.com/css2?family=IBM+Plex+Mono:ital,wght@0,400;0,500;0,600;0,700;1,400;1,500&family=IBM+Plex+Sans:wght@400;500;600;700&display=swap',
  ],

  presets: [
    [
      'classic',
      {
        docs: {
          sidebarPath: './sidebars.ts',
          editUrl: 'https://github.com/evcoreco/octane/tree/master/website/',
          // Versioning per ADR 0013; enabled once the first release is tagged.
        },
        blog: false,
        theme: {
          customCss: './src/css/custom.css',
        },
      } satisfies Preset.Options,
    ],
  ],

  plugins: [
    [
      require.resolve('@easyops-cn/docusaurus-search-local'),
      {
        hashed: true,
        indexBlog: false,
        docsRouteBasePath: '/docs',
      },
    ],
  ],

  themeConfig: {
    metadata: [
      {
        name: 'keywords',
        content:
          'OCPP, OCPP 1.6J, conformance testing, CSMS, EV charging, charge point, charging station, wire-level testing, Robot Framework',
      },
      {
        name: 'description',
        content:
          'OCTANE is an open-source conformance harness for OCPP 1.6J. It impersonates charging stations over the wire and verifies your CSMS responds to spec — no CSMS changes required.',
      },
    ],
    colorMode: {
      defaultMode: 'dark',
      respectPrefersColorScheme: true,
    },
    navbar: {
      title: 'OCTANE',
      hideOnScroll: true,
      logo: {
        alt: 'OCTANE',
        src: 'img/logo.svg',
      },
      items: [
        {
          type: 'docSidebar',
          sidebarId: 'docs',
          position: 'left',
          label: 'Docs',
        },
        {
          to: '/docs/getting-started',
          label: 'Get started',
          position: 'left',
        },
        {
          to: '/docs/reference/cli',
          label: 'Reference',
          position: 'left',
        },
        {
          href: 'https://github.com/evcoreco/octane',
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
            { label: 'Introduction', to: '/docs/intro' },
            { label: 'Getting started', to: '/docs/getting-started' },
            { label: 'Authoring stories', to: '/docs/authoring/first-story' },
            { label: 'CLI reference', to: '/docs/reference/cli' },
          ],
        },
        {
          title: 'Concepts',
          items: [
            { label: 'Wire conformance', to: '/docs/concepts/wire-conformance' },
            { label: 'Architecture', to: '/docs/concepts/architecture' },
            {
              label: 'Dependency graph & caching',
              to: '/docs/concepts/dependency-graph',
            },
            { label: 'OCPP 1.6 coverage', to: '/docs/reference/ocpp-coverage' },
          ],
        },
        {
          title: 'Project',
          items: [
            { label: 'GitHub', href: 'https://github.com/evcoreco/octane' },
            {
              label: 'OCPP 1.6 (Open Charge Alliance)',
              href: 'https://www.openchargealliance.org/protocols/ocpp-16/',
            },
            { label: 'CitrineOS', href: 'https://citrineos.github.io/' },
          ],
        },
      ],
      copyright: `Apache-2.0 · © ${new Date().getFullYear()} The OCTANE Authors · OCTANE is not affiliated with or endorsed by the Open Charge Alliance.`,
    },
    prism: {
      additionalLanguages: ['bash', 'yaml', 'json', 'go'],
    },
  } satisfies Preset.ThemeConfig,
};

export default config;
