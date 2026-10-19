import { DEFAULT_SITE_URL, PRODUCT_NAME } from '../src/config/site'
import { authenticateUpdaterSignature } from '../src/lib/releases/updaterSignature'
import { updaterPublicKey } from '../src/lib/releases/updaterConfig'
import { updaterTargets, type UpdaterFeed } from '../src/lib/releases/updater'

const origin = process.env.SMOKE_SITE_URL ?? DEFAULT_SITE_URL
const canonicalOrigin = process.env.SMOKE_CANONICAL_URL ?? origin
const routes = [
  '/',
  '/download/',
  '/docs/',
  '/docs/start/',
  '/changelog/',
  '/community/',
  '/security/',
]
const problems: string[] = []
for (const route of routes) {
  const url = new URL(route, origin)
  const response = await fetch(url, { redirect: 'follow' })
  const body = await response.text()
  if (!response.ok) problems.push(`${url.href}: HTTP ${response.status}`)
  if (response.url !== url.href) problems.push(`${url.href}: resolved to ${response.url}`)
  if (!body.includes(PRODUCT_NAME))
    problems.push(`${url.href}: differentiated product identity missing`)
  const canonical = new URL(route, canonicalOrigin)
  if (!body.includes(`<link rel="canonical" href="${canonical.href}"`))
    problems.push(`${url.href}: canonical link missing or incorrect`)
}
const updater = await fetch(new URL('/updates/stable.json', origin))
if (updater.status !== 204) {
  if (updater.status !== 200) problems.push(`Updater feed returned HTTP ${updater.status}`)
  else {
    const feed = await updater.json() as UpdaterFeed
    for (const target of updaterTargets) {
      const platform = feed.platforms?.[target]
      if (!platform?.url.startsWith(`https://github.com/dcuellar322/envarium/releases/download/v${feed.version}/`))
        problems.push(`Updater feed has missing or foreign asset for ${target}`)
      try { authenticateUpdaterSignature(platform?.signature ?? '', updaterPublicKey) }
      catch { problems.push(`Updater feed has invalid signature for ${target}`) }
    }
  }
}
if (problems.length > 0) throw new Error(`Production smoke failed:\n- ${problems.join('\n- ')}`)
console.log(
  `Production smoke passed for ${routes.length} pages and the updater feed at ${origin} (canonical: ${canonicalOrigin}).`,
)
