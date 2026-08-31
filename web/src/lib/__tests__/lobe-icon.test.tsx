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
import { render } from '@testing-library/react'
import { describe, expect, it, vi } from 'vitest'

vi.mock('@lobehub/icons', () => ({}))

const { getLobeIcon } = await import('../lobe-icon')

describe('getLobeIcon Xiaomi Mimo', () => {
  it('renders the bundled MiMo SVG for the XiaomiMiMo key', () => {
    const { container } = render(<>{getLobeIcon('XiaomiMiMo', 14)}</>)
    const svg = container.querySelector('svg')
    expect(svg).not.toBeNull()
    expect(svg).toHaveAttribute('viewBox', '0 0 24 24')
    expect(svg).toHaveAttribute('width', '14')
    expect(svg).toHaveAttribute('height', '14')
    expect(svg).toHaveAttribute('aria-hidden', 'true')
  })

  it('renders a public image path as an img', () => {
    const { container } = render(<>{getLobeIcon('/xiaomimimo.svg', 28)}</>)
    const img = container.querySelector('img')
    expect(img).not.toBeNull()
    expect(img).toHaveAttribute('src', '/xiaomimimo.svg')
    expect(img).toHaveAttribute('width', '28')
  })
})
