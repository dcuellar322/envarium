import { appendFile, readFile, writeFile } from 'node:fs/promises'
import { resolve } from 'node:path'
import { isPreview } from '../src/config/site'
import { siteRoot } from './paths'
import type { UpdaterFeed } from '../src/lib/releases/updater'
import { updaterWorkerSource } from '../src/lib/releases/updaterWorker'

const feed = JSON.parse(await readFile(resolve(siteRoot, 'src/data/updater-manifest.json'), 'utf8')) as UpdaterFeed | null
await writeFile(resolve(siteRoot, 'dist/_worker.js'), updaterWorkerSource(isPreview ? null : feed))
await writeFile(resolve(siteRoot, 'dist/_routes.json'), JSON.stringify({ version: 1, include: ['/updates/stable.json'], exclude: [] }))

if (isPreview) {
  await appendFile(
    resolve(siteRoot, 'dist/_headers'),
    '\n/*\n  X-Robots-Tag: noindex, nofollow\n  Cache-Control: no-store\n',
    'utf8',
  )
}
