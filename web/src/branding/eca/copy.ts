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
export const ECA_CC_SWITCH_DEFAULT_NAMES = {
  claude: 'ECA API Claude',
  codex: 'ECA API Codex',
  gemini: 'ECA API Gemini',
} as const

export const ECA_NEW_API_HREF = 'https://github.com/QuantumNous/new-api'

export type EcaAboutCopy = {
  audience: string
  supportTitle: string
  support: string
  creditTitle: string
  creditBefore: string
  creditAfter: string
}

export const ECA_ABOUT_COPY = {
  en: {
    audience:
      'This site provides API gateway services exclusively for ECA Healthcare. It is not a public service.',
    supportTitle: 'Access and support',
    support:
      'For access or technical support, contact the company IT specialist.',
    creditTitle: 'Upstream project',
    creditBefore: 'This site is a secondary development based on ',
    creditAfter: '.',
  },
  zhCN: {
    audience: '此站点仅为 ECA Healthcare 提供服务，不对外提供公开访问。',
    supportTitle: '开通与支持',
    support: '如需开通或技术支持，请联系公司信息化专员。',
    creditTitle: '项目来源',
    creditBefore: '本站点基于 ',
    creditAfter: ' 二次开发。',
  },
  zhTW: {
    audience: '此站點僅為 ECA Healthcare 提供服務，不對外提供公開存取。',
    supportTitle: '開通與支援',
    support: '如需開通或技術支援，請聯繫公司資訊化專員。',
    creditTitle: '專案來源',
    creditBefore: '本站點基於 ',
    creditAfter: ' 二次開發。',
  },
  fr: {
    audience:
      "Ce site fournit des services de passerelle API exclusivement pour ECA Healthcare. Il ne s'agit pas d'un service public.",
    supportTitle: 'Accès et support',
    support:
      "Pour l'accès ou le support technique, contactez le spécialiste informatique de l'entreprise.",
    creditTitle: 'Projet amont',
    creditBefore: 'Ce site est un développement secondaire basé sur ',
    creditAfter: '.',
  },
  ja: {
    audience:
      'このサイトは ECA Healthcare 専用の API ゲートウェイです。一般向けの公開サービスではありません。',
    supportTitle: '利用とサポート',
    support:
      '利用や技術サポートについては、社内の情報化担当者にお問い合わせください。',
    creditTitle: 'アップストリーム',
    creditBefore: '本サイトは',
    creditAfter: 'を二次開発したものです。',
  },
  ru: {
    audience:
      'Этот сайт предоставляет услуги API-шлюза исключительно для ECA Healthcare. Это не публичный сервис.',
    supportTitle: 'Доступ и поддержка',
    support:
      'По вопросам доступа и технической поддержки обратитесь к специалисту по информатизации компании.',
    creditTitle: 'Исходный проект',
    creditBefore: 'Этот сайт является вторичной разработкой на основе ',
    creditAfter: '.',
  },
  vi: {
    audience:
      'Trang này chỉ cung cấp dịch vụ API gateway cho ECA Healthcare. Đây không phải dịch vụ công khai.',
    supportTitle: 'Cấp quyền và hỗ trợ',
    support:
      'Để được cấp quyền hoặc hỗ trợ kỹ thuật, hãy liên hệ chuyên viên tin học hóa của công ty.',
    creditTitle: 'Dự án gốc',
    creditBefore: 'Trang này được phát triển lại dựa trên ',
    creditAfter: '.',
  },
} as const satisfies Record<string, EcaAboutCopy>

export function resolveEcaAboutCopy(language?: string | null): EcaAboutCopy {
  const normalized = (language ?? '').replaceAll('-', '')
  if (normalized === 'zhTW') {
    return ECA_ABOUT_COPY.zhTW
  }
  if (normalized === 'zhCN' || normalized === 'zh') {
    return ECA_ABOUT_COPY.zhCN
  }
  if (normalized === 'fr') {
    return ECA_ABOUT_COPY.fr
  }
  if (normalized === 'ja') {
    return ECA_ABOUT_COPY.ja
  }
  if (normalized === 'ru') {
    return ECA_ABOUT_COPY.ru
  }
  if (normalized === 'vi') {
    return ECA_ABOUT_COPY.vi
  }
  return ECA_ABOUT_COPY.en
}

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
const VIEW_PRICING_KEY = 'View Pricing'

export const ECA_LANDING_COPY = {
  en: {
    [BADGE_KEY]: 'AI Application Foundation',
    [HEADLINE_ACCENT_KEY]: 'ECA Healthcare',
    [SUBTITLE_KEY]:
      'Access models via a standard, unified API protocol. Power AI applications, manage digital assets.',
    [APP_SUPPORT_I18N_KEY]: APP_SUPPORT_I18N_KEY,
    [FOOTER_TAGLINE_KEY]: '',
    [FOOTER_WITH_KEY]: 'with',
    [VIEW_PRICING_KEY]: 'View Models',
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
    [VIEW_PRICING_KEY]: '查看模型',
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
    [VIEW_PRICING_KEY]: '查看模型',
  },
  fr: {
    [BADGE_KEY]: "Socle d'applications IA",
    [HEADLINE_ACCENT_KEY]: 'ECA Healthcare',
    [SUBTITLE_KEY]:
      "Accédez aux modèles via un protocole API standard et unifié. Propulsez les applications d'IA, gérez les actifs numériques.",
    [APP_SUPPORT_I18N_KEY]:
      "Prend en charge la configuration en un clic et s'adapte à la configuration multi-protocole de <repo>ECA API</repo>.",
    [FOOTER_TAGLINE_KEY]: '',
    [FOOTER_WITH_KEY]: 'avec',
    [VIEW_PRICING_KEY]: 'Voir les modèles',
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
    [VIEW_PRICING_KEY]: 'モデルを見る',
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
    [VIEW_PRICING_KEY]: 'Смотреть модели',
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
    [VIEW_PRICING_KEY]: 'Xem mô hình',
  },
} as const

export function applyEcaLandingCopy(instance: I18nInstance): void {
  for (const [lng, resources] of Object.entries(ECA_LANDING_COPY)) {
    instance.addResourceBundle(lng, 'translation', resources, true, true)
  }
}
