import { copyFile, mkdir } from 'node:fs/promises'
import { resolve } from 'node:path'
import { repositoryRoot, siteRoot } from './paths'

// These are synthetic, sanitized fixtures captured by the real Vue app's
// visual suite. Publishing uses neutral names rather than a host OS suffix.
const platform = process.argv.find((argument) => argument.startsWith('--platform='))?.split('=')[1] ?? process.platform
if (platform !== 'darwin' && platform !== 'linux') throw new Error('Select a reviewed darwin or linux visual capture with --platform=')
const shots = [
  ['dashboard.spec.ts', 'dashboard-alpha', 'dashboard'],
  ['project.spec.ts', 'project-alpha', 'project'],
  ['project.spec.ts', 'project-terminal', 'terminal'],
  ['ports.spec.ts', 'port-registry', 'ports'],
  ['workspaces.spec.ts', 'workspace-progress', 'workspaces'],
  ['plugins.spec.ts', 'plugin-permission-review', 'plugins'],
  ['diagnostics.spec.ts', 'diagnostic-automation-review', 'diagnostics'],
] as const
const destination = resolve(siteRoot, 'src/assets/product')
await mkdir(destination, { recursive: true })
for (const [spec, capture, name] of shots) {
  await copyFile(
    resolve(repositoryRoot, `web/tests/visual/${spec}-snapshots/${capture}-${platform}.png`),
    resolve(destination, `${name}.png`),
  )
}
console.log(`Published ${shots.length} reviewed ${platform} fixture screenshots to site/src/assets/product.`)
