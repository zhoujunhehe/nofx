import React from 'react'
import {
  SecurityIcon,
  NaturalLanguageIcon,
  IntelligentBacktestIcon,
  AlwaysOnIcon,
} from '../../components/vergex/svg-assets'
import { useLanguage } from '../../contexts/LanguageContext'
import { t } from '../../i18n/translations'

export const WhyVergeXSection: React.FC = () => {
  const { language } = useLanguage()

  return (
    <section className="w-full">
      <div className="w-[1440px] mx-auto px-[64px] py-[100px]">
        <h2 className="font-vergex-body font-semibold text-[48px] leading-[72px] text-vergex-text-primary mb-[64px]">
          {t('vergex.whyVergeX', language)}
        </h2>

        <div className="flex gap-[32px] items-start">
          {/* Large Card - Security First */}
          <div className="w-[527px] border border-vergex-border-light rounded-[12px] p-[32px] relative overflow-hidden self-stretch">
            {/* Purple Blur Background */}
            <div
              className="absolute left-[200px] top-[190px] w-[340px] h-[340px] rounded-[500px] bg-[rgba(153,140,255,0.4)] pointer-events-none"
              style={{ filter: 'blur(200px)' }}
            />

            {/* Background Decorative SVG - positioned at bottom right */}
            <div className="absolute bottom-[61.73px] right-[-433.1px] w-[877.097px] h-[735.273px] opacity-10 pointer-events-none">
              <svg width="877" height="736" viewBox="0 0 877 736" fill="none">
                <g opacity="0.1">
                  <circle
                    cx="438"
                    cy="367"
                    r="300"
                    stroke="currentColor"
                    strokeWidth="1"
                    className="text-vergex-primary"
                  />
                  <circle
                    cx="438"
                    cy="367"
                    r="250"
                    stroke="currentColor"
                    strokeWidth="1"
                    className="text-vergex-primary"
                  />
                  <circle
                    cx="438"
                    cy="367"
                    r="200"
                    stroke="currentColor"
                    strokeWidth="1"
                    className="text-vergex-primary"
                  />
                </g>
              </svg>
            </div>

            {/* Content */}
            <div className="relative z-10 flex flex-col h-full justify-between">
              {/* Icon */}
              <div className="shrink-0">
                <SecurityIcon />
              </div>

              {/* Text Content */}
              <div className="flex flex-col gap-[4px] w-full">
                <h3 className="font-vergex-body font-medium text-[36px] leading-[54px] text-vergex-text-primary">
                  {t('vergex.securityFirst', language)}
                </h3>
                <p className="font-vergex-body font-light text-[20px] leading-[30px] text-vergex-text-secondary">
                  {t('vergex.securityFirstDesc', language)}
                </p>
              </div>
            </div>
          </div>

          {/* Right Column - Two rows */}
          <div className="flex-1 flex flex-col gap-[32px] self-stretch">
            {/* Top Row - Natural Language & Intelligent Backtest */}
            <div className="flex gap-[32px]">
              {/* Natural Language Card */}
              <div className="flex-1 border border-vergex-border-light rounded-[12px] p-[24px] flex flex-col gap-[16px]">
                <div className="shrink-0">
                  <NaturalLanguageIcon />
                </div>
                <div className="flex flex-col gap-[4px] w-full">
                  <h3 className="font-vergex-body font-medium text-[24px] leading-[36px] text-vergex-text-primary">
                    {t('vergex.naturalLanguage', language)}
                  </h3>
                  <p className="font-vergex-body font-light text-[18px] leading-[27px] text-vergex-text-secondary">
                    {t('vergex.naturalLanguageDesc', language)}
                  </p>
                </div>
              </div>

              {/* Intelligent Backtest Card */}
              <div className="flex-1 border border-vergex-border-light rounded-[12px] p-[24px] flex flex-col gap-[16px]">
                <div className="shrink-0">
                  <IntelligentBacktestIcon />
                </div>
                <div className="flex flex-col gap-[4px] w-full">
                  <h3 className="font-vergex-body font-medium text-[24px] leading-[36px] text-vergex-text-primary">
                    {t('vergex.intelligentBacktest', language)}
                  </h3>
                  <p className="font-vergex-body font-light text-[18px] leading-[27px] text-vergex-text-secondary">
                    {t('vergex.intelligentBacktestDesc', language)}
                  </p>
                </div>
              </div>
            </div>

            {/* Bottom Row - Always-On Execution (Full Width) */}
            <div className="border border-vergex-border-light rounded-[12px] p-[24px] flex flex-col gap-[16px]">
              <div className="shrink-0">
                <AlwaysOnIcon />
              </div>
              <div className="flex flex-col gap-[4px] w-full">
                <h3 className="font-vergex-body font-medium text-[24px] leading-[36px] text-vergex-text-primary">
                  {t('vergex.alwaysOnExecution', language)}
                </h3>
                <p className="font-vergex-body font-light text-[18px] leading-[27px] text-vergex-text-secondary">
                  {t('vergex.alwaysOnExecutionDesc', language)}
                </p>
              </div>
            </div>
          </div>
        </div>
      </div>
    </section>
  )
}
