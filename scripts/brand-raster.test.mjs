import assert from 'node:assert/strict'
import test from 'node:test'
import { equalNativePixels } from './brand-raster.mjs'

test('native rounding accepts at most one channel differing by one intensity level', () => {
  assert.equal(equalNativePixels(Buffer.from([0, 128, 255]), Buffer.from([0, 128, 255])), true)
  assert.equal(equalNativePixels(Buffer.from([0, 128, 255]), Buffer.from([0, 127, 255])), true)
  assert.equal(equalNativePixels(Buffer.from([0, 128, 255]), Buffer.from([0, 129, 255])), true)
  assert.equal(equalNativePixels(Buffer.from([0, 128, 255]), Buffer.from([0, 126, 255])), false)
  assert.equal(equalNativePixels(Buffer.from([0, 128, 255]), Buffer.from([1, 127, 255])), false)
  assert.equal(equalNativePixels(Buffer.from([0, 128, 255]), Buffer.from([0, 128])), false)
})
