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

import { EcaAboutPage } from '../about-page'
import { ECA_NEW_API_HREF } from '../copy'
import { ECA_MARK_HREF } from '../favicon'

describe('EcaAboutPage', () => {
  test('uses a branded card with stacked panels that become two columns on md', () => {
    const { container } = render(<EcaAboutPage language='zhCN' />)

    const page = container.querySelector('[data-eca-about="page"]')
    expect(page).not.toBeNull()
    expect(page?.getAttribute('data-brand')).toBe('eca')
    expect(page?.className).toContain('min-h-svh')
    expect(page?.className).toContain('flex')
    expect(page?.className).toContain('relative')
    expect(page?.querySelector('.eca-optical-flow')).not.toBeNull()

    const stage = page?.querySelector('[data-eca-about="stage"]')
    expect(stage?.className).toContain('min-h-svh')
    expect(stage?.className).toContain('flex-1')
    expect(stage?.className).toContain('items-center')
    expect(stage?.className).toContain('justify-center')
    expect(stage?.className).toContain('pt-24')

    const article = page?.querySelector('article')
    expect(article?.className).toContain('max-w-5xl')
    expect(article?.className).toContain('w-full')
    expect(article?.className).toContain('overflow-hidden')

    const header = article?.querySelector('header')
    expect(header?.className).toContain('flex')
    expect(header?.className).toContain('items-start')
    expect(header?.querySelector('p')?.className).toContain('flex-1')
    expect(header?.querySelector('p')?.className).toContain('min-w-0')
    expect(article?.querySelector('h1')).toBeNull()

    const panels = container.querySelector('[data-eca-about="panels"]')
    expect(panels?.className).toContain('grid')
    expect(panels?.className).toContain('grid-cols-1')
    expect(panels?.className).toContain('md:grid-cols-2')
    expect(panels?.querySelectorAll('section.min-w-0')).toHaveLength(2)
    expect(panels?.querySelector('section')?.className).toContain('min-h-36')
  })

  test('shows Chinese service copy, IT specialist support, and New API secondary-development credit', () => {
    const { container } = render(<EcaAboutPage language='zhCN' />)

    expect(screen.queryByRole('heading', { level: 1 })).toBeNull()
    expect(screen.queryByText('关于 ECA API Gateway')).toBeNull()
    expect(
      screen.getByText(
        '此站点仅为 ECA Healthcare 提供服务，不对外提供公开访问。'
      )
    ).toBeTruthy()
    expect(
      screen.getByText('如需开通或技术支持，请联系公司信息化专员。')
    ).toBeTruthy()
    expect(screen.getByText(/二次开发/)).toBeTruthy()

    const credit = screen.getByRole('link', { name: 'New API' })
    expect(credit.getAttribute('href')).toBe(ECA_NEW_API_HREF)
    expect(credit.getAttribute('target')).toBe('_blank')
    expect(container.querySelector('img')?.getAttribute('src')).toBe(
      ECA_MARK_HREF
    )
  })
})
