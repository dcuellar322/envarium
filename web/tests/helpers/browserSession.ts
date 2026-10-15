import { execFileSync } from 'node:child_process'

export function browserBootstrapPath(pathname = '/'): string {
  const daemonAddress = process.env.ENVARIUM_E2E_DAEMON_ADDRESS ?? '127.0.0.1:29616'
  const output = execFileSync(
    '../.cache/envarium-e2e',
    ['--address', daemonAddress, '--data-dir', '../.envarium-data/e2e', 'ui'],
    { encoding: 'utf8' },
  ).trim()
  const token = new URL(output).searchParams.get('bootstrap')
  if (!token) throw new Error('envarium ui did not return a browser bootstrap token')
  return `${pathname}?bootstrap=${encodeURIComponent(token)}`
}
