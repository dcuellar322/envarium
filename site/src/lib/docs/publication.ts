/** Repository-relative docs paths retained for maintainers, never published. */
export const internalDocumentation = [
  'progress/',
  'public-site.md',
  'public-site-launch-checklist.md',
  'envarium-rebrand-checklist.md',
  'reference/arcane-capability-review.md',
] as const

export function isPublicDocumentation(path: string): boolean {
  const normalized = path.replaceAll('\\', '/').replace(/^\.\//, '')
  return !internalDocumentation.some((entry) => entry.endsWith('/')
    ? normalized.startsWith(entry)
    : normalized === entry)
}

/** Apply the same boundary before Starlight creates routes or search entries. */
export const publicDocumentationPatterns = [
  '**/*.{md,mdx}',
  ...internalDocumentation.map((entry) => `!${entry}${entry.endsWith('/') ? '**' : ''}`),
]
