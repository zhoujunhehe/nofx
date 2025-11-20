import React from 'react'

export const FooterSection: React.FC = () => {
  return (
    <footer className="border-t border-vergex-border-light bg-vergex-bg-secondary w-full">
      <div className="w-[1440px] mx-auto">
        <div className="px-[64px] py-[48px]">
          <div className="flex justify-between mb-[32px]">
            {/* Products */}
            <div className="flex flex-col gap-[12px]">
              <p className="font-vergex-body font-normal text-[14px] leading-[21px] text-vergex-text-primary/40">
                Products
              </p>
              <a
                href="#"
                className="font-vergex-body font-normal text-[16px] leading-[24px] text-vergex-text-primary hover:text-vergex-primary transition-colors"
              >
                AI Competition
              </a>
              <a
                href="#"
                className="font-vergex-body font-normal text-[16px] leading-[24px] text-vergex-text-primary hover:text-vergex-primary transition-colors"
              >
                AI Trader
              </a>
              <a
                href="#"
                className="font-vergex-body font-normal text-[16px] leading-[24px] text-vergex-text-primary hover:text-vergex-primary transition-colors"
              >
                Performance Dashboard
              </a>
            </div>

            {/* Resources */}
            <div className="flex flex-col gap-[12px]">
              <p className="font-vergex-body font-normal text-[14px] leading-[21px] text-vergex-text-primary/40">
                Resources
              </p>
              <a
                href="#"
                className="font-vergex-body font-normal text-[16px] leading-[24px] text-vergex-text-primary hover:text-vergex-primary transition-colors"
              >
                FAQ
              </a>
              <a
                href="#"
                className="font-vergex-body font-normal text-[16px] leading-[24px] text-vergex-text-primary hover:text-vergex-primary transition-colors"
              >
                Docs
              </a>
              <a
                href="#"
                className="font-vergex-body font-normal text-[16px] leading-[24px] text-vergex-text-primary hover:text-vergex-primary transition-colors"
              >
                Github
              </a>
              <a
                href="#"
                className="font-vergex-body font-normal text-[16px] leading-[24px] text-vergex-text-primary hover:text-vergex-primary transition-colors"
              >
                Security
              </a>
            </div>

            {/* Company */}
            <div className="flex flex-col gap-[12px]">
              <p className="font-vergex-body font-normal text-[14px] leading-[21px] text-vergex-text-primary/40">
                Company
              </p>
              <a
                href="#"
                className="font-vergex-body font-normal text-[16px] leading-[24px] text-vergex-text-primary hover:text-vergex-primary transition-colors"
              >
                About
              </a>
              <a
                href="#"
                className="font-vergex-body font-normal text-[16px] leading-[24px] text-vergex-text-primary hover:text-vergex-primary transition-colors"
              >
                Contact
              </a>
              <a
                href="#"
                className="font-vergex-body font-normal text-[16px] leading-[24px] text-vergex-text-primary hover:text-vergex-primary transition-colors"
              >
                Privacy Policy
              </a>
              <a
                href="#"
                className="font-vergex-body font-normal text-[16px] leading-[24px] text-vergex-text-primary hover:text-vergex-primary transition-colors"
              >
                Terms of Use
              </a>
            </div>

            {/* Supporters */}
            <div className="flex flex-col gap-[12px]">
              <p className="font-vergex-body font-normal text-[14px] leading-[21px] text-vergex-text-primary/40">
                Supporters
              </p>
              <a
                href="#"
                className="font-vergex-body font-normal text-[16px] leading-[24px] text-vergex-text-primary hover:text-vergex-primary transition-colors"
              >
                Amber.ac (Strategic Investment)
              </a>
              <a
                href="#"
                className="font-vergex-body font-normal text-[16px] leading-[24px] text-vergex-text-primary hover:text-vergex-primary transition-colors"
              >
                Aster DEX
              </a>
              <a
                href="#"
                className="font-vergex-body font-normal text-[16px] leading-[24px] text-vergex-text-primary hover:text-vergex-primary transition-colors"
              >
                Hyperliquid
              </a>
              <a
                href="#"
                className="font-vergex-body font-normal text-[16px] leading-[24px] text-vergex-text-primary hover:text-vergex-primary transition-colors"
              >
                Binance
              </a>
            </div>

            {/* Follow us */}
            <div className="w-[192px] flex flex-col gap-[12px]">
              <p className="font-vergex-body font-normal text-[14px] leading-[21px] text-vergex-text-primary/40">
                Follow us
              </p>
              <div className="flex gap-[24px]">
                <a href="#" className="hover:opacity-80 transition-opacity">
                  <svg
                    width="36"
                    height="36"
                    viewBox="0 0 36 36"
                    fill="white"
                  >
                    <path
                      fillRule="evenodd"
                      clipRule="evenodd"
                      d="M18 4C10.268 4 4 10.268 4 18C4 25.732 10.268 32 18 32C25.732 32 32 25.732 32 18C32 10.268 25.732 4 18 4ZM16 24H14V15H16V24ZM15 13.5C14.172 13.5 13.5 12.828 13.5 12C13.5 11.172 14.172 10.5 15 10.5C15.828 10.5 16.5 11.172 16.5 12C16.5 12.828 15.828 13.5 15 13.5ZM23 24H21V19.5C21 18.672 20.328 18 19.5 18C18.672 18 18 18.672 18 19.5V24H16V15H18V16.07C18.526 15.403 19.337 15 20.25 15C21.769 15 23 16.231 23 17.75V24Z"
                    />
                  </svg>
                </a>
                <a href="#" className="hover:opacity-80 transition-opacity">
                  <svg
                    width="36"
                    height="36"
                    viewBox="0 0 36 36"
                    fill="white"
                  >
                    <path d="M28 10c-1 .5-2 .8-3 1 1-.6 1.8-1.6 2-2.8-1 .6-2 1-3.1 1.2-1-1-2.3-1.6-3.9-1.6-3 0-5.4 2.4-5.4 5.4 0 .4 0 .8.1 1.2-4.5-.2-8.5-2.4-11.2-5.7-.5.8-.7 1.7-.7 2.7 0 1.9 1 3.5 2.4 4.5-.9 0-1.7-.3-2.4-.7v.1c0 2.6 1.9 4.8 4.3 5.3-.4.1-.9.2-1.4.2-.3 0-.7 0-1-.1.7 2.1 2.6 3.6 4.9 3.6-1.8 1.4-4.1 2.2-6.5 2.2-.4 0-.8 0-1.3-.1 2.3 1.5 5 2.3 8 2.3 9.6 0 14.8-7.9 14.8-14.8v-.7c1-.7 1.9-1.6 2.6-2.6z" />
                  </svg>
                </a>
                <a href="#" className="hover:opacity-80 transition-opacity">
                  <svg
                    width="36"
                    height="36"
                    viewBox="0 0 36 36"
                    fill="white"
                  >
                    <circle
                      cx="18"
                      cy="18"
                      r="14"
                      stroke="white"
                      strokeWidth="2"
                      fill="none"
                    />
                    <path
                      d="M14 16L16 18L22 12"
                      stroke="white"
                      strokeWidth="2"
                      fill="none"
                    />
                  </svg>
                </a>
              </div>
            </div>
          </div>

          <p className="font-vergex-body font-normal text-[12px] leading-[18px] text-vergex-text-primary/40">
            DISCLAIMER: Vergex does not custody user funds and operates only
            with trade-only API permissions. Cryptocurrency trading carries
            risk. Please assess carefully before participating.
            <br />
            <br />© 2025 Vergex. All rights reserved.
          </p>
        </div>
      </div>
    </footer>
  )
}
