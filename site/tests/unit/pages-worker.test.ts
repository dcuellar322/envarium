import { describe, expect, it } from 'vitest'
import { pagesWorkerSource } from '../../src/lib/hosting/pagesWorker'
import { internalDocumentationPaths, internalDocumentationInvocationRoutes } from '../../src/lib/docs/publication'

async function worker() {
  return (await import(/* @vite-ignore */ `data:text/javascript;base64,${Buffer.from(pagesWorkerSource(null, internalDocumentationPaths)).toString('base64')}`)).default
}

describe('retired documentation at the Pages edge', () => {
  it('returns uncached 404s for every retired path and alias even with stale assets', async () => {
    const handler = await worker()
    for (const path of internalDocumentationPaths) {
      for (const suffix of ['', '/', '/index.html', '.html', '.md', '/future-report.md']) {
        const requested: string[] = []
        const response = await handler.fetch(new Request(`https://example.org${path}${suffix}?search=test`), {
          ASSETS: { fetch: (request: Request) => {
            requested.push(new URL(request.url).pathname)
            return new Response(requested[0] === '/404.html' ? 'Page not found' : 'stale internal report')
          } },
        })
        expect(requested).toEqual(['/404.html'])
        expect(response.status).toBe(404)
        expect(await response.text()).toBe('Page not found')
        expect(response.headers.get('Cache-Control')).toBe('no-store')
        expect(response.headers.get('X-Robots-Tag')).toBe('noindex')
      }
      for (const suffix of ['', '/*', '.md', '.html']) expect(internalDocumentationInvocationRoutes).toContain(`${path}${suffix}`)
    }
    expect(internalDocumentationInvocationRoutes.length + 1).toBeLessThanOrEqual(100)
  })
  it('handles HEAD and delegates public paths without swallowing asset errors', async () => {
    const handler = await worker()
    const assets = { fetch: () => new Response('public page') }
    const head = await handler.fetch(new Request('https://example.org/docs/progress/phase-04/', { method: 'HEAD' }), { ASSETS: assets })
    expect(head.status).toBe(404)
    expect(await head.text()).toBe('')
    for (const path of ['/docs/cli/', '/docs/progress-guide/', '/docs/public-site-guide/']) {
      expect(await (await handler.fetch(new Request(`https://example.org${path}`), { ASSETS: assets })).text()).toBe('public page')
    }
    await expect(handler.fetch(new Request('https://example.org/docs/progress/phase-04/'), {
      ASSETS: { fetch: () => Promise.reject(new Error('Asset unavailable')) },
    })).rejects.toThrow('Asset unavailable')
  })
})
