import { createRequire } from 'node:module'
import { readFile, readdir } from 'node:fs/promises'
import { extname, join } from 'node:path'

const { zipSync } = createRequire(new URL('../site/package.json', import.meta.url))('fflate')
const topLevel = ['README.md', 'icons.ts', 'palette.json', 'preview.png', 'preview.svg', 'social.svg']
const directories = ['app', 'interface', 'logos', 'marks', 'tray', 'web']
const assetExtensions = new Set(['.svg', '.png', '.ico', '.icns'])

async function assets(directory, prefix = '') {
  const result = []
  for (const entry of await readdir(directory, { withFileTypes: true })) {
    const path = join(directory, entry.name)
    const name = prefix + entry.name
    if (entry.isDirectory()) result.push(...await assets(path, name + '/'))
    else if (entry.isFile() && assetExtensions.has(extname(name))) result.push([name, await readFile(path)])
  }
  return result
}

/** Build a public asset kit, independently of internal reference files and mtimes. */
export async function createPublicBrandArchive(kit, license, notice) {
  const entries = await Promise.all(topLevel.map(async (name) => [name, await readFile(join(kit, name))]))
  const readme = entries.find(([name]) => name === 'README.md')[1].toString('utf8')
  if (/\bphases?\b|references\/|generation prompt|creation notes/i.test(readme)) {
    throw new Error('Public brand guide contains internal planning or creation notes')
  }
  for (const directory of directories) {
    entries.push(...(await assets(join(kit, directory), directory + '/')))
  }
  entries.push(['LICENSE', license], ['NOTICE', notice])
  entries.sort(([left], [right]) => left < right ? -1 : left > right ? 1 : 0)
  // ZIP stores local DOS calendar fields. Construct the same fields in every TZ.
  return Buffer.from(zipSync(Object.fromEntries(entries.map(([name, bytes]) => [`envarium-v1/${name}`, bytes])), {
    level: 9, mtime: new Date(1980, 0, 1), os: 3, attrs: 0o100644 << 16,
  }))
}
