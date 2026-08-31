import { describe, expect, it } from 'vitest'

import {
  PASSWORD_MAX_LENGTH,
  PASSWORD_MIN_LENGTH,
  registerFormSchema,
} from '../constants'

describe('register password length', () => {
  it('accepts a grouped 23-character password', () => {
    const result = registerFormSchema.safeParse({
      username: 'alice',
      password: 'abcde-fghij-klmno-pqrst',
      confirmPassword: 'abcde-fghij-klmno-pqrst',
    })
    expect(result.success).toBe(true)
  })

  it('rejects a password longer than the maximum', () => {
    const tooLong = 'a'.repeat(PASSWORD_MAX_LENGTH + 1)
    const result = registerFormSchema.safeParse({
      username: 'alice',
      password: tooLong,
      confirmPassword: tooLong,
    })
    expect(result.success).toBe(false)
  })

  it('rejects a password shorter than the minimum', () => {
    const tooShort = 'a'.repeat(PASSWORD_MIN_LENGTH - 1)
    const result = registerFormSchema.safeParse({
      username: 'alice',
      password: tooShort,
      confirmPassword: tooShort,
    })
    expect(result.success).toBe(false)
  })
})
