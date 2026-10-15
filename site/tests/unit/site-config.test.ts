import { describe, expect, it } from 'vitest'
import { canonicalUrl, DEFAULT_SITE_URL, PRODUCT_NAME, REPOSITORY } from '../../src/config/site'

describe('canonical site configuration', () => {
  it('uses one differentiated identity and production origin', () => {
    expect(PRODUCT_NAME).toBe('Envarium — Local Development Command Center')
    expect(REPOSITORY).toBe('dcuellar322/envarium')
    expect(canonicalUrl('/docs/start/').href).toBe(`${DEFAULT_SITE_URL}/docs/start/`)
  })
})
