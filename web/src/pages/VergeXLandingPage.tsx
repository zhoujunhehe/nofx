import React from 'react'
import {
  SecurityIcon,
  NaturalLanguageIcon,
  IntelligentBacktestIcon,
  AlwaysOnIcon,
  GithubIcon,
} from '../components/vergex/svg-assets'
import { VergeXHeader } from '../components/vergex/VergeXHeader'

export function VergeXLandingPage() {

  return (
    <div className="bg-[#080808] min-h-screen relative overflow-hidden">
      {/* ============= HERO SECTION ============= */}
      <section className="relative h-[960px] w-full">
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
        <div className="w-[1440px] mx-auto relative h-full">
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
          <svg className="absolute left-[140px] top-[286px]" width="1" height="238">
            <line x1="0" y1="0" x2="0" y2="238" stroke="white" strokeOpacity="0.2" strokeWidth="1" />
          </svg>
          <svg className="absolute left-[1088px] top-[286px]" width="1" height="108">
            <line x1="0" y1="0" x2="0" y2="108" stroke="white" strokeOpacity="0.2" strokeWidth="1" />
          </svg>
          <svg className="absolute left-[307px] top-[167px]" width="1" height="307">
            <line x1="0" y1="0" x2="0" y2="307" stroke="white" strokeOpacity="0.2" strokeWidth="1" />
          </svg>
          <svg className="absolute left-[1236px] top-[384px]" width="1" height="183">
            <line x1="0" y1="0" x2="0" y2="183" stroke="white" strokeOpacity="0.2" strokeWidth="1" />
          </svg>
          <svg className="absolute left-[1363px] top-[137px]" width="1" height="339">
            <line x1="0" y1="0" x2="0" y2="339" stroke="white" strokeOpacity="0.2" strokeWidth="1" />
          </svg>
        </div>

        {/* Header */}
        <div className="absolute top-0 left-0 right-0 z-50">
          <VergeXHeader />
        </div>

        {/* Hero Content */}
        <div className="w-[1440px] mx-auto relative h-full">
          <div className="absolute left-[64px] top-[700px] w-[976px] flex flex-col gap-[8px] items-start">
            <h1 className="font-['Poppins',sans-serif] font-medium text-[64px] leading-[80px] text-white w-full">
              The <span className="text-[#998cff]">Agentic</span> AI Trading Platform
            </h1>
            <p className="font-['Red_Hat_Text',sans-serif] font-normal text-[20px] leading-[36px] text-white w-full">
              Vergex is an agentic AI trading platform that lets large language models manage your
              positions. On a fixed schedule it reads live market data plus your prompt, decides
              whether to open, adjust, or close trades, and executes under your risk constraints.
            </p>
          </div>

          {/* Open App Button */}
          <button className="absolute left-[1221px] top-[772px] bg-[#998cff] rounded-[32px] px-[32px] py-[12px] flex items-center gap-[8px] hover:bg-[#a89dff] transition-colors">
            <span className="font-['Red_Hat_Text',sans-serif] font-medium text-[18px] leading-[28px] text-black">
              Open App
            </span>
            <svg width="24" height="24" viewBox="0 0 24 24" className="rotate-90 scale-y-[-1]">
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

      {/* ============= WHY VERGEX SECTION ============= */}
      <section className="w-full">
        <div className="w-[1440px] mx-auto px-[64px] py-[100px]">
        <h2 className="font-['Red_Hat_Text',sans-serif] font-semibold text-[48px] leading-[72px] text-white mb-[64px]">
          Why Vergex?
        </h2>

        <div className="flex gap-[32px]">
          {/* Large Card - Security First */}
          <div className="w-[527px] border border-white/10 rounded-[12px] p-[32px] relative overflow-hidden">
            {/* Background Decorative SVG */}
            <div className="absolute bottom-[61.73px] right-[-433.1px] w-[877.097px] h-[735.273px] opacity-10">
              <svg width="877" height="735" viewBox="0 0 877 735" fill="none">
                <circle cx="438" cy="367" r="300" stroke="#998cff" strokeWidth="1" opacity="0.3" />
                <circle cx="438" cy="367" r="250" stroke="#998cff" strokeWidth="1" opacity="0.2" />
                <circle cx="438" cy="367" r="200" stroke="#998cff" strokeWidth="1" opacity="0.1" />
              </svg>
            </div>

            <div className="relative z-10 flex flex-col justify-between h-full">
              <SecurityIcon />

              <div className="flex flex-col gap-[4px] mt-auto">
                <h3 className="font-['Red_Hat_Text',sans-serif] font-medium text-[36px] leading-[54px] text-white">
                  Security First
                </h3>
                <p className="font-['Red_Hat_Text',sans-serif] font-light text-[20px] leading-[30px] text-white/60">
                  Built on trade-only API permissions with no withdrawal access, keeping your funds
                  safe and fully under your control.
                </p>
              </div>
            </div>
          </div>

          {/* Right Column */}
          <div className="flex-1 flex flex-col gap-[32px]">
            {/* Top Row */}
            <div className="flex gap-[32px]">
              {/* Natural Language */}
              <div className="flex-1 border border-white/10 rounded-[12px] p-[24px]">
                <NaturalLanguageIcon />
                <div className="flex flex-col gap-[4px] mt-[16px]">
                  <h3 className="font-['Red_Hat_Text',sans-serif] font-medium text-[24px] leading-[36px] text-white">
                    Natural Language
                  </h3>
                  <p className="font-['Red_Hat_Text',sans-serif] font-light text-[18px] leading-[27px] text-white/60">
                    Describe your trading logic in natual language—no scripting or coding required
                    for the AI to operate.
                  </p>
                </div>
              </div>

              {/* Intelligent Backtest */}
              <div className="flex-1 border border-white/10 rounded-[12px] p-[24px]">
                <IntelligentBacktestIcon />
                <div className="flex flex-col gap-[4px] mt-[16px]">
                  <h3 className="font-['Red_Hat_Text',sans-serif] font-medium text-[24px] leading-[36px] text-white">
                    Intelligent Backtest
                  </h3>
                  <p className="font-['Red_Hat_Text',sans-serif] font-light text-[18px] leading-[27px] text-white/60">
                    Preview how your prompt would behave in past market conditions before committing
                    real capital.
                  </p>
                </div>
              </div>
            </div>

            {/* Always-On Execution - Full Width */}
            <div className="border border-white/10 rounded-[12px] p-[24px]">
              <AlwaysOnIcon />
              <div className="flex flex-col gap-[4px] mt-[16px]">
                <h3 className="font-['Red_Hat_Text',sans-serif] font-medium text-[24px] leading-[36px] text-white">
                  Always-On Execution
                </h3>
                <p className="font-['Red_Hat_Text',sans-serif] font-light text-[18px] leading-[27px] text-white/60">
                  Vergex evaluates markets on a fixed schedule and manages positions 24/7 so you
                  don't have to monitor the screen.
                </p>
              </div>
            </div>
          </div>
        </div>
        </div>
      </section>

      {/* ============= OUR PRODUCTS SECTION ============= */}
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

        <h2 className="font-['Red_Hat_Text',sans-serif] font-semibold text-[48px] leading-[72px] text-white mb-[64px] relative z-10">
          Our Products
        </h2>

        <div className="flex items-center justify-between relative z-10">
          {/* Architecture Diagram */}
          <div className="w-[588.897px] h-[620px] relative">
            {/* This would be the isometric diagram - using placeholder */}
            <div className="w-full h-full bg-gradient-to-br from-white/[0.05] to-transparent border border-white/10 rounded-[12px] flex items-center justify-center">
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
                  <circle cx="12" cy="8" r="4" stroke="white" strokeWidth="1.5" />
                  <path
                    d="M6 21C6 17.686 8.686 15 12 15C15.314 15 18 17.686 18 21"
                    stroke="white"
                    strokeWidth="1.5"
                    strokeLinecap="round"
                  />
                </svg>
                <h3 className="font-['Red_Hat_Text',sans-serif] font-semibold text-[24px] leading-[36px] tracking-[2.4px] uppercase text-white">
                  User Layer
                </h3>
              </div>
              <p className="font-['Red_Hat_Text',sans-serif] font-light text-[20px] leading-[30px] text-white/50">
                Your trading intent, written in natural language. Risk parameters and trade-only API
                keys stay fully under your control.
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
                <h3 className="font-['Red_Hat_Text',sans-serif] font-semibold text-[24px] leading-[36px] tracking-[2.4px] uppercase text-white">
                  VergeX Agentic
                </h3>
              </div>
              <p className="font-['Red_Hat_Text',sans-serif] font-light text-[20px] leading-[30px] text-white/50">
                An autonomous decision engine that evaluates markets on a fixed schedule, interprets
                your prompt, and manages positions within your risk limits.
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
                <h3 className="font-['Red_Hat_Text',sans-serif] font-semibold text-[24px] leading-[36px] tracking-[2.4px] uppercase text-white">
                  Trading Venues
                </h3>
              </div>
              <p className="font-['Red_Hat_Text',sans-serif] font-light text-[20px] leading-[30px] text-white/50">
                Connected directly to your existing exchange accounts — Binance, Hyperliquid, Aster,
                and more.
              </p>
            </div>

            {/* AI Models */}
            <div className="flex flex-col gap-[12px]">
              <div className="flex items-center gap-[12px]">
                <svg width="24" height="24" viewBox="0 0 24 24" fill="none">
                  <circle cx="12" cy="12" r="3" stroke="white" strokeWidth="1.5" />
                  <path
                    d="M12 3V5M12 19V21M21 12H19M5 12H3M18.364 5.636L16.95 7.05M7.05 16.95L5.636 18.364M18.364 18.364L16.95 16.95M7.05 7.05L5.636 5.636"
                    stroke="white"
                    strokeWidth="1.5"
                    strokeLinecap="round"
                  />
                </svg>
                <h3 className="font-['Red_Hat_Text',sans-serif] font-semibold text-[24px] leading-[36px] tracking-[2.4px] uppercase text-white">
                  Ai Models
                </h3>
              </div>
              <p className="font-['Red_Hat_Text',sans-serif] font-light text-[20px] leading-[30px] text-white/50">
                Powered by leading large language models, including DeepSeek, Qwen, Gemini, GPT,
                Claude, and other agentic AI systems.
              </p>
            </div>
          </div>
        </div>
        </div>
      </section>

      {/* ============= SOCIAL MEDIA SECTION ============= */}
      <section className="w-full">
        <div className="w-[1440px] mx-auto px-[64px] py-[100px]">
        <h2 className="font-['Red_Hat_Text',sans-serif] font-semibold text-[48px] leading-[72px] text-white mb-[64px]">
          Social Media
        </h2>

        <div className="flex items-center justify-between mb-[64px]">
          {/* GitHub Card */}
          <div className="w-[864px] border-[1.5px] border-white/10 rounded-[12px] p-[48px] bg-white/[0.04] relative overflow-hidden">
            {/* Background Decorative SVG */}
            <div className="absolute bottom-[-0.27px] right-[-320.1px] w-[877.097px] h-[735.273px] opacity-5">
              <svg width="877" height="735" viewBox="0 0 877 735" fill="none">
                <circle cx="438" cy="367" r="300" stroke="#998cff" strokeWidth="1" />
                <circle cx="438" cy="367" r="250" stroke="#998cff" strokeWidth="1" />
              </svg>
            </div>

            <div className="relative z-10">
              <GithubIcon />

              <div className="flex gap-[48px] mt-[48px] mb-[48px]">
                <div>
                  <p className="font-['Red_Hat_Text',sans-serif] font-semibold text-[40px] leading-[60px] text-white">
                    7,859
                  </p>
                  <p className="font-['Red_Hat_Text',sans-serif] font-light text-[18px] leading-[28px] text-[#98989a]">
                    Stars
                  </p>
                </div>
                <div>
                  <p className="font-['Red_Hat_Text',sans-serif] font-semibold text-[40px] leading-[60px] text-white">
                    1,985
                  </p>
                  <p className="font-['Red_Hat_Text',sans-serif] font-light text-[18px] leading-[28px] text-[#98989a]">
                    Forks
                  </p>
                </div>
                <div>
                  <p className="font-['Red_Hat_Text',sans-serif] font-semibold text-[40px] leading-[60px] text-white">
                    44
                  </p>
                  <p className="font-['Red_Hat_Text',sans-serif] font-light text-[18px] leading-[28px] text-[#98989a]">
                    Contraibutors
                  </p>
                </div>
              </div>

              <button className="bg-[#998cff] rounded-[32px] h-[60px] px-[32px] py-[12px] flex items-center gap-[8px] hover:bg-[#a89dff] transition-colors">
                <span className="font-['Red_Hat_Text',sans-serif] font-medium text-[20px] leading-[30px] text-black">
                  Contribute on GitHub
                </span>
                <svg width="24" height="24" viewBox="0 0 24 24" className="rotate-90 scale-y-[-1]">
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
          </div>

          {/* Social Links Card */}
          <div className="w-[416px] border-[1.5px] border-white/10 rounded-[12px] px-[48px] py-[56px] flex flex-col gap-[80px]">
            {/* Twitter */}
            <a
              href="#"
              className="flex items-center gap-[24px] group hover:opacity-80 transition-opacity"
            >
              <div className="w-[56px] h-[56px] bg-blue-500 rounded-full flex items-center justify-center">
                <svg width="28" height="28" viewBox="0 0 24 24" fill="white">
                  <path d="M23 3a10.9 10.9 0 01-3.14 1.53 4.48 4.48 0 00-7.86 3v1A10.66 10.66 0 013 4s-4 9 5 13a11.64 11.64 0 01-7 2c9 5 20 0 20-11.5a4.5 4.5 0 00-.08-.83A7.72 7.72 0 0023 3z" />
                </svg>
              </div>
              <span className="flex-1 font-['Red_Hat_Text',sans-serif] font-medium text-[28px] leading-[42px] text-white">
                Twitter
              </span>
              <svg
                width="24"
                height="24"
                viewBox="0 0 24 24"
                className="group-hover:translate-x-1 group-hover:-translate-y-1 transition-transform"
              >
                <path
                  d="M7 17L17 7M17 7H7M17 7V17"
                  stroke="white"
                  strokeWidth="2"
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  fill="none"
                />
              </svg>
            </a>

            {/* Telegram */}
            <a
              href="#"
              className="flex items-center gap-[24px] group hover:opacity-80 transition-opacity"
            >
              <div className="w-[56px] h-[56px] bg-cyan-500 rounded-full flex items-center justify-center">
                <svg width="28" height="28" viewBox="0 0 24 24" fill="white">
                  <path d="M21 3L3 10.5L10 13.5M21 3L13.5 21L10 13.5M21 3L10 13.5" />
                </svg>
              </div>
              <span className="flex-1 font-['Red_Hat_Text',sans-serif] font-medium text-[28px] leading-[42px] text-white">
                Telegram
              </span>
              <svg
                width="24"
                height="24"
                viewBox="0 0 24 24"
                className="group-hover:translate-x-1 group-hover:-translate-y-1 transition-transform"
              >
                <path
                  d="M7 17L17 7M17 7H7M17 7V17"
                  stroke="white"
                  strokeWidth="2"
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  fill="none"
                />
              </svg>
            </a>

            {/* Discord */}
            <a
              href="#"
              className="flex items-center gap-[24px] group hover:opacity-80 transition-opacity"
            >
              <div className="w-[56px] h-[56px] bg-indigo-600 rounded-full flex items-center justify-center">
                <svg width="28" height="28" viewBox="0 0 24 24" fill="white">
                  <path d="M9 12h.01M15 12h.01M8 21H4a2 2 0 01-2-2V5a2 2 0 012-2h16a2 2 0 012 2v10a2 2 0 01-2 2h-3.5L12 21l-4-0z" />
                </svg>
              </div>
              <span className="flex-1 font-['Red_Hat_Text',sans-serif] font-medium text-[28px] leading-[42px] text-white">
                Discord
              </span>
              <svg
                width="24"
                height="24"
                viewBox="0 0 24 24"
                className="group-hover:translate-x-1 group-hover:-translate-y-1 transition-transform"
              >
                <path
                  d="M7 17L17 7M17 7H7M17 7V17"
                  stroke="white"
                  strokeWidth="2"
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  fill="none"
                />
              </svg>
            </a>
          </div>
        </div>

        {/* Community Posts */}
        <div className="pt-[64px] flex flex-col gap-[80px]">
          {/* Post 1 */}
          <div className="flex gap-[48px] items-start">
            <div className="w-[384px] flex items-center gap-[16px]">
              <div className="w-[80px] h-[80px] rounded-[42px] overflow-hidden">
                <img src="/vergex/avatar-1.png" alt="" className="w-full h-full object-cover" />
              </div>
              <div>
                <p className="font-['Red_Hat_Text',sans-serif] font-semibold text-[20px] leading-[30px] text-white">
                  John Techson
                </p>
                <p className="font-['Red_Hat_Text',sans-serif] font-light text-[18px] leading-[27px] text-white/60">
                  Quantum Computing
                </p>
              </div>
            </div>

            <div className="flex-1 flex flex-col gap-[8px]">
              <p className="font-['Red_Hat_Text',sans-serif] font-normal text-[20px] leading-[30px] text-white/60">
                October 15, 2025
              </p>
              <p className="font-['Red_Hat_Text',sans-serif] font-normal text-[20px] leading-[36px] text-white">
                "Open-source NOFX revives the legendary Alpha Arena, an AI-powered crypto futures
                battleground. Built on DeepSeek/Qwen AI, it trades live on Binance, Hyperliquid,
                and Aster DEX, featuring multi-AI battles and self-learning bots"
              </p>
            </div>

            <button className="bg-white/[0.04] border border-white/[0.06] rounded-[12px] h-[64px] px-[24px] flex items-center gap-[10px] hover:bg-white/[0.08] transition-colors">
              <span className="font-['Red_Hat_Text',sans-serif] font-normal text-[18px] leading-[27px] text-white">
                View Blog
              </span>
              <svg width="24" height="24" viewBox="0 0 24 24">
                <path
                  d="M7 17L17 7M17 7H7M17 7V17"
                  stroke="white"
                  strokeWidth="2"
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  fill="none"
                />
              </svg>
            </button>
          </div>

          <div className="h-[1px] bg-white/10" />

          {/* Post 2 */}
          <div className="flex gap-[48px] items-start">
            <div className="w-[384px] flex items-center gap-[16px]">
              <div className="w-[80px] h-[80px] rounded-[42px] overflow-hidden">
                <img src="/vergex/avatar-2.png" alt="" className="w-full h-full object-cover" />
              </div>
              <div>
                <p className="font-['Red_Hat_Text',sans-serif] font-semibold text-[20px] leading-[30px] text-white">
                  Sarah Ethicist
                </p>
                <p className="font-['Red_Hat_Text',sans-serif] font-light text-[18px] leading-[27px] text-white/60">
                  AI Ethics
                </p>
              </div>
            </div>

            <div className="flex-1 flex flex-col gap-[8px]">
              <p className="font-['Red_Hat_Text',sans-serif] font-normal text-[20px] leading-[30px] text-white/60">
                October 16, 2025
              </p>
              <p className="font-['Red_Hat_Text','Noto_Sans_SC',sans-serif] font-normal text-[20px] leading-[36px] text-white">
                "跑了一晚上 @nofx_ai 开源的 AI 自动交易,太有意思了,就看 AI
                在那一会开空一会开多,一顿操作,虽然看不懂为什么,但是一晚上帮我赚了 6% 收益"
              </p>
            </div>

            <button className="bg-white/[0.04] border border-white/[0.06] rounded-[12px] h-[64px] px-[24px] flex items-center gap-[10px] hover:bg-white/[0.08] transition-colors">
              <span className="font-['Red_Hat_Text',sans-serif] font-normal text-[18px] leading-[27px] text-white">
                View Blog
              </span>
              <svg width="24" height="24" viewBox="0 0 24 24">
                <path
                  d="M7 17L17 7M17 7H7M17 7V17"
                  stroke="white"
                  strokeWidth="2"
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  fill="none"
                />
              </svg>
            </button>
          </div>

          <div className="h-[1px] bg-white/10" />

          {/* Post 3 */}
          <div className="flex gap-[48px] items-start">
            <div className="w-[384px] flex items-center gap-[16px]">
              <div className="w-[80px] h-[80px] rounded-[42px] overflow-hidden">
                <img src="/vergex/avatar-3.png" alt="" className="w-full h-full object-cover" />
              </div>
              <div>
                <p className="font-['Red_Hat_Text',sans-serif] font-semibold text-[20px] leading-[30px] text-white">
                  Astronomer X
                </p>
                <p className="font-['Red_Hat_Text',sans-serif] font-light text-[18px] leading-[27px] text-white/60">
                  Space Exploration
                </p>
              </div>
            </div>

            <div className="flex-1 flex flex-col gap-[8px]">
              <p className="font-['Red_Hat_Text',sans-serif] font-normal text-[20px] leading-[30px] text-white/60">
                October 30, 2025
              </p>
              <p className="font-['Red_Hat_Text','Noto_Sans_SC',sans-serif] font-normal text-[20px] leading-[36px] text-white">
                "前不久非常火的 AI 量化交易系统 NOF1,在 GitHub 上有人将其复刻并开源,这就是 NOFX
                项目。基于 DeepSeek、Qwen 等大语言模型,打造的通用架构 AI
                交易操作系统,完成了从决策、到交易、再到复盘的闭环。GitHub:
                https://github.com/NoFxAiOS/nofx"
              </p>
            </div>

            <button className="bg-white/[0.04] border border-white/[0.06] rounded-[12px] h-[64px] px-[24px] flex items-center gap-[10px] hover:bg-white/[0.08] transition-colors">
              <span className="font-['Red_Hat_Text',sans-serif] font-normal text-[18px] leading-[27px] text-white">
                View Blog
              </span>
              <svg width="24" height="24" viewBox="0 0 24 24">
                <path
                  d="M7 17L17 7M17 7H7M17 7V17"
                  stroke="white"
                  strokeWidth="2"
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  fill="none"
                />
              </svg>
            </button>
          </div>
        </div>
        </div>
      </section>

      {/* ============= FOOTER ============= */}
      <footer className="border-t border-white/10 bg-white/[0.02] w-full">
        <div className="w-[1440px] mx-auto">
        <div className="px-[64px] py-[48px]">
          <div className="flex justify-between mb-[32px]">
            {/* Products */}
            <div className="flex flex-col gap-[12px]">
              <p className="font-['Red_Hat_Text',sans-serif] font-normal text-[14px] leading-[21px] text-white/40">
                Products
              </p>
              <a
                href="#"
                className="font-['Red_Hat_Text',sans-serif] font-normal text-[16px] leading-[24px] text-white hover:text-[#998cff] transition-colors"
              >
                AI Competition
              </a>
              <a
                href="#"
                className="font-['Red_Hat_Text',sans-serif] font-normal text-[16px] leading-[24px] text-white hover:text-[#998cff] transition-colors"
              >
                AI Trader
              </a>
              <a
                href="#"
                className="font-['Red_Hat_Text',sans-serif] font-normal text-[16px] leading-[24px] text-white hover:text-[#998cff] transition-colors"
              >
                Performance Dashboard
              </a>
            </div>

            {/* Resources */}
            <div className="flex flex-col gap-[12px]">
              <p className="font-['Red_Hat_Text',sans-serif] font-normal text-[14px] leading-[21px] text-white/40">
                Resources
              </p>
              <a
                href="#"
                className="font-['Red_Hat_Text',sans-serif] font-normal text-[16px] leading-[24px] text-white hover:text-[#998cff] transition-colors"
              >
                FAQ
              </a>
              <a
                href="#"
                className="font-['Red_Hat_Text',sans-serif] font-normal text-[16px] leading-[24px] text-white hover:text-[#998cff] transition-colors"
              >
                Dosc
              </a>
              <a
                href="#"
                className="font-['Red_Hat_Text',sans-serif] font-normal text-[16px] leading-[24px] text-white hover:text-[#998cff] transition-colors"
              >
                Github
              </a>
              <a
                href="#"
                className="font-['Red_Hat_Text',sans-serif] font-normal text-[16px] leading-[24px] text-white hover:text-[#998cff] transition-colors"
              >
                Security
              </a>
            </div>

            {/* Company */}
            <div className="flex flex-col gap-[12px]">
              <p className="font-['Red_Hat_Text',sans-serif] font-normal text-[14px] leading-[21px] text-white/40">
                Company
              </p>
              <a
                href="#"
                className="font-['Red_Hat_Text',sans-serif] font-normal text-[16px] leading-[24px] text-white hover:text-[#998cff] transition-colors"
              >
                About
              </a>
              <a
                href="#"
                className="font-['Red_Hat_Text',sans-serif] font-normal text-[16px] leading-[24px] text-white hover:text-[#998cff] transition-colors"
              >
                Contact
              </a>
              <a
                href="#"
                className="font-['Red_Hat_Text',sans-serif] font-normal text-[16px] leading-[24px] text-white hover:text-[#998cff] transition-colors"
              >
                Privacy Policy
              </a>
              <a
                href="#"
                className="font-['Red_Hat_Text',sans-serif] font-normal text-[16px] leading-[24px] text-white hover:text-[#998cff] transition-colors"
              >
                Terms of Use
              </a>
            </div>

            {/* Supporters */}
            <div className="flex flex-col gap-[12px]">
              <p className="font-['Red_Hat_Text',sans-serif] font-normal text-[14px] leading-[21px] text-white/40">
                Supporters
              </p>
              <a
                href="#"
                className="font-['Red_Hat_Text',sans-serif] font-normal text-[16px] leading-[24px] text-white hover:text-[#998cff] transition-colors"
              >
                Amber.ac (Strategic Investment)
              </a>
              <a
                href="#"
                className="font-['Red_Hat_Text',sans-serif] font-normal text-[16px] leading-[24px] text-white hover:text-[#998cff] transition-colors"
              >
                Aster DEX
              </a>
              <a
                href="#"
                className="font-['Red_Hat_Text',sans-serif] font-normal text-[16px] leading-[24px] text-white hover:text-[#998cff] transition-colors"
              >
                Hyperliquid
              </a>
              <a
                href="#"
                className="font-['Red_Hat_Text',sans-serif] font-normal text-[16px] leading-[24px] text-white hover:text-[#998cff] transition-colors"
              >
                Binance
              </a>
            </div>

            {/* Follow us */}
            <div className="w-[192px] flex flex-col gap-[12px]">
              <p className="font-['Red_Hat_Text',sans-serif] font-normal text-[14px] leading-[21px] text-white/40">
                Follow us
              </p>
              <div className="flex gap-[24px]">
                <a href="#" className="hover:opacity-80 transition-opacity">
                  <svg width="36" height="36" viewBox="0 0 36 36" fill="white">
                    <path
                      fillRule="evenodd"
                      clipRule="evenodd"
                      d="M18 4C10.268 4 4 10.268 4 18C4 25.732 10.268 32 18 32C25.732 32 32 25.732 32 18C32 10.268 25.732 4 18 4ZM16 24H14V15H16V24ZM15 13.5C14.172 13.5 13.5 12.828 13.5 12C13.5 11.172 14.172 10.5 15 10.5C15.828 10.5 16.5 11.172 16.5 12C16.5 12.828 15.828 13.5 15 13.5ZM23 24H21V19.5C21 18.672 20.328 18 19.5 18C18.672 18 18 18.672 18 19.5V24H16V15H18V16.07C18.526 15.403 19.337 15 20.25 15C21.769 15 23 16.231 23 17.75V24Z"
                    />
                  </svg>
                </a>
                <a href="#" className="hover:opacity-80 transition-opacity">
                  <svg width="36" height="36" viewBox="0 0 36 36" fill="white">
                    <path d="M28 10c-1 .5-2 .8-3 1 1-.6 1.8-1.6 2-2.8-1 .6-2 1-3.1 1.2-1-1-2.3-1.6-3.9-1.6-3 0-5.4 2.4-5.4 5.4 0 .4 0 .8.1 1.2-4.5-.2-8.5-2.4-11.2-5.7-.5.8-.7 1.7-.7 2.7 0 1.9 1 3.5 2.4 4.5-.9 0-1.7-.3-2.4-.7v.1c0 2.6 1.9 4.8 4.3 5.3-.4.1-.9.2-1.4.2-.3 0-.7 0-1-.1.7 2.1 2.6 3.6 4.9 3.6-1.8 1.4-4.1 2.2-6.5 2.2-.4 0-.8 0-1.3-.1 2.3 1.5 5 2.3 8 2.3 9.6 0 14.8-7.9 14.8-14.8v-.7c1-.7 1.9-1.6 2.6-2.6z" />
                  </svg>
                </a>
                <a href="#" className="hover:opacity-80 transition-opacity">
                  <svg width="36" height="36" viewBox="0 0 36 36" fill="white">
                    <circle cx="18" cy="18" r="14" stroke="white" strokeWidth="2" fill="none" />
                    <path d="M14 16L16 18L22 12" stroke="white" strokeWidth="2" fill="none" />
                  </svg>
                </a>
              </div>
            </div>
          </div>

          <p className="font-['Red_Hat_Text',sans-serif] font-normal text-[12px] leading-[18px] text-white/40">
            DISCLAIMER: Vergex does not custody user funds and operates only with trade-only API
            permissions. Cryptocurrency trading carries risk. Please assess carefully before
            participating.
            <br />
            <br />© 2025 Vergex. All rights reserved.
          </p>
        </div>
        </div>
      </footer>
    </div>
  )
}

export default VergeXLandingPage
