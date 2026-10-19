import { readFile } from 'node:fs/promises'
import { preview } from 'astro'

// Keep the real Astro preview server in Playwright's process group. The CLI
// detaches automatically in agent environments, preventing reliable cleanup.
const headersFile = await readFile(new URL('../public/_headers', import.meta.url), 'utf8')
const contentSecurityPolicy = /^\s+Content-Security-Policy:\s*(.+)$/m.exec(headersFile)?.[1]
if (!contentSecurityPolicy) throw new Error('Public site Content-Security-Policy is missing')
// WebKit upgrades loopback subresources to HTTPS under this transport directive.
// The preview is HTTP; keep every script/worker restriction from production.
const previewSecurityPolicy = contentSecurityPolicy.replace(/;\s*upgrade-insecure-requests\b/g, '')
const server = await preview({
  server: {
    host: '127.0.0.1', port: 4322, open: false,
    headers: { 'Content-Security-Policy': previewSecurityPolicy },
  },
})
let stopping = false
for (const signal of ['SIGINT', 'SIGTERM'] as const) {
  process.once(signal, () => {
    if (stopping) return
    stopping = true
    void server.stop().then(() => process.exit(0)).catch((error: unknown) => {
      console.error(error)
      process.exit(1)
    })
  })
}
