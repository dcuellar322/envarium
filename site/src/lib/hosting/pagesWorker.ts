import type { UpdaterFeed } from '../releases/updater'

/** Bounded hosting responses: no runtime credentials, external calls or storage. */
export function pagesWorkerSource(feed: UpdaterFeed | null, retiredPaths: readonly string[] = []): string {
  return `const feed = ${JSON.stringify(feed)};
const retiredPaths = ${JSON.stringify(retiredPaths)};
export default {
  async fetch(request, env) {
    const pathname = new URL(request.url).pathname;
    if (retiredPaths.some((path) => pathname === path || pathname.startsWith(path + '/') || pathname === path + '.md' || pathname === path + '.html')) {
      const page = await env.ASSETS.fetch(new Request(new URL('/404.html', request.url), { method: request.method === 'HEAD' ? 'HEAD' : 'GET' }));
      const headers = new Headers(page.headers);
      headers.set('Cache-Control', 'no-store');
      headers.set('X-Robots-Tag', 'noindex');
      return new Response(request.method === 'HEAD' ? null : page.body, { status: 404, headers });
    }
    if (pathname !== '/updates/stable.json') return env.ASSETS.fetch(request);
    const headers = { 'Cache-Control': 'no-store', 'X-Content-Type-Options': 'nosniff' };
    if (request.method !== 'GET' && request.method !== 'HEAD') {
      return new Response(null, { status: 405, headers: { ...headers, Allow: 'GET, HEAD' } });
    }
    if (feed === null) return new Response(null, { status: 204, headers });
    return new Response(request.method === 'HEAD' ? null : JSON.stringify(feed), {
      headers: { ...headers, 'Content-Type': 'application/json' }
    });
  }
};
`
}
