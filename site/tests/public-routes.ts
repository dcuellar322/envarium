import { readdir } from 'node:fs/promises'
import { resolve } from 'node:path'

export async function publicRoutes(directory = resolve('dist'), prefix = ''): Promise<string[]> {
  const entries = await readdir(directory, { withFileTypes: true })
  return (await Promise.all(entries.map(async (entry) => {
    const path = `${prefix}/${entry.name}`
    if (entry.isDirectory()) return publicRoutes(resolve(directory, entry.name), path)
    return entry.name.endsWith('.html') ? [path.replace(/index\.html$/, '')] : []
  }))).flat().sort()
}
