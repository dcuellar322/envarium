export const PRODUCT_NAME = 'Envarium — Local Development Command Center'
export const SHORT_NAME = 'Envarium'
export const REPOSITORY = 'dcuellar322/envarium' as const
export const REPOSITORY_URL = `https://github.com/${REPOSITORY}`
export const DEFAULT_SITE_URL = 'https://envarium.davidcuellar.tech'
export const SITE_DESCRIPTION =
  'An open-source, local-first command center for Docker Compose, native processes, Git, ports, logs, terminals, and coding agents.'

function parseSiteUrl(value: string): URL {
  const url = new URL(value)
  if (url.protocol !== 'https:') {
    throw new Error(`SITE_URL must use HTTPS, received ${url.protocol}`)
  }
  url.pathname = '/'
  url.search = ''
  url.hash = ''
  return url
}

export const siteUrl = parseSiteUrl(process.env.SITE_URL ?? DEFAULT_SITE_URL)
export const isPreview =
  process.env.SITE_PREVIEW === '1' ||
  (process.env.CF_PAGES === '1' && process.env.CF_PAGES_BRANCH !== 'master')

export function canonicalUrl(path = '/'): URL {
  return new URL(path.replace(/^\/?/, '/'), siteUrl)
}
