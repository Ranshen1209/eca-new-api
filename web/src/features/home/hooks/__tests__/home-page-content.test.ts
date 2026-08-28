/*
Copyright (C) 2023-2026 QuantumNous

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as
published by the Free Software Foundation, either version 3 of the
License, or (at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program. If not, see <https://www.gnu.org/licenses/>.

For commercial licensing, please contact support@quantumnous.com
*/
import { renderHook, waitFor } from '@testing-library/react'
import { beforeEach, describe, expect, test, vi } from 'vitest'

const { toast } = vi.hoisted(() => ({
  toast: { error: vi.fn() },
}))

vi.mock('sonner', () => ({ toast }))

vi.mock('../../api', () => ({
  getHomePageContent: vi.fn(),
}))

const { getHomePageContent } = await import('../../api')
const { useHomePageContent } = await import('../use-home-page-content')

describe('useHomePageContent', () => {
  beforeEach(() => {
    toast.error.mockReset()
    vi.mocked(getHomePageContent).mockReset()
    localStorage.removeItem('home_page_content')
  })

  test('falls back to the default landing without a toast when homepage content cannot load', async () => {
    vi.mocked(getHomePageContent).mockRejectedValue(
      new Error('Request failed with status code 504')
    )

    const { result } = renderHook(() => useHomePageContent())

    await waitFor(() => {
      expect(result.current.isLoaded).toBe(true)
    })

    expect(result.current.content).toBe('')
    expect(toast.error).not.toHaveBeenCalled()
  })
})
