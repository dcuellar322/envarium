import { preview } from 'astro'

// Keep the real Astro preview server in Playwright's process group. The CLI
// detaches automatically in agent environments, preventing reliable cleanup.
const server = await preview({ server: { host: '127.0.0.1', port: 4322, open: false } })
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
