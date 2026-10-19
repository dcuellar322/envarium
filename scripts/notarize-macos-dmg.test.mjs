import { mkdtempSync, rmSync, symlinkSync, writeFileSync } from 'node:fs';
import { tmpdir } from 'node:os';
import { join } from 'node:path';
import test from 'node:test';
import assert from 'node:assert/strict';
import { notarizeMacDiskImages } from './notarize-macos-dmg.mjs';

const env = { APPLE_API_KEY_PATH: '/runner/temp/notarization.p8', APPLE_API_KEY: 'ABCDEFGHIJ', APPLE_API_ISSUER: '12345678-1234-1234-1234-123456789abc' };
test('notarizes the signed disk image, then validates its ticket and Gatekeeper assessment', () => {
  const directory = mkdtempSync(join(tmpdir(), 'envarium-dmg-'));
  try {
    writeFileSync(join(directory, 'Envarium_x64.dmg'), 'fixture');
    const calls = [];
    const run = (command, args) => {
      calls.push([command, args]);
      return args[0] === 'notarytool' ? JSON.stringify({ status: 'Accepted', id: 'submission' }) : '';
    };
    assert.deepEqual(notarizeMacDiskImages(directory, env, run), [{ name: 'Envarium_x64.dmg', id: 'submission', status: 'Accepted' }]);
    assert.deepEqual(calls.map(([command]) => command), ['codesign', 'xcrun', 'xcrun', 'xcrun', 'spctl', 'hdiutil']);
    assert.ok(calls[1][1].includes('--wait'));
    assert.ok(calls[1][1].includes(env.APPLE_API_KEY_PATH));
    assert.ok(calls[4][1].includes('context:primary-signature'));
    let ticketAttempts = 0;
    notarizeMacDiskImages(directory, env, (command, args) => {
      if (args[0] === 'stapler' && args[1] === 'validate' && ticketAttempts++ === 0)
        throw new Error('transient ticket service failure');
      return run(command, args);
    });
    assert.equal(ticketAttempts, 2);
    assert.throws(() => notarizeMacDiskImages(directory, env, (command, args) => {
      if (args[0] === 'stapler') throw new Error('persistent ticket failure');
      return run(command, args);
    }), /persistent ticket failure/);
    assert.throws(() => notarizeMacDiskImages(directory, env, () => '{"status":"Invalid"}'), /did not accept/);
    assert.throws(() => notarizeMacDiskImages(directory, env, () => { throw new Error('cancelled'); }), /cancelled/);
    assert.throws(() => notarizeMacDiskImages(directory, {}, run), /configuration/);
    rmSync(join(directory, 'Envarium_x64.dmg'));
    assert.throws(() => notarizeMacDiskImages(directory, env, run), /exactly one/);
    symlinkSync('/dev/null', join(directory, 'Envarium_x64.dmg'));
    assert.throws(() => notarizeMacDiskImages(directory, env, run), /regular file/);
  } finally { rmSync(directory, { recursive: true, force: true }); }
});
