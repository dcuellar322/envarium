import { defineCollection } from 'astro:content'
import { glob } from 'astro/loaders'
import { z } from 'astro/zod'
import { docsSchema } from '@astrojs/starlight/schema'
import { publicDocumentationPatterns } from './lib/docs/publication'

const docMetadata = z.object({
  category: z
    .enum(['tutorial', 'how-to', 'concept', 'reference', 'troubleshooting', 'contributor'])
    .optional(),
  audience: z.array(z.enum(['user', 'contributor', 'integrator', 'maintainer'])).optional(),
  platforms: z.array(z.enum(['macos', 'linux', 'windows', 'wsl'])).optional(),
  since: z.string().optional(),
  lastVerified: z.coerce.date().optional(),
  searchTerms: z.array(z.string()).optional(),
})

export const collections = {
  docs: defineCollection({
    loader: glob({
      pattern: publicDocumentationPatterns,
      base: '../docs',
      generateId: ({ entry }) => {
        const path = entry.replace(/\.(md|mdx)$/, '').replace(/(^|\/)README$/, '$1index')
        return `docs/${path}`
      },
    }),
    schema: docsSchema({ extend: docMetadata }),
  }),
}
