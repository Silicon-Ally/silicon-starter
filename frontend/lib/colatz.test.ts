import { colatzCount } from './colatz'
import { describe, expect, it } from '@jest/globals'

// An example test, since we don't use Jest in any of the fairly trivial code
// in our libraries.
describe('Colatz', () => {
  it('handles negative numbers', () => {
    expect(() => colatzCount(-1)).toThrowError(/negative number/)
  })
  it('handles non-ints', () => {
    expect(() => colatzCount(1.1)).toThrowError(/real number/)
  })
  it('base case', () => {
    expect(colatzCount(1)).toBe(0)
  })
  it('four yields 2', () => {
    expect(colatzCount(4)).toBe(2)
  })
  it('six yields 8', () => {
    expect(colatzCount(6)).toBe(8)
  })
  it('eleven yields 14', () => {
    expect(colatzCount(11)).toBe(14) 
  })
})