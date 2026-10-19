import { existsSync, readFileSync } from 'node:fs'
import { dirname, relative, resolve, sep } from 'node:path'
import { fileURLToPath } from 'node:url'
import { REPOSITORY_URL } from '../../config/site'
import { isPublicDocumentation } from './publication'

const repositoryRoot = fileURLToPath(new URL('../../../../', import.meta.url))
const docsRoot = resolve(repositoryRoot, 'docs')

interface MarkdownNode {
  type?: string
  url?: string
  children?: MarkdownNode[]
}

/** Keep GitHub-friendly source links while serving their canonical HTML routes. */
export function documentationHref(sourcePath: string, href: string): string {
  if (/^(?:[a-z]+:|\/|#)/i.test(href)) return href
  const match = /^([^?#]+\.mdx?)([?#].*)?$/.exec(href)
  if (!match) return href
  const target = resolve(dirname(sourcePath), decodeURIComponent(match[1]!))
  if (!existsSync(target)) throw new Error(`Missing documentation link in ${sourcePath}: ${href}`)
  const suffix = match[2] ?? ''
  const docPath = relative(docsRoot, target).split(sep).join('/')
  if (docPath.startsWith('../') || !isPublicDocumentation(docPath)) {
    const repositoryPath = relative(repositoryRoot, target).split(sep).join('/')
    if (repositoryPath.startsWith('../')) throw new Error(`Documentation link leaves repository: ${href}`)
    return `${REPOSITORY_URL}/blob/master/${repositoryPath}${suffix}`
  }
  const content = readFileSync(target, 'utf8')
  const frontmatter = /^---\r?\n([\s\S]*?)\r?\n---/.exec(content)?.[1] ?? ''
  const explicitSlug = /^slug:\s*["']?([^"'\r\n]+?)["']?\s*$/m.exec(frontmatter)?.[1]
  const slug = explicitSlug ?? `docs/${docPath.replace(/\.mdx?$/, '').replace(/(^|\/)README$/, '$1index')}`
  return `/${slug.replace(/\/index$/, '').replace(/^\/+|\/+$/g, '')}/${suffix}`
}

export default function remarkDocumentationLinks() {
  return (tree: MarkdownNode, file: { path: string }) => {
    const visit = (node: MarkdownNode) => {
      if ((node.type === 'link' || node.type === 'definition') && node.url) node.url = documentationHref(file.path, node.url)
      node.children?.forEach(visit)
    }
    visit(tree)
  }
}
