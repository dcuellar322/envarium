import { expect, test } from 'vitest'
import { presentationLabel } from '../../src/api/presentation'

test('local request labels contain only client and coarse page categories', () => {
  expect(presentationLabel('/', true)).toBe('native:dashboard')
  expect(presentationLabel('/projects/private-project-id', false)).toBe('browser:projects')
  expect(presentationLabel('/logs', true)).toBe('native:logs')
  expect(presentationLabel('/private-path', true)).toBe('native:other')
})
