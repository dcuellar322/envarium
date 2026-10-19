import { execFileSync } from 'node:child_process';
import { lstatSync, readdirSync } from 'node:fs';
import { resolve } from 'node:path';
import { fileURLToPath } from 'node:url';

export function notarizeMacDiskImages(directory, env = process.env, run = (command, args) =>
  execFileSync(command, args, { encoding: 'utf8', timeout: 25 * 60_000, maxBuffer: 4 * 1024 * 1024 })) {
  if (!env.APPLE_API_KEY_PATH || !/^[A-Z0-9]{10}$/.test(env.APPLE_API_KEY ?? '') ||
      !/^[0-9a-f-]{36}$/i.test(env.APPLE_API_ISSUER ?? ''))
    throw new Error('Apple API signing configuration is missing or invalid');
  const images = readdirSync(directory).filter((name) => name.endsWith('.dmg'));
  if (images.length !== 1) throw new Error('Expected exactly one Mac disk image per architecture');
  const receipts = [];
  // Apple's ticket service can briefly fail after an accepted submission.
  // Retry bounded read/staple checks; persistent failures still fail the job.
  const ticket = (action, path) => {
    for (let attempt = 0; attempt < 3; attempt++) {
      try { run('xcrun', ['stapler', action, path]); return; }
      catch (error) {
        if (attempt === 2) throw error;
        run('sleep', ['5']);
      }
    }
  };
  for (const name of images) {
    const path = resolve(directory, name);
    if (!lstatSync(path).isFile()) throw new Error('Mac disk image must be a regular file');
    run('codesign', ['--verify', '--strict', path]);
    const receipt = JSON.parse(run('xcrun', ['notarytool', 'submit', path,
      '--key', env.APPLE_API_KEY_PATH, '--key-id', env.APPLE_API_KEY,
      '--issuer', env.APPLE_API_ISSUER, '--wait', '--timeout', '20m', '--output-format', 'json']));
    if (receipt.status !== 'Accepted') throw new Error('Apple did not accept the Mac disk image');
    ticket('staple', path);
    ticket('validate', path);
    run('spctl', ['--assess', '--type', 'open', '--context', 'context:primary-signature', '--verbose=2', path]);
    run('hdiutil', ['verify', path]);
    receipts.push({ name, id: receipt.id, status: receipt.status });
  }
  return receipts;
}

if (process.argv[1] && resolve(process.argv[1]) === fileURLToPath(import.meta.url)) {
  console.log(JSON.stringify(notarizeMacDiskImages(
    resolve('desktop/src-tauri/target/release/bundle/dmg'),
  )));
}
