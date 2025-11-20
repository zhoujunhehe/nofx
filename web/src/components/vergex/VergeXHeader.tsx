import React from 'react'
import { useLanguage } from '../../contexts/LanguageContext'
import { t } from '../../i18n/translations'

// SVG Icons matching Figma design
const GlobeIcon = () => (
  <svg width="24" height="24" viewBox="0 0 24 24" fill="none">
    <circle cx="12" cy="12" r="9" stroke="currentColor" strokeWidth="1.5" />
    <path
      d="M3 12h18M12 3a15.3 15.3 0 0 1 4 10 15.3 15.3 0 0 1-4 10M12 3a15.3 15.3 0 0 0-4 10 15.3 15.3 0 0 0 4 10"
      stroke="currentColor"
      strokeWidth="1.5"
    />
  </svg>
)

const ExternalLinkIcon = ({ className = '' }: { className?: string }) => (
  <svg
    width="20"
    height="20"
    viewBox="0 0 20 20"
    fill="none"
    className={className}
  >
    <path
      d="M5 15L15 5M15 5H8.33333M15 5V11.6667"
      stroke="currentColor"
      strokeWidth="1.5"
      strokeLinecap="round"
      strokeLinejoin="round"
    />
  </svg>
)

const CheckIcon = () => (
  <svg width="16" height="16" viewBox="0 0 16 16" fill="none">
    <path
      d="M13.3337 4L6.00033 11.3333L2.66699 8"
      stroke="currentColor"
      strokeWidth="1.5"
      strokeLinecap="round"
      strokeLinejoin="round"
    />
  </svg>
)

// VergeX Logo Component
const VergeXLogo = () => (
  <div className="flex items-center gap-[12px]">
    <img
      src="/vergex/logo.png"
      alt="VergeX logo"
      className="w-[191px] h-[40px]"
    />
  </div>
)

// Language Selector Component
const LanguageSelector: React.FC = () => {
  const { language, setLanguage } = useLanguage()
  const [isOpen, setIsOpen] = React.useState(false)
  const timeoutRef = React.useRef<NodeJS.Timeout | null>(null)

  const languages = [
    { label: 'English', value: 'en' as const },
    { label: '简体中文', value: 'zh' as const },
  ]

  const currentLanguageLabel = language === 'en' ? 'English' : '简体中文'

  const handleMouseEnter = () => {
    if (timeoutRef.current) {
      clearTimeout(timeoutRef.current)
      timeoutRef.current = null
    }
    setIsOpen(true)
  }

  const handleMouseLeave = () => {
    timeoutRef.current = setTimeout(() => {
      setIsOpen(false)
    }, 100)
  }

  React.useEffect(() => {
    return () => {
      if (timeoutRef.current) {
        clearTimeout(timeoutRef.current)
      }
    }
  }, [])

  return (
    <div
      className="relative"
      onMouseEnter={handleMouseEnter}
      onMouseLeave={handleMouseLeave}
    >
      <button className="flex items-center gap-[4px] bg-vergex-bg-secondary hover:bg-vergex-bg-tertiary rounded-[24px] pl-[6px] pr-[8px] py-[6px] transition-colors cursor-pointer border-none outline-none">
        <div className="text-vergex-text-primary">
          <GlobeIcon />
        </div>
        <span className="font-vergex-body font-medium text-[16px] leading-[24px] text-vergex-text-primary">
          {currentLanguageLabel}
        </span>
      </button>

      {isOpen && (
        <div
          className="absolute right-0 top-[calc(100%+16px)] w-[280px] backdrop-blur-vergex bg-gradient-to-b from-vergex-gradient-start to-vergex-gradient-end border border-vergex-border-light rounded-[8px] p-[8px] z-50 flex flex-col gap-[8px]"
          onMouseEnter={handleMouseEnter}
          onMouseLeave={handleMouseLeave}
        >
          <div className="flex items-center gap-[6px] p-[8px] rounded-[8px]">
            <div className="text-vergex-primary w-[20px] h-[20px] flex items-center justify-center">
              <svg width="20" height="20" viewBox="0 0 20 20" fill="none">
                <circle
                  cx="10"
                  cy="10"
                  r="7.5"
                  stroke="currentColor"
                  strokeWidth="1.5"
                />
                <path
                  d="M2.5 10h15M10 2.5a12.75 12.75 0 0 1 3.333 8.333A12.75 12.75 0 0 1 10 19.167M10 2.5a12.75 12.75 0 0 0-3.333 8.333A12.75 12.75 0 0 0 10 19.167"
                  stroke="currentColor"
                  strokeWidth="1.5"
                />
              </svg>
            </div>
            <span className="font-vergex-body font-medium text-[14px] leading-[21px] text-vergex-primary">
              {t('vergex.language', language)}
            </span>
          </div>

          <div className="w-full h-[1px] bg-vergex-border-light" />

          {languages.map((lang) => {
            const active = language === lang.value
            return (
              <button
                key={lang.value}
                className={`w-full flex items-center justify-between p-[8px] rounded-[6px] cursor-pointer transition-colors border-none outline-none text-left ${
                  active ? 'bg-vergex-bg-secondary' : 'hover:bg-vergex-bg-secondary'
                }`}
                onClick={() => setLanguage(lang.value)}
              >
                <span className="font-vergex-body font-medium text-[16px] leading-[24px] text-vergex-text-primary">
                  {lang.label}
                </span>
                {active && (
                  <div className="text-vergex-text-primary shrink-0">
                    <CheckIcon />
                  </div>
                )}
              </button>
            )
          })}
        </div>
      )}
    </div>
  )
}

// Navigation Item with Dropdown
interface NavItemProps {
  label: string
  href: string
  active?: boolean
  items?: Array<{ label: string; href: string }>
}

const NavItem: React.FC<NavItemProps> = ({ label, href, active, items }) => {
  const [isOpen, setIsOpen] = React.useState(false)
  const timeoutRef = React.useRef<NodeJS.Timeout | null>(null)

  const handleMouseEnter = () => {
    if (timeoutRef.current) {
      clearTimeout(timeoutRef.current)
      timeoutRef.current = null
    }
    setIsOpen(true)
  }

  const handleMouseLeave = () => {
    timeoutRef.current = setTimeout(() => {
      setIsOpen(false)
    }, 100)
  }

  React.useEffect(() => {
    return () => {
      if (timeoutRef.current) {
        clearTimeout(timeoutRef.current)
      }
    }
  }, [])

  if (!items) {
    return (
      <a
        href={href}
        className={`font-vergex-body font-normal text-[16px] leading-[24px] transition-colors ${
          active ? 'text-vergex-text-primary' : 'text-vergex-text-primary/60 hover:text-vergex-text-primary'
        }`}
      >
        {label}
      </a>
    )
  }

  // Calculate dropdown width based on menu type
  const getDropdownWidth = () => {
    if (label === 'Products') return 'w-[240px]'
    if (label === 'Resources' || label === 'Company') return 'w-[200px]'
    return 'w-[240px]'
  }

  return (
    <div
      className="relative"
      onMouseEnter={handleMouseEnter}
      onMouseLeave={handleMouseLeave}
    >
      <button
        className={`font-vergex-body font-normal text-[16px] leading-[24px] transition-colors cursor-pointer bg-transparent border-none outline-none ${
          isOpen ? 'text-vergex-text-primary' : 'text-vergex-text-primary/60 hover:text-vergex-text-primary'
        }`}
      >
        {label}
      </button>

      {isOpen && (
        <div
          className={`absolute left-0 top-[calc(100%+12px)] ${getDropdownWidth()} backdrop-blur-vergex bg-gradient-to-b from-vergex-gradient-start to-vergex-gradient-end border border-vergex-border-light rounded-[8px] p-[8px] z-50 flex flex-col gap-[8px]`}
          onMouseEnter={handleMouseEnter}
          onMouseLeave={handleMouseLeave}
        >
          <div className="flex gap-[6px] items-center p-[8px] rounded-[8px]">
            <p className="font-vergex-body font-medium text-[14px] leading-[21px] text-vergex-primary">
              {label}
            </p>
          </div>

          <div className="w-full h-[1px] bg-vergex-border-light" />

          {items.map((sub) => (
            <a
              key={sub.label}
              href={sub.href}
              className="flex items-center justify-between p-[8px] rounded-[6px] hover:bg-vergex-bg-secondary transition-colors group cursor-pointer no-underline"
            >
              <span className="font-vergex-body font-medium text-[16px] leading-[24px] text-vergex-text-primary">
                {sub.label}
              </span>
              <div className="text-vergex-text-primary/40 group-hover:text-vergex-text-primary transition-colors shrink-0">
                <ExternalLinkIcon />
              </div>
            </a>
          ))}
        </div>
      )}
    </div>
  )
}

// Main Header Component
export const VergeXHeader: React.FC = () => {
  const { language } = useLanguage()

  const menuConfig = {
    products: [
      { labelKey: 'vergex.aiCompetition', href: '/competition' },
      { labelKey: 'vergex.aiTrader', href: '/traders' },
      { labelKey: 'vergex.performanceDashboard', href: '/dashboard' },
    ],
    resources: [
      { labelKey: 'vergex.faq', href: '/faq' },
      { labelKey: 'vergex.docs', href: '/docs' },
      { labelKey: 'vergex.github', href: 'https://github.com' },
      { labelKey: 'vergex.security', href: '/security' },
    ],
    company: [
      { labelKey: 'vergex.about', href: '/about' },
      { labelKey: 'vergex.contact', href: '/contact' },
      { labelKey: 'vergex.privacyPolicy', href: '/privacy' },
      { labelKey: 'vergex.termsOfUse', href: '/terms' },
    ],
  }

  const navItems: (NavItemProps & { labelKey?: string })[] = [
    { labelKey: 'vergex.home', label: t('vergex.home', language), href: '/', active: true },
    {
      labelKey: 'vergex.products',
      label: t('vergex.products', language),
      href: '#',
      items: menuConfig.products.map((item) => ({
        label: t(item.labelKey, language),
        href: item.href,
      })),
    },
    {
      labelKey: 'vergex.resources',
      label: t('vergex.resources', language),
      href: '#',
      items: menuConfig.resources.map((item) => ({
        label: t(item.labelKey, language),
        href: item.href,
      })),
    },
    {
      labelKey: 'vergex.company',
      label: t('vergex.company', language),
      href: '#',
      items: menuConfig.company.map((item) => ({
        label: t(item.labelKey, language),
        href: item.href,
      })),
    },
  ]

  return (
    <header className="relative h-[80px] w-full bg-vergex-bg-secondary">
      <div className="w-[1440px] mx-auto h-full relative">
        {/* Logo */}
        <div className="absolute left-[64px] top-[20px]">
          <VergeXLogo />
        </div>

        {/* Navigation */}
        <nav className="absolute left-1/2 top-[28px] -translate-x-1/2 flex items-center gap-[48px]">
          {navItems.map((item) => (
            <NavItem key={item.label} {...item} />
          ))}
        </nav>

        {/* Language Selector */}
        <div className="absolute right-[64px] top-[22px]">
          <LanguageSelector />
        </div>
      </div>
    </header>
  )
}

export default VergeXHeader
