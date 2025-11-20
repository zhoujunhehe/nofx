import React from 'react'
import { useLanguage } from '../../contexts/LanguageContext'
import { t } from '../../i18n/translations'

export const OurProductsSection: React.FC = () => {
  const { language } = useLanguage()

  return (
    <section className="w-full relative">
      <div className="w-[1440px] mx-auto px-[64px] py-[100px] relative">
        {/* Background Blur */}
        <div
          className="absolute left-1/2 top-[302px] -translate-x-1/2 w-[676px] h-[676px] rounded-full overflow-hidden"
          style={{ filter: 'blur(300px)' }}
        >
          <img
            src="/vergex/blur-center.png"
            alt=""
            className="w-[277.94%] h-[208.13%] object-cover"
            style={{ transform: 'translate(-21.97%, -53.66%)' }}
          />
        </div>

        <h2 className="font-vergex-body font-semibold text-[48px] leading-[72px] text-vergex-text-primary mb-[64px] relative z-10">
          {t('vergex.ourProducts', language)}
        </h2>

        <div className="flex items-center justify-between relative z-10">
          {/* Architecture Diagram */}
          <div className="w-[588.897px] h-[620px] relative">
            <div className="w-full h-full bg-gradient-to-br from-white/[0.05] to-transparent border border-vergex-border-light rounded-[12px] flex items-center justify-center">
              <svg width="589" height="620" viewBox="0 0 589 620" fill="none">
                {/* User Layer */}
                <g>
                  <path
                    d="M100 170 L489 170 L544 254 L489 338 L100 338 L45 254 Z"
                    fill="rgba(153, 140, 255, 0.1)"
                    stroke="rgba(153, 140, 255, 0.3)"
                    strokeWidth="1.5"
                  />
                  <text
                    x="294"
                    y="264"
                    fill="white"
                    fontSize="20"
                    textAnchor="middle"
                    fontFamily="Red Hat Text"
                    letterSpacing="0.8"
                  >
                    USER
                  </text>
                </g>

                {/* VergeX Layer */}
                <g>
                  <path
                    d="M100 310 L489 310 L544 394 L489 478 L100 478 L45 394 Z"
                    fill="rgba(153, 140, 255, 0.15)"
                    stroke="rgba(153, 140, 255, 0.4)"
                    strokeWidth="1.5"
                  />
                  <text
                    x="294"
                    y="404"
                    fill="white"
                    fontSize="20"
                    textAnchor="middle"
                    fontFamily="Red Hat Text"
                    letterSpacing="0.8"
                  >
                    VERGEX
                  </text>
                </g>

                {/* Exchange Layer */}
                <g>
                  <path
                    d="M100 450 L489 450 L544 534 L489 618 L100 618 L45 534 Z"
                    fill="rgba(153, 140, 255, 0.1)"
                    stroke="rgba(153, 140, 255, 0.3)"
                    strokeWidth="1.5"
                  />
                  <text
                    x="294"
                    y="544"
                    fill="white"
                    fontSize="20"
                    textAnchor="middle"
                    fontFamily="Red Hat Text"
                    letterSpacing="0.8"
                  >
                    EXCHANGE
                  </text>
                </g>

                {/* AI Models Layer - rotated text */}
                <text
                  x="102.99"
                  y="539.25"
                  fill="white"
                  fontSize="20"
                  fontFamily="Red Hat Text"
                  letterSpacing="0.8"
                  transform="rotate(30 102.99 539.25)"
                >
                  AI MODELS
                </text>
              </svg>
            </div>
          </div>

          {/* Product Descriptions */}
          <div className="w-[580px] flex flex-col gap-[80px]">
            {/* User Layer */}
            <div className="flex flex-col gap-[12px]">
              <div className="flex items-center gap-[12px]">
                <svg width="24" height="24" viewBox="0 0 24 24" fill="none">
                  <circle
                    cx="12"
                    cy="8"
                    r="4"
                    stroke="white"
                    strokeWidth="1.5"
                  />
                  <path
                    d="M6 21C6 17.686 8.686 15 12 15C15.314 15 18 17.686 18 21"
                    stroke="white"
                    strokeWidth="1.5"
                    strokeLinecap="round"
                  />
                </svg>
                <h3 className="font-vergex-body font-semibold text-[24px] leading-[36px] tracking-[2.4px] uppercase text-vergex-text-primary">
                  {t('vergex.userLayer', language)}
                </h3>
              </div>
              <p className="font-vergex-body font-light text-[20px] leading-[30px] text-vergex-text-primary/50">
                {t('vergex.userLayerDesc', language)}
              </p>
            </div>

            {/* VergeX Agentic */}
            <div className="flex flex-col gap-[12px]">
              <div className="flex items-center gap-[12px]">
                <div className="w-[24px] h-[24px] bg-gradient-to-br from-[#998cff] to-[#7d6ee5] rounded flex items-center justify-center rotate-90">
                  <div
                    className="-rotate-90 w-0 h-0"
                    style={{
                      borderLeft: '7px solid transparent',
                      borderRight: '7px solid transparent',
                      borderBottom: '12px solid white',
                    }}
                  />
                </div>
                <h3 className="font-vergex-body font-semibold text-[24px] leading-[36px] tracking-[2.4px] uppercase text-vergex-text-primary">
                  {t('vergex.vergexAgentic', language)}
                </h3>
              </div>
              <p className="font-vergex-body font-light text-[20px] leading-[30px] text-vergex-text-primary/50">
                {t('vergex.vergexAgenticDesc', language)}
              </p>
            </div>

            {/* Trading Venues */}
            <div className="flex flex-col gap-[12px]">
              <div className="flex items-center gap-[12px]">
                <svg width="24" height="24" viewBox="0 0 24 24" fill="none">
                  <path
                    d="M3 21H21M5 21V7L12 3L19 7V21M9 9H10M14 9H15M9 13H10M14 13H15M9 17H10M14 17H15"
                    stroke="white"
                    strokeWidth="1.5"
                    strokeLinecap="round"
                    strokeLinejoin="round"
                  />
                </svg>
                <h3 className="font-vergex-body font-semibold text-[24px] leading-[36px] tracking-[2.4px] uppercase text-vergex-text-primary">
                  {t('vergex.tradingVenues', language)}
                </h3>
              </div>
              <p className="font-vergex-body font-light text-[20px] leading-[30px] text-vergex-text-primary/50">
                {t('vergex.tradingVenuesDesc', language)}
              </p>
            </div>

            {/* AI Models */}
            <div className="flex flex-col gap-[12px]">
              <div className="flex items-center gap-[12px]">
                <svg width="24" height="24" viewBox="0 0 24 24" fill="none">
                  <circle
                    cx="12"
                    cy="12"
                    r="3"
                    stroke="white"
                    strokeWidth="1.5"
                  />
                  <path
                    d="M12 3V5M12 19V21M21 12H19M5 12H3M18.364 5.636L16.95 7.05M7.05 16.95L5.636 18.364M18.364 18.364L16.95 16.95M7.05 7.05L5.636 5.636"
                    stroke="white"
                    strokeWidth="1.5"
                    strokeLinecap="round"
                  />
                </svg>
                <h3 className="font-vergex-body font-semibold text-[24px] leading-[36px] tracking-[2.4px] uppercase text-vergex-text-primary">
                  {t('vergex.aiModels', language)}
                </h3>
              </div>
              <p className="font-vergex-body font-light text-[20px] leading-[30px] text-vergex-text-primary/50">
                {t('vergex.aiModelsDesc', language)}
              </p>
            </div>
          </div>
        </div>
      </div>
    </section>
  )
}
