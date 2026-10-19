import { fileURLToPath } from 'node:url'
import { describe, expect, it } from 'vitest'
import remarkDocumentationLinks, { documentationHref } from '../../src/lib/docs/remark-links'

const source = (path: string) => fileURLToPath(new URL(`../../../docs/${path}`, import.meta.url))

describe('documentation links', () => {
  it('resolves source-relative Markdown links to HTML routes and preserves fragments', () => {
    expect(documentationHref(source('cli.md'), 'settings.md#settings')).toBe('/docs/settings/#settings')
    expect(documentationHref(source('start/install.md'), '../getting-started.md')).toBe('/docs/getting-started/')
    expect(documentationHref(source('architecture/README.md'), '../adr/README.md')).toBe('/docs/adr/')
    expect(documentationHref(source('index.md'), 'start/index.md')).toBe('/docs/start/')
  })
  it('preserves external URLs, HTML routes, and same-page fragments', () => {
    for (const href of ['https://example.org/guide.md', '/download/', '#install', 'mailto:security@example.org']) {
      expect(documentationHref(source('index.md'), href)).toBe(href)
    }
  })
  it('links repository files outside docs to their source', () => {
    expect(documentationHref(source('index.md'), '../AGENTS.md')).toBe('https://github.com/dcuellar322/envarium/blob/master/AGENTS.md')
  })
  it('keeps links to unpublished maintainer documents on GitHub', () => {
    expect(documentationHref(source('contributing/index.md'), '../progress/phase-00.md')).toBe('https://github.com/dcuellar322/envarium/blob/master/docs/progress/phase-00.md')
    expect(documentationHref(source('index.md'), 'public-site.md#privacy')).toBe('https://github.com/dcuellar322/envarium/blob/master/docs/public-site.md#privacy')
  })
  it('fails the build for a missing Markdown destination', () => {
    expect(() => documentationHref(source('index.md'), 'missing-document.md')).toThrow('Missing documentation link')
  })
  it('rewrites reference definitions and inline links without changing code', () => {
    const tree = { children: [{ type: 'link', url: 'cli.md' }, { type: 'definition', url: 'settings.md' }, { type: 'code', url: 'cli.md' }] }
    remarkDocumentationLinks()(tree, { path: source('index.md') })
    expect(tree.children.map((node) => node.url)).toEqual(['/docs/cli/', '/docs/settings/', 'cli.md'])
  })
})
