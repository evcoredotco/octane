---
sidebar_position: 2
---

# Stories

Stories are `.story` files written in OCTANE's Gherkin-flavored DSL. They
contain metadata, optional setup and teardown sections, and one or more
scenarios made of keyword steps.

Conformance stories must include a `Spec-Ref`. Helper stories must omit
`Spec-Ref` and carry the `helper` tag.

For the full syntax reference, see `docs/concepts/story-syntax.md`.

