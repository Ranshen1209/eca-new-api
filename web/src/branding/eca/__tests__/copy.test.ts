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
import { readFileSync } from 'node:fs'
import path from 'node:path'
import { fileURLToPath } from 'node:url'

import i18n from 'i18next'
import { afterEach, describe, expect, test } from 'vitest'

import {
  APP_SUPPORT_I18N_KEY,
  applyEcaLandingCopy,
  ECA_ABOUT_COPY,
  ECA_CC_SWITCH_DEFAULT_NAMES,
  ECA_CC_SWITCH_HREF,
  ECA_CHERRY_STUDIO_HREF,
  ECA_DEMO_API_KEY,
  ECA_DOCS_HREF,
  ECA_NEW_API_HREF,
  ECA_REPO_HREF,
  resolveEcaAboutCopy,
  resolveEcaDocsHref,
} from '../copy'

const brandingDir = path.dirname(fileURLToPath(import.meta.url))
const demoSource = readFileSync(
  path.resolve(
    brandingDir,
    '../../../features/home/components/hero-terminal-demo.tsx'
  ),
  'utf8'
)
const heroSource = readFileSync(
  path.resolve(
    brandingDir,
    '../../../features/home/components/sections/hero.tsx'
  ),
  'utf8'
)
const footerSource = readFileSync(
  path.resolve(
    brandingDir,
    '../../../components/layout/components/footer.tsx'
  ),
  'utf8'
)
const homeApiSource = readFileSync(
  path.resolve(brandingDir, '../../../features/home/api.ts'),
  'utf8'
)
const statusApiSource = readFileSync(
  path.resolve(brandingDir, '../../../lib/api.ts'),
  'utf8'
)

describe('ECA landing copy', () => {
  afterEach(async () => {
    await i18n.changeLanguage('en')
  })

  test('overrides Chinese landing copy without editing locale JSON', async () => {
    applyEcaLandingCopy(i18n)
    await i18n.changeLanguage('zhCN')

    expect(i18n.t('AI Application Infrastructure Foundation')).toBe(
      'AI 应用基座'
    )
    expect(i18n.t('Vast Range of AI Models')).toBe('ECA Healthcare')
    expect(
      i18n.t(
        'Access a vast selection of models via a standard, unified API protocol. Power AI applications, manage digital assets, and connect the Future.'
      )
    ).toBe(
      '通过统一、标准的接口协议接入模型。承载 AI 应用，管理数字资产。'
    )
    expect(i18n.t(APP_SUPPORT_I18N_KEY)).toBe(
      '支持一键配置并适配 <repo>ECA API</repo> 多协议配置'
    )
    expect(i18n.t('with')).toBe('与')
    expect(i18n.t('Powerful API Management Platform')).toBe('')
    expect(i18n.t('View Pricing')).toBe('查看模型')
  })

  test('links ECA API in the apps line to the company fork', () => {
    expect(ECA_REPO_HREF).toBe('https://github.com/Ranshen1209/eca-new-api')
    expect(heroSource).toContain('Trans')
    expect(heroSource).toContain('ECA_REPO_HREF')
    expect(heroSource).toContain("repo:")
  })

  test('fills the About page with company-only service copy when admin content is empty', () => {
    expect(resolveEcaAboutCopy('zhCN')).toEqual(ECA_ABOUT_COPY.zhCN)
    expect(resolveEcaAboutCopy('zh-CN')).toEqual(ECA_ABOUT_COPY.zhCN)
    expect(resolveEcaAboutCopy('en')).toEqual(ECA_ABOUT_COPY.en)
    expect(resolveEcaAboutCopy('zhTW')).toEqual(ECA_ABOUT_COPY.zhTW)
    expect(resolveEcaAboutCopy('fr')).toEqual(ECA_ABOUT_COPY.fr)
    expect(ECA_ABOUT_COPY.zhCN.audience).toContain('ECA Healthcare')
    expect(ECA_ABOUT_COPY.zhCN.support).toContain('公司信息化专员')
    expect(ECA_ABOUT_COPY.zhCN.creditAfter).toContain('二次开发')
    expect(ECA_NEW_API_HREF).toBe('https://github.com/QuantumNous/new-api')

    const aboutSource = readFileSync(
      path.resolve(brandingDir, '../../../features/about/index.tsx'),
      'utf8'
    )
    expect(aboutSource).toContain('EcaAboutPage')
    expect(aboutSource).toContain('language={i18n.language}')
  })

  test('names CC Switch imports ECA API Claude, Codex, and Gemini instead of My Claude', () => {
    expect(ECA_CC_SWITCH_DEFAULT_NAMES).toEqual({
      claude: 'ECA API Claude',
      codex: 'ECA API Codex',
      gemini: 'ECA API Gemini',
    })

    const dialogSource = readFileSync(
      path.resolve(
        brandingDir,
        '../../../features/keys/components/dialogs/cc-switch-dialog.tsx'
      ),
      'utf8'
    )
    expect(dialogSource).toContain('ECA_CC_SWITCH_DEFAULT_NAMES')
    expect(dialogSource).not.toContain('My Claude')
    expect(dialogSource).not.toContain('My Codex')
    expect(dialogSource).not.toContain('My Gemini')
  })

  test('points Cherry Studio and CC Switch buttons at company OpenList folders', () => {
    expect(ECA_CHERRY_STUDIO_HREF).toBe(
      'https://openlist.eca-healthcare.org/Cherry%20Studio'
    )
    expect(ECA_CC_SWITCH_HREF).toBe(
      'https://openlist.eca-healthcare.org/CC%20Switch'
    )
    expect(heroSource).toContain('ECA_CHERRY_STUDIO_HREF')
    expect(heroSource).toContain('ECA_CC_SWITCH_HREF')
    expect(heroSource).not.toContain("href='https://cherry-ai.com'")
    expect(heroSource).not.toContain("href='https://ccswitch.io'")
  })

  test('points Docs links at the company docs site instead of upstream New API docs', () => {
    expect(ECA_DOCS_HREF).toBe('https://docs.eca-healthcare.org/')
    expect(resolveEcaDocsHref(undefined)).toBe(ECA_DOCS_HREF)
    expect(resolveEcaDocsHref('')).toBe(ECA_DOCS_HREF)
    expect(resolveEcaDocsHref('/docs')).toBe(ECA_DOCS_HREF)
    expect(resolveEcaDocsHref('https://docs.newapi.pro')).toBe(ECA_DOCS_HREF)
    expect(resolveEcaDocsHref('https://docs.newapi.pro/api/')).toBe(ECA_DOCS_HREF)
    expect(resolveEcaDocsHref('https://example.com/custom-docs')).toBe(
      'https://example.com/custom-docs'
    )
    expect(heroSource).toContain('resolveEcaDocsHref')
    expect(footerSource).toContain('ECA_DOCS_HREF')
    expect(footerSource).not.toContain('docs.newapi.pro')

    const navSource = readFileSync(
      path.resolve(brandingDir, '../../../hooks/use-top-nav-links.ts'),
      'utf8'
    )
    expect(navSource).toContain('resolveEcaDocsHref(docsLink)')
  })

  test('links ECA API in the footer credit to the company fork', () => {
    expect(footerSource).toContain('resolveEcaCompactMarkHref')
    expect(footerSource).toContain('ECA_REPO_HREF')
    expect(footerSource).toContain('DEFAULT_SYSTEM_NAME')
    expect(footerSource).toContain("t('New API')")
    expect(footerSource).toContain('https://github.com/QuantumNous/new-api')
    expect(footerSource).not.toContain("{displayName}.{' '}")
    expect(footerSource).not.toMatch(/<\/a>\s*\.\s*\{t\('with'\)/)
    expect(footerSource).not.toMatch(/<\/a>\s*\.\s*\{t\('footer\.defaultCopyright'\)/)
  })

  test('uses the ECA-API-KEY env var in the demo request instead of a fake sk- key', () => {
    expect(ECA_DEMO_API_KEY).toBe('$ECA-API-KEY')
    expect(demoSource).toContain(ECA_DEMO_API_KEY)
    expect(demoSource).not.toMatch(/sk-•/)
    expect(demoSource).not.toContain('sk-eca-demo')
  })

  test('does not toast when optional homepage bootstrap APIs time out', () => {
    expect(homeApiSource).toMatch(
      /\/api\/home_page_content[\s\S]*skipErrorHandler:\s*true/
    )
    expect(statusApiSource).toMatch(/\/api\/status[\s\S]*skipErrorHandler:\s*true/)
  })
})
