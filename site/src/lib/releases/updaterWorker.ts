import type { UpdaterFeed } from './updater'

/** A generated, immutable catalog: no runtime credentials, GitHub calls or storage. */
export function updaterWorkerSource(feed: UpdaterFeed | null): string {
  return `const feed = ${JSON.stringify(feed)};
export default {
  async fetch(request, env) {
    if (new URL(request.url).pathname !== '/updates/stable.json') return env.ASSETS.fetch(request);
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
