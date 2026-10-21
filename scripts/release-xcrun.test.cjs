const assert = require('node:assert/strict')
const test = require('node:test')
const { releaseXcrun } = require('./release-xcrun.cjs')

const staple = ['stapler', 'staple', '-v', 'Envarium Release.app']
const validate = ['stapler', 'validate', 'Envarium Release.app']

test('delegates SDK/notary commands and preserves their status without retrying', () => {
  for (const args of [['--show-sdk-path'], ['notarytool', 'submit', 'Envarium.zip']]) {
    const calls = []
    assert.equal(releaseXcrun(args, (values) => {
      calls.push(values)
      return { status: 7 }
    }, () => assert.fail('unexpected retry')), 7)
    assert.deepEqual(calls, [args])
  }
})

test('requires ticket validation before returning successful stapling to the bundler', () => {
  const calls = []
  assert.equal(releaseXcrun(staple, (args) => {
    calls.push(args)
    return { status: 0 }
  }, () => assert.fail('unexpected retry')), 0)
  assert.deepEqual(calls, [staple, validate])
})

test('retries failed stapling and failed validation with a bounded delay', () => {
  for (const failed of ['staple', 'validate']) {
    let attempts = 0
    let pauses = 0
    assert.equal(releaseXcrun(staple, (args) => {
      if (args[1] === failed && attempts++ === 0) return { status: 65 }
      return { status: 0 }
    }, () => pauses++), 0)
    assert.equal(attempts, 2)
    assert.equal(pauses, 1)
  }
})

test('persistent missing tickets cannot become successful and cancellation stops retries', () => {
  let attempts = 0
  let pauses = 0
  assert.equal(releaseXcrun(staple, () => {
    attempts++
    return { status: 65 }
  }, () => pauses++), 1)
  assert.equal(attempts, 3)
  assert.equal(pauses, 2)
  for (const result of [{ status: null, signal: 'SIGTERM' }, { status: null, error: new Error('spawn failed') }]) {
    assert.equal(releaseXcrun(staple, () => result, () => assert.fail('unexpected retry')), 1)
  }
})


test('independent validation retries only transient ticket-service failure and stays fail closed', () => {
  let calls = 0
  let pauses = 0
  assert.equal(releaseXcrun(validate, () => ({ status: calls++ === 0 ? 68 : 0 }), () => pauses++), 0)
  assert.equal(calls, 2)
  assert.equal(pauses, 1)
  calls = 0
  assert.equal(releaseXcrun(validate, () => { calls++; return { status: 68 } }, () => {}), 68)
  assert.equal(calls, 3)
  assert.equal(releaseXcrun(validate, () => ({ status: 65 }), () => assert.fail('missing ticket must not retry')), 65)
  assert.equal(releaseXcrun(validate, () => ({ status: null, signal: 'SIGTERM' }), () => assert.fail('cancellation must not retry')), 1)
})
