import { describe, expect, it } from 'vitest'

import { stringFormat } from './string-format'

describe('stringFormat', () => {
  it('replaces sequential %s placeholders', () => {
    expect(stringFormat('Hello %s from %s', 'world', 'Spire')).toBe('Hello world from Spire')
  })

  it('keeps unmatched placeholders when values are missing', () => {
    expect(stringFormat('Value: %s %s', 'one')).toBe('Value: one %s')
  })

  it('appends extra values like util.format', () => {
    expect(stringFormat('%s', 'one', 'two', 3)).toBe('one two 3')
  })

  it('supports escaped percent signs and stringifies non-string values', () => {
    expect(stringFormat('Load %% %s %s', 50, null)).toBe('Load % 50 null')
  })
})
