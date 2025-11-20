import React from 'react'
import { VergeXHeader } from '../../components/vergex/VergeXHeader'
import { useLanguage } from '../../contexts/LanguageContext'
import { t } from '../../i18n/translations'
import SplitText from '../../../@/components/SplitText'

export const HeroSection: React.FC = () => {
  const { language } = useLanguage()

  return (
    <section className="relative h-[960px] w-full overflow-hidden">
      {/* Background Gradient Line */}
      <div className="absolute bottom-0 left-0 right-0 h-[1px] bg-gradient-to-r from-transparent via-white/20 to-transparent" />

      {/* Hero Background Image */}
      <div className="absolute top-0 left-0 w-full h-[810px] overflow-hidden">
        <div className="w-[1440px] mx-auto h-full">
          <img
            src="/vergex/hero-bg.png"
            alt=""
            className="w-full h-full object-cover object-[50%_50%]"
          />
        </div>
      </div>

      {/* Purple Blur - Left */}
      <div className="w-[1440px] mx-auto absolute h-full">
        <div
          className="absolute left-[-19px] top-[110px] w-[422px] h-[422px] rounded-full overflow-hidden"
          style={{ filter: 'blur(200px)' }}
        >
          <img
            src="/vergex/blur-left.png"
            alt=""
            className="w-[277.94%] h-[208.13%] object-cover"
            style={{ transform: 'translate(-21.97%, -53.66%)' }}
          />
        </div>

        {/* Purple Blur - Right */}
        <div
          className="absolute left-[911px] top-[206px] w-[676px] h-[676px] rounded-full overflow-hidden"
          style={{ filter: 'blur(300px)' }}
        >
          <img
            src="/vergex/blur-right.png"
            alt=""
            className="w-[277.94%] h-[208.13%] object-cover"
            style={{ transform: 'translate(-21.97%, -53.66%)' }}
          />
        </div>

        {/* Decorative Lines */}
        <svg
          className="absolute left-[140px] top-[286px]"
          width="1"
          height="238"
        >
          <line
            x1="0"
            y1="0"
            x2="0"
            y2="238"
            stroke="white"
            strokeOpacity="0.2"
            strokeWidth="1"
          />
        </svg>
        <svg
          className="absolute left-[1088px] top-[286px]"
          width="1"
          height="108"
        >
          <line
            x1="0"
            y1="0"
            x2="0"
            y2="108"
            stroke="white"
            strokeOpacity="0.2"
            strokeWidth="1"
          />
        </svg>
        <svg
          className="absolute left-[307px] top-[167px]"
          width="1"
          height="307"
        >
          <line
            x1="0"
            y1="0"
            x2="0"
            y2="307"
            stroke="white"
            strokeOpacity="0.2"
            strokeWidth="1"
          />
        </svg>
        <svg
          className="absolute left-[1236px] top-[384px]"
          width="1"
          height="183"
        >
          <line
            x1="0"
            y1="0"
            x2="0"
            y2="183"
            stroke="white"
            strokeOpacity="0.2"
            strokeWidth="1"
          />
        </svg>
        <svg
          className="absolute left-[1363px] top-[137px]"
          width="1"
          height="339"
        >
          <line
            x1="0"
            y1="0"
            x2="0"
            y2="339"
            stroke="white"
            strokeOpacity="0.2"
            strokeWidth="1"
          />
        </svg>
      </div>

      {/* Header */}
      <div className="absolute top-0 left-0 right-0 z-50">
        <VergeXHeader />
      </div>

      {/* Hero Content */}
      <div className="w-[1440px] mx-auto relative h-full">
        <div className="absolute left-[64px] top-[650px] w-[976px] flex flex-col gap-[8px] items-start">
          <div className="w-full [&_.split-word:nth-child(2)]:text-vergex-primary">
            <SplitText
              text={
                language === 'en'
                  ? 'The Agentic AI Trading Platform'
                  : t('vergex.heroTitle', language)
              }
              tag="h1"
              className="font-vergex-heading font-medium text-[64px] leading-[80px] text-vergex-text-primary"
              splitType="words"
              delay={50}
              duration={0.8}
              from={{ opacity: 0, y: 60 }}
              to={{ opacity: 1, y: 0 }}
              textAlign="left"
            />
          </div>
          <div className="w-full">
            <SplitText
              text={t('vergex.heroDescription', language)}
              tag="p"
              className="font-vergex-body font-normal text-[20px] leading-[36px] text-vergex-text-primary"
              splitType="words"
              delay={30}
              duration={0.6}
              from={{ opacity: 0, y: 40 }}
              to={{ opacity: 1, y: 0 }}
              textAlign="left"
            />
          </div>
        </div>

        {/* Open App Button */}
        <button className="absolute left-[1221px] top-[722px] bg-vergex-primary rounded-[32px] px-[32px] py-[12px] flex items-center gap-[8px] hover:bg-vergex-primary-light transition-colors">
          <span className="font-vergex-body font-medium text-[18px] leading-[28px] text-black">
            {t('vergex.openApp', language)}
          </span>
          <svg
            width="24"
            height="24"
            viewBox="0 0 24 24"
            className="rotate-90 scale-y-[-1]"
          >
            <path
              d="M12 5L12 19M12 5L7 10M12 5L17 10"
              stroke="black"
              strokeWidth="2"
              strokeLinecap="round"
              strokeLinejoin="round"
              fill="none"
            />
          </svg>
        </button>
      </div>
    </section>
  )
}
