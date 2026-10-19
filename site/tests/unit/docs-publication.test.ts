import { describe, expect, it } from 'vitest'
import { internalDocumentation, isPublicDocumentation, publicDocumentationPatterns } from '../../src/lib/docs/publication'

describe('documentation publication boundary', () => {
  it('excludes all current and future progress reports and maintainer records', () => {
    for (const path of ['progress/phase-04.md', 'progress/site-phase-10.md', 'progress/future/note.mdx', ...internalDocumentation.filter((path) => !path.endsWith('/'))]) {
      expect(isPublicDocumentation(path), path).toBe(false)
    }
    expect(isPublicDocumentation('progress\\phase-04.md')).toBe(false)
    expect(isPublicDocumentation('./progress/phase-04.md')).toBe(false)
    expect(publicDocumentationPatterns).toContain('!progress/**')
    for (const path of internalDocumentation.filter((path) => !path.endsWith('/'))) {
      expect(publicDocumentationPatterns).toContain(`!${path}`)
    }
  })
  it('retains user, contributor, architecture, and security documentation', () => {
    for (const path of ['index.md', 'cli.md', 'start/install.md', 'contributing/index.md', 'release.md', 'architecture/README.md', 'adr/0017-static-public-site.md', 'security/threat-model.md']) {
      expect(isPublicDocumentation(path), path).toBe(true)
    }
  })
})
