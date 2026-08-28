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
import { GitBranch, Headset } from 'lucide-react'

import { ECA_NEW_API_HREF, resolveEcaAboutCopy } from './copy'
import { ECA_MARK_HREF } from './favicon'
import { EcaOpticalFlow } from './optical-flow'

interface EcaAboutPageProps {
  language?: string | null
}

export function EcaAboutPage(props: EcaAboutPageProps) {
  const copy = resolveEcaAboutCopy(props.language)

  return (
    <div
      data-brand='eca'
      data-eca-about='page'
      className='relative flex min-h-svh flex-col'
    >
      <EcaOpticalFlow />
      <div
        data-eca-about='stage'
        className='relative z-10 flex min-h-svh flex-1 flex-col items-center justify-center px-4 pt-24 pb-10 sm:px-8 sm:pb-16 lg:px-12'
      >
        <article className='border-border/60 bg-card/90 w-full max-w-5xl overflow-hidden rounded-3xl border shadow-sm'>
          <header className='flex items-start gap-5 bg-gradient-to-br from-primary/12 via-background to-background px-6 py-8 sm:gap-6 sm:px-10 sm:py-10 lg:px-12 lg:py-12'>
            <img
              src={ECA_MARK_HREF}
              alt=''
              className='size-14 shrink-0 rounded-2xl object-contain sm:size-16'
            />
            <p className='text-muted-foreground min-w-0 flex-1 text-lg leading-relaxed text-pretty sm:text-xl lg:text-2xl'>
              {copy.audience}
            </p>
          </header>

          <div
            data-eca-about='panels'
            className='grid grid-cols-1 gap-5 border-t px-6 py-6 sm:px-10 sm:py-8 md:grid-cols-2 lg:gap-6 lg:px-12 lg:py-10'
          >
            <section className='bg-muted/35 min-h-36 min-w-0 rounded-2xl p-5 sm:min-h-40 sm:p-6 lg:p-7'>
              <div className='mb-3 flex items-center gap-2'>
                <Headset
                  aria-hidden='true'
                  className='text-primary size-4 shrink-0'
                />
                <h2 className='text-sm font-semibold sm:text-base'>
                  {copy.supportTitle}
                </h2>
              </div>
              <p className='text-muted-foreground text-sm leading-relaxed sm:text-base'>
                {copy.support}
              </p>
            </section>

            <section className='bg-muted/35 min-h-36 min-w-0 rounded-2xl p-5 sm:min-h-40 sm:p-6 lg:p-7'>
              <div className='mb-3 flex items-center gap-2'>
                <GitBranch
                  aria-hidden='true'
                  className='text-primary size-4 shrink-0'
                />
                <h2 className='text-sm font-semibold sm:text-base'>
                  {copy.creditTitle}
                </h2>
              </div>
              <p className='text-muted-foreground text-sm leading-relaxed sm:text-base'>
                {copy.creditBefore}
                <a
                  href={ECA_NEW_API_HREF}
                  target='_blank'
                  rel='noopener noreferrer'
                  className='text-primary font-medium underline-offset-2 hover:underline'
                >
                  New API
                </a>
                {copy.creditAfter}
              </p>
            </section>
          </div>
        </article>
      </div>
    </div>
  )
}
