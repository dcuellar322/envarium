#!/usr/bin/env node
// Release-only workaround: validate the ticket before Tauri packages the app.
const { spawnSync } = require('node:child_process')
const { appendFileSync } = require('node:fs')
const { join } = require('node:path')

function nativeXcrun(args) {
  if (args[0] !== 'stapler')
    return spawnSync('/usr/bin/xcrun', args, { stdio: 'inherit' })
  const result = spawnSync('/usr/bin/xcrun', args, { encoding: 'utf8' })
  if (result.stdout) process.stdout.write(result.stdout)
  if (result.stderr) process.stderr.write(result.stderr)
  // Tauri captures and discards stapler output. Keep only ticket-tool diagnostics,
  // never notarization commands or credentials, for a failed build's next step.
  if (process.env.RUNNER_TEMP) {
    appendFileSync(join(process.env.RUNNER_TEMP, 'envarium-stapler.log'),
      `${JSON.stringify(args)}: exit ${result.status}\n${result.stdout ?? ''}${result.stderr ?? ''}\n`,
      { mode: 0o600 })
  }
  return result
}

function releaseXcrun(
  args,
  run = nativeXcrun,
  pause = () => spawnSync('/bin/sleep', ['5'], { stdio: 'inherit' }),
) {
  const status = (result) => result.status ?? 1
  if (args[0] !== 'stapler' || args[1] !== 'staple' || args.length < 3)
    return status(run(args))

  const path = args.at(-1)
  for (let attempt = 0; attempt < 3; attempt++) {
    const stapled = run(args)
    if (stapled.signal || stapled.error) return status(stapled)
    if (stapled.status === 0) {
      const validated = run(['stapler', 'validate', path])
      if (validated.status === 0) return 0
      if (validated.signal || validated.error) return status(validated)
    }
    if (attempt < 2) pause()
  }
  return 1
}

module.exports = { releaseXcrun }
if (require.main === module) process.exitCode = releaseXcrun(process.argv.slice(2))
