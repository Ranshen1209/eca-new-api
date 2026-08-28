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
import { render, screen } from '@testing-library/react'
import { describe, expect, test } from 'vitest'

import { EcaOpticalFlow } from '../optical-flow'

describe('EcaOpticalFlow', () => {
  test('is a decorative wash without oval path-progress strokes', () => {
    render(<EcaOpticalFlow />)

    const root = document.querySelector('.eca-optical-flow')
    expect(root).not.toBeNull()
    expect(root?.getAttribute('aria-hidden')).toBe('true')
    expect(root?.querySelectorAll('path')).toHaveLength(0)
    expect(root?.querySelectorAll('.eca-flow-dash, .eca-flow-dash-gold')).toHaveLength(
      0
    )
    expect(screen.queryByRole('img')).toBeNull()
  })
})
