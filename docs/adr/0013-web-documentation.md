# ADR 0013: Web Documentation — Docusaurus, Separate from Man Pages

- **Status:** Accepted
- **Date:** 2026-04-26
- **Deciders:** Project maintainer, Docs
- **Constitution principles touched:** II (Two Distribution Surfaces)

## Context

Man pages (ADR 0011) cover operator reference for users with the
binary installed. They are not appropriate for:

- Onboarding flows ("getting started in 5 minutes").
- Conceptual narrative with diagrams.
- Deep tutorials that walk through writing a story and running it.
- Conformance scenario catalog browsing with cross-references.
- Profile registry discovery.
- Public-facing project marketing.

A web docs site is the right surface for these. The question is
whether the website should re-render man-page content (single source,
two outputs) or maintain its own content tree separate from man
sources.

## Decision

Adopt **Docusaurus** under `website/`, with content **separate** from
man-page sources. The website is its own information architecture,
not an HTML render of the man pages.

### Rationale for separation

- Man pages and web docs serve different reading modes. `man octane.7`
  is a 400-line concept dump; the website's "Getting Started" is a
  guided walkthrough with screenshots and code samples that wouldn't
  render in a terminal.
- Single-source-pipelines collapse under different content shapes:
  forcing a tutorial into the section-7 mdoc format makes both the
  man page and the web page worse.
- Cross-references work naturally between the two surfaces:
  - Web pages link to man pages via `man:octane(1)` style hints.
  - Man pages cite the website as `https://octane.dev/docs/<topic>`.

### Site structure

```
website/
├── docs/
│   ├── intro.md                      # what is OCTANE, why use it
│   ├── getting-started.md            # 5-minute first run
│   ├── installation.md               # apt/dnf/brew/scoop/docker
│   ├── concepts/
│   │   ├── wire-conformance.md
│   │   ├── stories.md
│   │   ├── profiles.md
│   │   └── multi-station.md
│   ├── authoring/
│   │   ├── first-story.md
│   │   ├── keywords-reference.md     # generated, see below
│   │   └── multi-station-patterns.md
│   ├── operations/
│   │   ├── ci-integration.md
│   │   ├── reports.md
│   │   └── troubleshooting.md
│   ├── reference/
│   │   ├── cli.md                    # generated from cobra
│   │   ├── config-schema.md          # generated from JSON Schema
│   │   ├── story-grammar.md          # cross-link to ADR 0006
│   │   └── exit-codes.md
│   └── adrs/                         # symlink to docs/adr/
├── src/
├── docusaurus.config.ts
├── sidebars.ts
└── package.json
```

### Generated content (the limited single-source pipeline)

Three pages are mechanically generated, not hand-written:

1. **`reference/cli.md`** — produced by a `make docs-cli-reference`
   target that runs `octane gen-docs --format markdown` (a
   cobra-doc-driven companion to the man-page generator from ADR 0011).
2. **`reference/config-schema.md`** — produced by feeding the
   versioned config JSON Schema through `json-schema-for-humans` or
   equivalent.
3. **`authoring/keywords-reference.md`** — produced by walking the
   registered keyword library and emitting the pattern + summary
   per keyword.

These three pages live under `website/docs/` only after generation;
they are gitignored. CI regenerates them on every build.

Everything else is hand-written.

### Hosting and deployment

- Hosted at `https://octane.dev` (or a chosen domain).
- Built and deployed by GitHub Pages from a `gh-pages` branch.
- Workflow `.github/workflows/docs.yml` runs on every push to `main`:
  - Generates the three pages.
  - Builds the Docusaurus site.
  - Validates internal links.
  - Deploys.
- A versioned snapshot is taken on every release tag using
  Docusaurus's built-in versioning (`docusaurus docs:version`).

### Search

Algolia DocSearch (free for open-source) indexes the site. Until the
DocSearch application is approved, the local `@easyops-cn/docusaurus-search-local`
plugin provides client-side search.

### Themes and styling

- Default Docusaurus theme with a custom logo and a single accent
  colour aligned with the OCTANE brand (decision deferred to a
  brand sub-issue).
- No analytics in v1. If added later, a privacy notice will be
  required and an ADR amendment opened.

### Build dependencies

- Node 20+ LTS for Docusaurus.
- Confined to `website/`; the rest of the repo remains Go-only.
- `package-lock.json` committed; CI runs `npm ci`, never `npm install`.

## Consequences

### Positive

- Onboarding, tutorials, and concept narrative get the right medium.
- The man pages remain terse and reference-grade, which is what they
  should be.
- Generated reference pages eliminate drift between the binary and
  three high-traffic docs surfaces (CLI, config, keywords).
- Versioned snapshots let users targeting an older OCTANE release
  read the docs for that release.

### Negative

- Node toolchain in a Go-first project. Confined to `website/`;
  contributors who do not edit docs do not need Node.
- Docusaurus version churn. Mitigated by pinning the major version
  and bumping deliberately on a documented cadence.

### Neutral

- The site competes for the project's mindshare with the README and
  the man pages. Resolved by treating each surface as having a
  specific job: README is repository-level introduction, man pages
  are reference, website is everything in between.

## Alternatives considered

- **Single-source-of-truth with man pages as input.** Considered and
  rejected: man-flavored content is too constrained for tutorials.
- **mdBook / Hugo / Zola.** Considered. Docusaurus has stronger
  versioning, search, and CSDS-friendly defaults; ecosystem
  familiarity is also higher among the OCPP / industrial-IT user
  base.
- **No website — README + GitHub wiki only.** Rejected: GitHub Wiki
  is unindexed by search engines and discourages PR-reviewable docs
  changes.
- **Hosted GitBook / ReadTheDocs.** Considered. Self-hosting via
  GitHub Pages avoids vendor lock-in and matches the project's
  open-source posture.

## References

- Constitution: principle II
- ADR 0011 (man pages)
- Docusaurus: https://docusaurus.io/
