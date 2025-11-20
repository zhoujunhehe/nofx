import { useState } from 'react'
import { motion, AnimatePresence } from 'framer-motion'
import {
  X,
  ChevronRight,
  ChevronLeft,
  ExternalLink,
  Copy,
  Check,
  Coins,
  Globe,
  Link,
  ArrowDownToLine,
  Key,
  Waves,
} from 'lucide-react'
import { useLanguage } from '../contexts/LanguageContext'
import { t } from '../i18n/translations'

interface HyperLiquidDepositModalProps {
  isOpen: boolean
  onClose: () => void
}

export function HyperLiquidDepositModal({
  isOpen,
  onClose,
}: HyperLiquidDepositModalProps) {
  const { language } = useLanguage()
  const [currentStep, setCurrentStep] = useState(0)
  const [copied, setCopied] = useState(false)

  if (!isOpen) return null

  const steps = [
    {
      title: '准备 USDC (Arbitrum)',
      description: '确保您的钱包在 Arbitrum One 网络上有 USDC 余额。',
      icon: <Coins className="w-8 h-8 text-[#F0B90B]" />,
      content: (
        <div className="space-y-4">
          <p className="text-[#848E9C] text-sm">
            HyperLiquid 仅支持 Arbitrum One 网络上的 USDC。
          </p>
          <div className="bg-[#1E2329] p-4 rounded-lg border border-[#2B3139]">
            <div className="flex items-center justify-between mb-2">
              <span className="text-[#EAECEF] text-sm">网络</span>
              <span className="text-[#F0B90B] text-sm font-medium">
                Arbitrum One
              </span>
            </div>
            <div className="flex items-center justify-between">
              <span className="text-[#EAECEF] text-sm">代币</span>
              <span className="text-[#F0B90B] text-sm font-medium">
                USDC (Native)
              </span>
            </div>
          </div>
        </div>
      ),
    },
    {
      title: '访问 HyperLiquid',
      description: '前往 HyperLiquid 官方网站并连接钱包。',
      icon: <Globe className="w-8 h-8 text-[#F0B90B]" />,
      content: (
        <div className="space-y-4">
          <a
            href="https://app.hyperliquid.xyz/join/AITRADING"
            target="_blank"
            rel="noopener noreferrer"
            className="flex items-center justify-between bg-[#1E2329] p-4 rounded-lg border border-[#2B3139] hover:border-[#F0B90B] transition-colors group"
          >
            <div className="flex items-center gap-3">
              <div className="w-10 h-10 rounded-full bg-[#0B0E11] flex items-center justify-center text-xl">
                H
              </div>
              <div>
                <div className="text-[#EAECEF] font-medium group-hover:text-[#F0B90B] transition-colors">
                  app.hyperliquid.xyz
                </div>
                <div className="text-[#848E9C] text-xs">点击打开官网</div>
              </div>
            </div>
            <ExternalLink className="w-5 h-5 text-[#848E9C] group-hover:text-[#F0B90B] transition-colors" />
          </a>
        </div>
      ),
    },
    {
      title: '连接钱包',
      description: '在 HyperLiquid 页面右上角点击 "Connect Wallet"。',
      icon: <Link className="w-8 h-8 text-[#F0B90B]" />,
      content: (
        <div className="space-y-4">
          <div className="relative bg-[#1E2329] rounded-lg border border-[#2B3139] overflow-hidden aspect-video flex items-center justify-center">
            <div className="absolute inset-0 bg-gradient-to-br from-[#1E2329] to-[#0B0E11]" />
            <div className="relative z-10 flex flex-col items-center gap-2">
              <div className="px-4 py-2 bg-[#F0B90B] text-black font-bold rounded shadow-lg shadow-[#F0B90B]/20 animate-pulse">
                Connect Wallet
              </div>
              <span className="text-[#848E9C] text-xs">点击右上角按钮</span>
            </div>
          </div>
        </div>
      ),
    },
    {
      title: '存入资金',
      description: '点击 "Deposit" 按钮并输入金额。',
      icon: <ArrowDownToLine className="w-8 h-8 text-[#F0B90B]" />,
      content: (
        <div className="space-y-4">
          <div className="bg-[#1E2329] p-4 rounded-lg border border-[#2B3139]">
            <div className="flex items-center gap-2 mb-3">
              <div className="w-2 h-2 rounded-full bg-[#F0B90B]" />
              <span className="text-[#EAECEF] text-sm">Enable Trading</span>
            </div>
            <p className="text-[#848E9C] text-xs leading-relaxed">
              首次使用需要先点击 "Enable Trading" 进行授权，然后点击 "Deposit"
              存入 USDC。
            </p>
          </div>
        </div>
      ),
    },
  ]

  const handleNext = () => {
    if (currentStep < steps.length - 1) {
      setCurrentStep((prev) => prev + 1)
    } else {
      onClose()
    }
  }

  const handlePrev = () => {
    if (currentStep > 0) {
      setCurrentStep((prev) => prev - 1)
    }
  }

  return (
    <div
      className="fixed inset-0 z-[60] flex items-center justify-center bg-black/80 backdrop-blur-sm p-4"
      onClick={onClose}
    >
      <motion.div
        initial={{ opacity: 0, scale: 0.95, y: 20 }}
        animate={{ opacity: 1, scale: 1, y: 0 }}
        exit={{ opacity: 0, scale: 0.95, y: 20 }}
        onClick={(e) => e.stopPropagation()}
        className="bg-[#0B0E11] border border-[#2B3139] rounded-2xl shadow-2xl w-full max-w-md overflow-hidden flex flex-col max-h-[90vh]"
      >
        {/* Header */}
        <div className="p-6 border-b border-[#2B3139] flex items-center justify-between bg-[#1E2329]">
          <div>
            <h3 className="text-xl font-bold text-[#EAECEF] flex items-center gap-2">
              <Waves className="w-6 h-6 text-[#F0B90B]" /> HyperLiquid 充值教程
            </h3>
            <p className="text-[#848E9C] text-sm mt-1">
              步骤 {currentStep + 1} / {steps.length}
            </p>
          </div>
          <button
            onClick={onClose}
            className="w-8 h-8 rounded-lg text-[#848E9C] hover:text-[#EAECEF] hover:bg-[#2B3139] transition-colors flex items-center justify-center"
          >
            <X className="w-5 h-5" />
          </button>
        </div>

        {/* Content */}
        <div className="flex-1 overflow-y-auto p-6">
          <AnimatePresence mode="wait">
            <motion.div
              key={currentStep}
              initial={{ opacity: 0, x: 20 }}
              animate={{ opacity: 1, x: 0 }}
              exit={{ opacity: 0, x: -20 }}
              transition={{ duration: 0.2 }}
              className="space-y-6"
            >
              <div className="flex flex-col items-center text-center space-y-4">
                <div className="w-16 h-16 rounded-2xl bg-[#1E2329] border border-[#2B3139] flex items-center justify-center shadow-lg shadow-black/50">
                  {steps[currentStep].icon}
                </div>
                <div>
                  <h4 className="text-lg font-bold text-[#EAECEF] mb-2">
                    {steps[currentStep].title}
                  </h4>
                  <p className="text-[#848E9C] text-sm leading-relaxed">
                    {steps[currentStep].description}
                  </p>
                </div>
              </div>

              <div className="pt-4 border-t border-[#2B3139]/50">
                {steps[currentStep].content}
              </div>
            </motion.div>
          </AnimatePresence>
        </div>

        {/* Footer */}
        <div className="p-6 border-t border-[#2B3139] bg-[#1E2329] flex items-center justify-between gap-4">
          <button
            onClick={handlePrev}
            disabled={currentStep === 0}
            className={`flex items-center gap-2 px-4 py-2 rounded-lg text-sm font-medium transition-colors ${
              currentStep === 0
                ? 'text-[#2B3139] cursor-not-allowed'
                : 'text-[#848E9C] hover:text-[#EAECEF] hover:bg-[#2B3139]'
            }`}
          >
            <ChevronLeft className="w-4 h-4" />
            上一步
          </button>

          <div className="flex gap-1">
            {steps.map((_, index) => (
              <div
                key={index}
                className={`w-2 h-2 rounded-full transition-colors ${
                  index === currentStep ? 'bg-[#F0B90B]' : 'bg-[#2B3139]'
                }`}
              />
            ))}
          </div>

          <button
            onClick={handleNext}
            className="flex items-center gap-2 px-6 py-2 bg-[#F0B90B] text-black rounded-lg text-sm font-bold hover:bg-[#E1A706] transition-colors shadow-lg shadow-[#F0B90B]/20"
          >
            {currentStep === steps.length - 1 ? '完成' : '下一步'}
            {currentStep !== steps.length - 1 && (
              <ChevronRight className="w-4 h-4" />
            )}
          </button>
        </div>
      </motion.div>
    </div>
  )
}
