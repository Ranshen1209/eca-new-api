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
import type { i18n as I18nInstance } from 'i18next'

export const ECA_DEMO_API_KEY = '$ECA-API-KEY'
export const ECA_REPO_HREF = 'https://github.com/Ranshen1209/eca-new-api'
export const ECA_DOCS_HREF = 'https://docs.eca-healthcare.org/'
export const ECA_CHERRY_STUDIO_HREF =
  'https://openlist.eca-healthcare.org/Cherry%20Studio'
export const ECA_CC_SWITCH_HREF =
  'https://openlist.eca-healthcare.org/CC%20Switch'

const UPSTREAM_DOCS_HOST = 'docs.newapi.pro'

export function resolveEcaDocsHref(href?: string | null): string {
  const trimmed = href?.trim()
  if (!trimmed || trimmed === '/docs') {
    return ECA_DOCS_HREF
  }

  try {
    const url = new URL(trimmed, 'http://local.invalid')
    if (
      url.hostname === UPSTREAM_DOCS_HOST ||
      url.hostname === `www.${UPSTREAM_DOCS_HOST}`
    ) {
      return ECA_DOCS_HREF
    }
    if (url.hostname === 'local.invalid' && url.pathname === '/docs') {
      return ECA_DOCS_HREF
    }
  } catch {
    if (trimmed.includes(UPSTREAM_DOCS_HOST)) {
      return ECA_DOCS_HREF
    }
  }

  return trimmed
}

export const APP_SUPPORT_I18N_KEY =
  'Supports one-click configuration and adapts to <repo>ECA API</repo> multi-protocol configuration.'

const BADGE_KEY = 'AI Application Infrastructure Foundation'
const HEADLINE_ACCENT_KEY = 'Vast Range of AI Models'
const SUBTITLE_KEY =
  'Access a vast selection of models via a standard, unified API protocol. Power AI applications, manage digital assets, and connect the Future.'
const FOOTER_TAGLINE_KEY = 'Powerful API Management Platform'
const FOOTER_WITH_KEY = 'with'

export const ECA_LANDING_COPY = {
  en: {
    [BADGE_KEY]: 'AI Application Foundation',
    [HEADLINE_ACCENT_KEY]: 'ECA Healthcare',
    [SUBTITLE_KEY]:
      'Access models via a standard, unified API protocol. Power AI applications, manage digital assets.',
    [APP_SUPPORT_I18N_KEY]: APP_SUPPORT_I18N_KEY,
    [FOOTER_TAGLINE_KEY]: '',
    [FOOTER_WITH_KEY]: 'with',
  },
  zhCN: {
    [BADGE_KEY]: 'AI 应用基座',
    [HEADLINE_ACCENT_KEY]: 'ECA Healthcare',
    [SUBTITLE_KEY]:
      '通过统一、标准的接口协议接入模型。承载 AI 应用，管理数字资产。',
    [APP_SUPPORT_I18N_KEY]:
      '支持一键配置并适配 <repo>ECA API</repo> 多协议配置',
    [FOOTER_TAGLINE_KEY]: '',
    [FOOTER_WITH_KEY]: '与',
  },
  zhTW: {
    [BADGE_KEY]: 'AI 應用基座',
    [HEADLINE_ACCENT_KEY]: 'ECA Healthcare',
    [SUBTITLE_KEY]:
      '透過統一、標準的介面協定接入模型。承載 AI 應用，管理數位資產。',
    [APP_SUPPORT_I18N_KEY]:
      '支援一鍵設定並適配 <repo>ECA API</repo> 多協定設定',
    [FOOTER_TAGLINE_KEY]: '',
    [FOOTER_WITH_KEY]: '與',
  },
  fr: {
    [BADGE_KEY]: "Socle d'applications IA",
    [HEADLINE_ACCENT_KEY]: 'ECA Healthcare',
    [SUBTITLE_KEY]:
      "Accédez aux modèles via un protocole API standard et unifié. Propulsez les applications d'IA, gérez les actifs numériques.",
    [APP_SUPPORT_I18N_KEY]:
      'Prend en charge la configuration en un clic et s\'adapte à la configuration multi-protocole de <repo>ECA API</repo>.',
    [FOOTER_TAGLINE_KEY]: '',
    [FOOTER_WITH_KEY]: 'avec',
  },
  ja: {
    [BADGE_KEY]: 'AI アプリケーション基盤',
    [HEADLINE_ACCENT_KEY]: 'ECA Healthcare',
    [SUBTITLE_KEY]:
      '標準的で統一された API プロトコルを介してモデルにアクセス。AI アプリケーションを支え、デジタル資産を管理します。',
    [APP_SUPPORT_I18N_KEY]:
      'ワンクリック設定をサポートし、<repo>ECA API</repo> マルチプロトコル設定に適応します。',
    [FOOTER_TAGLINE_KEY]: '',
    [FOOTER_WITH_KEY]: 'および',
  },
  ru: {
    [BADGE_KEY]: 'Основание ИИ-приложений',
    [HEADLINE_ACCENT_KEY]: 'ECA Healthcare',
    [SUBTITLE_KEY]:
      'Получите доступ к моделям через стандартный единый протокол API. Развивайте приложения ИИ, управляйте цифровыми активами.',
    [APP_SUPPORT_I18N_KEY]:
      'Поддерживает настройку в один клик и адаптируется к многопротокольной конфигурации <repo>ECA API</repo>.',
    [FOOTER_TAGLINE_KEY]: '',
    [FOOTER_WITH_KEY]: 'и',
  },
  vi: {
    [BADGE_KEY]: 'Nền tảng ứng dụng AI',
    [HEADLINE_ACCENT_KEY]: 'ECA Healthcare',
    [SUBTITLE_KEY]:
      'Truy cập các mô hình thông qua giao thức API chuẩn hóa và thống nhất. Vận hành ứng dụng AI, quản lý tài sản số.',
    [APP_SUPPORT_I18N_KEY]:
      'Hỗ trợ cấu hình một cú nhấp và thích ứng với cấu hình đa giao thức <repo>ECA API</repo>.',
    [FOOTER_TAGLINE_KEY]: '',
    [FOOTER_WITH_KEY]: 'và',
  },
} as const

export function applyEcaLandingCopy(instance: I18nInstance): void {
  for (const [lng, resources] of Object.entries(ECA_LANDING_COPY)) {
    instance.addResourceBundle(lng, 'translation', resources, true, true)
  }
}
