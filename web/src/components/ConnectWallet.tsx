/**
 * ConnectWallet Button Component
 * Uses Wagmi v2 hooks for Web3 wallet connection
 * 简化版：点击直接连接，类似 PR #19 的体验
 */

import { useState, useRef, useEffect } from 'react'
import { useAccount, useConnect, useDisconnect } from 'wagmi'
import { Wallet, ChevronDown, LogOut } from 'lucide-react'
import { useLanguage } from '../contexts/LanguageContext'

export function ConnectWallet() {
  const { language } = useLanguage()
  const { address, isConnected, connector } = useAccount()
  const { connect, connectors } = useConnect()
  const { disconnect } = useDisconnect()
  const [dropdownOpen, setDropdownOpen] = useState(false)
  const dropdownRef = useRef<HTMLDivElement>(null)

  // Close dropdown when clicking outside
  useEffect(() => {
    function handleClickOutside(event: MouseEvent) {
      if (
        dropdownRef.current &&
        !dropdownRef.current.contains(event.target as Node)
      ) {
        setDropdownOpen(false)
      }
    }

    document.addEventListener('mousedown', handleClickOutside)
    return () => {
      document.removeEventListener('mousedown', handleClickOutside)
    }
  }, [])

  // 简单直接的连接：点击立即连接第一个连接器（类似 PR #19）
  const handleQuickConnect = () => {
    // 优先连接第一个连接器（通常是 Injected/MetaMask）
    const firstConnector = connectors[0]
    if (firstConnector) {
      connect({ connector: firstConnector })
    }
  }

  // If not connected, show connect button (简化版：点击直接连接)
  if (!isConnected) {
    return (
      <div className="relative" ref={dropdownRef}>
        <button
          onClick={handleQuickConnect}
          className="flex items-center gap-2 px-4 py-2 rounded-lg transition-all font-medium text-sm"
          style={{
            background: 'linear-gradient(135deg, #60a5fa 0%, #3b82f6 100%)',
            color: '#ffffff',
            border: '1px solid rgba(96, 165, 250, 0.3)',
          }}
          onMouseEnter={(e) => {
            e.currentTarget.style.transform = 'translateY(-1px)'
            e.currentTarget.style.boxShadow =
              '0 4px 12px rgba(96, 165, 250, 0.3)'
          }}
          onMouseLeave={(e) => {
            e.currentTarget.style.transform = 'translateY(0)'
            e.currentTarget.style.boxShadow = 'none'
          }}
        >
          <Wallet className="h-4 w-4" />
          {language === 'zh' ? '连接钱包' : 'Connect Wallet'}
        </button>
      </div>
    )
  }

  // If connected, show address with disconnect option
  return (
    <div className="relative" ref={dropdownRef}>
      <button
        onClick={() => setDropdownOpen(!dropdownOpen)}
        className="flex items-center gap-2 px-3 py-2 rounded transition-colors"
        style={{
          background: 'rgba(14, 203, 129, 0.1)',
          border: '1px solid rgba(14, 203, 129, 0.3)',
        }}
        onMouseEnter={(e) => {
          e.currentTarget.style.background = 'rgba(14, 203, 129, 0.15)'
        }}
        onMouseLeave={(e) => {
          e.currentTarget.style.background = 'rgba(14, 203, 129, 0.1)'
        }}
      >
        <div
          className="w-2 h-2 rounded-full"
          style={{ background: '#0ECB81' }}
        />
        <code className="text-sm font-mono" style={{ color: '#EAECEF' }}>
          {address?.slice(0, 6)}...{address?.slice(-4)}
        </code>
        <ChevronDown className="h-4 w-4" style={{ color: '#EAECEF' }} />
      </button>

      {dropdownOpen && (
        <div
          className="absolute right-0 top-full mt-2 w-64 rounded-lg shadow-xl overflow-hidden z-50"
          style={{
            background: '#0a0a0a',
            border: '1px solid #2b3139',
          }}
        >
          <div className="p-3">
            <div
              className="text-xs font-medium mb-1"
              style={{ color: '#848E9C' }}
            >
              {language === 'zh' ? '已连接地址' : 'Connected Address'}
            </div>
            <code
              className="block px-3 py-2 rounded font-mono text-xs mb-2 break-all"
              style={{
                background: '#0B0E11',
                color: '#0ECB81',
              }}
            >
              {address}
            </code>
            <div className="text-xs mb-3" style={{ color: '#848E9C' }}>
              {language === 'zh' ? '连接器：' : 'Connected via:'}{' '}
              <span style={{ color: '#EAECEF' }}>{connector?.name}</span>
            </div>
            <button
              onClick={() => {
                disconnect()
                setDropdownOpen(false)
              }}
              className="w-full flex items-center justify-center gap-2 px-3 py-2 rounded transition-colors font-medium text-sm"
              style={{
                background: 'rgba(246, 70, 93, 0.1)',
                color: '#F6465D',
                border: '1px solid rgba(246, 70, 93, 0.2)',
              }}
              onMouseEnter={(e) => {
                e.currentTarget.style.background = 'rgba(246, 70, 93, 0.15)'
              }}
              onMouseLeave={(e) => {
                e.currentTarget.style.background = 'rgba(246, 70, 93, 0.1)'
              }}
            >
              <LogOut className="h-4 w-4" />
              {language === 'zh' ? '断开连接' : 'Disconnect'}
            </button>
          </div>
        </div>
      )}
    </div>
  )
}
