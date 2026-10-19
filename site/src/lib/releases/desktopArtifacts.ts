import { classifyReleaseAsset } from './artifactContract'

export function desktopArtifactName(name: string, platform: string, version: string): string | undefined {
  if (!/^(macos-(arm64|amd64)|linux-amd64|windows-amd64)$/.test(platform))
    throw new Error('Unknown desktop artifact platform')
  if (name === 'Envarium.app.tar.gz' || name === 'Envarium.app.tar.gz.sig') {
    if (!platform.startsWith('macos-')) throw new Error('Mac updater archive on another platform')
    return `Envarium_${version}_${platform.endsWith('arm64') ? 'aarch64' : 'x64'}.app.tar.gz${name.endsWith('.sig') ? '.sig' : ''}`
  }
  const identity = classifyReleaseAsset(name.replace(/\.sig$/, ''))
  if (!identity || identity.packageType === 'tar.gz' || identity.packageType === 'zip') return undefined
  const expectedPlatform = platform.startsWith('macos-') ? 'macos' : platform.split('-')[0]
  if (identity.platform !== expectedPlatform || identity.architecture !== platform.split('-')[1])
    throw new Error(`Desktop artifact does not match its build platform: ${name}`)
  return name
}
