import type { SidebarsConfig } from '@docusaurus/plugin-content-docs';

const sidebars: SidebarsConfig = {
  docs: [
    'intro',
    'getting-started',
    'installation',
    {
      type: 'category',
      label: 'Concepts',
      items: [
        'concepts/wire-conformance',
        'concepts/stories',
        'concepts/profiles',
        'concepts/multi-station',
      ],
    },
    {
      type: 'category',
      label: 'Authoring',
      items: [
        'authoring/first-story',
        'authoring/keywords-reference',
        'authoring/multi-station-patterns',
      ],
    },
    {
      type: 'category',
      label: 'Operations',
      items: [
        'operations/ci-integration',
        'operations/reports',
        'operations/troubleshooting',
      ],
    },
    {
      type: 'category',
      label: 'Reference',
      items: [
        'reference/cli',
        'reference/config-schema',
        'reference/story-grammar',
        'reference/exit-codes',
      ],
    },
  ],
};

export default sidebars;
