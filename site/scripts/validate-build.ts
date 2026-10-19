import { readdir, readFile, stat } from 'node:fs/promises'
import { extname, relative, resolve } from 'node:path'
import { DEFAULT_SITE_URL, PRODUCT_NAME } from '../src/config/site'
import { docsRoot, siteRoot } from './paths'
import { internalDocumentationInvocationRoutes, isPublicDocumentation } from '../src/lib/docs/publication'

const dist = resolve(siteRoot, 'dist')

async function filesUnder(directory: string): Promise<string[]> {
  const entries = await readdir(directory, { withFileTypes: true })
  return (await Promise.all(entries.map((entry) => {
    const path = resolve(directory, entry.name)
    return entry.isDirectory() ? filesUnder(path) : [path]
  }))).flat()
}

const files = await filesUnder(dist)
const htmlFiles = files.filter((path) => path.endsWith('.html'))
const titles = new Map<string, string>()
const descriptions = new Map<string, string>()
const problems: string[] = []

for (const path of htmlFiles) {
  const html = await readFile(path, 'utf8')
  const label = relative(dist, path)
  const title = /<title>([^<]+)<\/title>/.exec(html)?.[1]
  const description = /<meta name="description" content="([^"]+)"/.exec(html)?.[1]
  const canonical = /<link rel="canonical" href="([^"]+)"/.exec(html)?.[1]
  if (!title) problems.push(`${label}: missing title`)
  else if (titles.has(title)) problems.push(`${label}: duplicate title with ${titles.get(title)}`)
  else titles.set(title, label)
  if (!description) problems.push(`${label}: missing description`)
  else if (descriptions.has(description) && !label.includes('generated/')) problems.push(`${label}: duplicate description with ${descriptions.get(description)}`)
  else descriptions.set(description, label)
  if (!canonical?.startsWith(DEFAULT_SITE_URL)) problems.push(`${label}: canonical URL is not on the production origin`)
  if (!html.includes(PRODUCT_NAME)) problems.push(`${label}: metadata does not contain the differentiated product identity`)
  if (/<img(?![^>]*\balt=)[^>]*>/i.test(html)) problems.push(`${label}: image without alt text`)
  const info = await stat(path)
  if (info.size > 300_000) problems.push(`${label}: HTML exceeds 300 KiB`)
}

const jsFiles = files.filter((path) => extname(path) === '.js')
const jsSizes = await Promise.all(jsFiles.map(async (path) => ({ path, size: (await stat(path)).size })))
const firstPartyJsBytes = jsSizes.filter(({ path }) => !path.includes('/pagefind/')).reduce((sum, value) => sum + value.size, 0)
const largestJs = jsSizes.toSorted((left, right) => right.size - left.size)[0]
if (firstPartyJsBytes > 200_000) problems.push(`First-party JavaScript output exceeds 200 KiB (${firstPartyJsBytes} bytes)`)
if (largestJs && largestJs.size > 200_000) problems.push(`JavaScript asset ${relative(dist, largestJs.path)} exceeds 200 KiB (${largestJs.size} bytes)`)

for (const route of ['index.html', 'download/index.html', 'docs/index.html', 'community/index.html', 'changelog/index.html', 'llms.txt', 'llms-full.txt', 'sitemap-index.xml']) {
  if (!files.includes(resolve(dist, route))) problems.push(`missing required build output ${route}`)
}

const invocation = JSON.parse(await readFile(resolve(dist, '_routes.json'), 'utf8')) as { include: string[]; exclude: string[] }
for (const route of internalDocumentationInvocationRoutes) {
  if (!invocation.include.includes(route)) problems.push(`missing retired-document invocation route: ${route}`)
}
if (invocation.include.includes('/*') || invocation.include.length + invocation.exclude.length > 100) problems.push('Pages Worker invocation must remain bounded')

// Check publication output, not just navigation: removed documents must have no
// HTML, raw Markdown, sitemap entry, or source section in either AI export.
const internalSources = (await filesUnder(docsRoot)).filter((path) => /\.mdx?$/.test(path) && !isPublicDocumentation(relative(docsRoot, path)))
const sitemap = (await Promise.all(files.filter((path) => /sitemap.*\.xml$/.test(path)).map((path) => readFile(path, 'utf8')))).join('\n')
const exports = await Promise.all(['llms.txt', 'llms-full.txt'].map((name) => readFile(resolve(dist, name), 'utf8')))
for (const source of internalSources) {
  const route = `/docs/${relative(docsRoot, source).replace(/\.mdx?$/, '')}`
  for (const output of [`${route}/index.html`, `${route}.md`, `${route}.html`]) {
    if (files.includes(resolve(dist, output.slice(1)))) problems.push(`internal document published: ${output}`)
  }
  if (sitemap.includes(`${route}/`)) problems.push(`internal document in sitemap: ${route}`)
  const sourceText = await readFile(source, 'utf8')
  const title = /^title:\s*["']?(.+?)["']?$/m.exec(sourceText)?.[1]
  if (title && exports.some((content) => content.includes(`\n# ${title}\n`) || content.includes(`[${title}](`))) problems.push(`internal document content in AI export: ${title}`)
  if (exports.some((content) => content.includes(`${DEFAULT_SITE_URL}${route}`))) problems.push(`internal document in AI export: ${route}`)
}

if (problems.length > 0) throw new Error(`Build validation failed:\n- ${problems.join('\n- ')}`)
console.log(`Validated ${htmlFiles.length} HTML pages; first-party JavaScript is ${Math.round(firstPartyJsBytes / 1024)} KiB and every asset is below 200 KiB.`)
