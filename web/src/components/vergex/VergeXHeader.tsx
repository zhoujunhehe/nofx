import React from 'react'

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
  const [currentLanguage, setCurrentLanguage] = React.useState('English')
  const [isOpen, setIsOpen] = React.useState(false)
  const timeoutRef = React.useRef<NodeJS.Timeout | null>(null)

  const languages = [
    { label: 'English', value: 'English' },
    { label: '简体中文', value: '简体中文' },
  ]

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
      <button className="flex items-center gap-[4px] bg-white/[0.04] hover:bg-white/[0.06] rounded-[24px] pl-[6px] pr-[8px] py-[6px] transition-colors cursor-pointer border-none outline-none">
        <div className="text-white">
          <GlobeIcon />
        </div>
        <span className="font-['Red_Hat_Text',sans-serif] font-medium text-[16px] leading-[24px] text-white">
          {currentLanguage}
        </span>
      </button>

      {isOpen && (
        <div
          className="absolute right-0 top-[calc(100%+16px)] w-[280px] backdrop-blur-[20px] bg-gradient-to-b from-white/[0.06] to-[rgba(153,153,153,0.06)] border border-white/10 rounded-[8px] p-[8px] z-50 flex flex-col gap-[8px]"
          onMouseEnter={handleMouseEnter}
          onMouseLeave={handleMouseLeave}
        >
          <div className="flex items-center gap-[6px] p-[8px] rounded-[8px]">
            <div className="text-[#998cff] w-[20px] h-[20px] flex items-center justify-center">
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
            <span className="font-['Red_Hat_Text',sans-serif] font-medium text-[14px] leading-[21px] text-[#998cff]">
              Language
            </span>
          </div>

          <div className="w-full h-[1px] bg-white/10" />

          {languages.map((lang) => {
            const active = currentLanguage === lang.value
            return (
              <button
                key={lang.value}
                className={`w-full flex items-center justify-between p-[8px] rounded-[6px] cursor-pointer transition-colors border-none outline-none text-left ${
                  active ? 'bg-white/[0.04]' : 'hover:bg-white/[0.04]'
                }`}
                onClick={() => setCurrentLanguage(lang.value)}
              >
                <span className="font-['Red_Hat_Text','Noto_Sans_SC',sans-serif] font-medium text-[16px] leading-[24px] text-white">
                  {lang.label}
                </span>
                {active && (
                  <div className="text-white shrink-0">
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
        className={`font-['Red_Hat_Text',sans-serif] font-normal text-[16px] leading-[24px] transition-colors ${
          active ? 'text-white' : 'text-white/60 hover:text-white'
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
        className={`font-['Red_Hat_Text',sans-serif] font-normal text-[16px] leading-[24px] transition-colors cursor-pointer bg-transparent border-none outline-none ${
          isOpen ? 'text-white' : 'text-white/60 hover:text-white'
        }`}
      >
        {label}
      </button>

      {isOpen && (
        <div
          className={`absolute left-0 top-[calc(100%+12px)] ${getDropdownWidth()} backdrop-blur-[20px] bg-gradient-to-b from-white/[0.06] to-[rgba(153,153,153,0.06)] border border-white/10 rounded-[8px] p-[8px] z-50 flex flex-col gap-[8px]`}
          onMouseEnter={handleMouseEnter}
          onMouseLeave={handleMouseLeave}
        >
          <div className="flex gap-[6px] items-center p-[8px] rounded-[8px]">
            <p className="font-['Red_Hat_Text',sans-serif] font-medium text-[14px] leading-[21px] text-[#998cff]">
              {label}
            </p>
          </div>

          <div className="w-full h-[1px] bg-white/10" />

          {items.map((sub) => (
            <a
              key={sub.label}
              href={sub.href}
              className="flex items-center justify-between p-[8px] rounded-[6px] hover:bg-white/[0.04] transition-colors group cursor-pointer no-underline"
            >
              <span className="font-['Red_Hat_Text',sans-serif] font-medium text-[16px] leading-[24px] text-white">
                {sub.label}
              </span>
              <div className="text-white/40 group-hover:text-white transition-colors shrink-0">
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
  const menuConfig: Record<string, Array<{ label: string; href: string }>> = {
    Products: [
      { label: 'AI Competition', href: '/competition' },
      { label: 'AI Trader', href: '/traders' },
      { label: 'Performance Dashboard', href: '/dashboard' },
    ],
    Resources: [
      { label: 'FAQ', href: '/faq' },
      { label: 'Docs', href: '/docs' },
      { label: 'Github', href: 'https://github.com' },
      { label: 'Security', href: '/security' },
    ],
    Company: [
      { label: 'About', href: '/about' },
      { label: 'Contact', href: '/contact' },
      { label: 'Privacy Policy', href: '/privacy' },
      { label: 'Terms of Use', href: '/terms' },
    ],
  }

  const navItems: NavItemProps[] = [
    { label: 'Home', href: '/', active: true },
    { label: 'Products', href: '#', items: menuConfig.Products },
    { label: 'Resources', href: '#', items: menuConfig.Resources },
    { label: 'Company', href: '#', items: menuConfig.Company },
  ]

  return (
    <header className="relative h-[80px] w-full bg-white/[0.04]">
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
