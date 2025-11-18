import React, { useState, useEffect } from 'react'
import type { Exchange } from '../../types'
import { t, type Language } from '../../i18n/translations'
import { api } from '../../lib/api'
import { getExchangeIcon } from '../ExchangeIcons'
import {
  TwoStageKeyModal,
  type TwoStageKeyModalResult,
} from '../TwoStageKeyModal'
import {
  WebCryptoEnvironmentCheck,
  type WebCryptoCheckStatus,
} from '../WebCryptoEnvironmentCheck'
import { BookOpen, Trash2, HelpCircle, RefreshCw } from 'lucide-react'
import { toast } from 'sonner'
import { Tooltip } from './Tooltip'
import { getShortName } from './utils'
import { getAgentWallet, type AgentWallet } from '../../lib/agentWalletBackend'
import { useAccount } from 'wagmi'

interface ExchangeConfigModalProps {
  allExchanges: Exchange[]
  editingExchangeId: string | null
  onSave: (
    exchangeId: string,
    apiKey: string,
    secretKey?: string,
    testnet?: boolean,
    hyperliquidWalletAddr?: string,
    asterUser?: string,
    asterSigner?: string,
    asterPrivateKey?: string
  ) => Promise<void>
  onDelete: (exchangeId: string) => void
  onClose: () => void
  language: Language
}

export function ExchangeConfigModal({
  allExchanges,
  editingExchangeId,
  onSave,
  onDelete,
  onClose,
  language,
}: ExchangeConfigModalProps) {
  const { address } = useAccount()

  const [selectedExchangeId, setSelectedExchangeId] = useState(
    editingExchangeId || ''
  )
  const [apiKey, setApiKey] = useState('')
  const [secretKey, setSecretKey] = useState('')
  const [passphrase, setPassphrase] = useState('')
  const [testnet, setTestnet] = useState(false)
  const [showGuide, setShowGuide] = useState(false)
  const [serverIP, setServerIP] = useState<{
    public_ip: string
    message: string
  } | null>(null)
  const [loadingIP, setLoadingIP] = useState(false)
  const [copiedIP, setCopiedIP] = useState(false)
  const [webCryptoStatus, setWebCryptoStatus] =
    useState<WebCryptoCheckStatus>('idle')

  // 币安配置指南展开状态
  const [showBinanceGuide, setShowBinanceGuide] = useState(false)

  // Aster 特定字段
  const [asterUser, setAsterUser] = useState('')
  const [asterSigner, setAsterSigner] = useState('')
  const [asterPrivateKey, setAsterPrivateKey] = useState('')

  // Hyperliquid 特定字段 - 后端生成的 Agent Wallet
  const [backendAgentWallet, setBackendAgentWallet] =
    useState<AgentWallet | null>(null)
  const [loadingAgentWallet, setLoadingAgentWallet] = useState(false)

  // 安全输入状态
  const [secureInputTarget, setSecureInputTarget] = useState<
    null | 'hyperliquid' | 'aster'
  >(null)

  // 获取当前编辑的交易所信息
  const selectedExchange = allExchanges?.find(
    (e) => e.id === selectedExchangeId
  )

  // 如果是编辑现有交易所，初始化表单数据
  useEffect(() => {
    if (editingExchangeId && selectedExchange) {
      setApiKey(selectedExchange.apiKey || '')
      setSecretKey(selectedExchange.secretKey || '')
      setPassphrase('') // Don't load existing passphrase for security
      setTestnet(selectedExchange.testnet || false)

      // Aster 字段
      setAsterUser(selectedExchange.asterUser || '')
      setAsterSigner(selectedExchange.asterSigner || '')
      setAsterPrivateKey('') // Don't load existing private key for security
    }
  }, [editingExchangeId, selectedExchange])

  // 加载服务器IP（当选择binance时）
  useEffect(() => {
    if (selectedExchangeId === 'binance' && !serverIP) {
      setLoadingIP(true)
      api
        .getServerIP()
        .then((data) => {
          setServerIP(data)
        })
        .catch((err) => {
          console.error('Failed to load server IP:', err)
        })
        .finally(() => {
          setLoadingIP(false)
        })
    }
  }, [selectedExchangeId])

  // 加载后端 Agent Wallet（当选择 Hyperliquid 且连接了钱包时）
  useEffect(() => {
    if (selectedExchangeId === 'hyperliquid' && address && !editingExchangeId) {
      setLoadingAgentWallet(true)
      getAgentWallet(address)
        .then((response) => {
          if (response.success && response.data) {
            setBackendAgentWallet(response.data)
          } else {
            setBackendAgentWallet(null)
          }
        })
        .catch((err) => {
          // 404 means no agent wallet exists, which is fine
          if (!err.message.includes('404')) {
            console.error('Failed to load agent wallet:', err)
          }
          setBackendAgentWallet(null)
        })
        .finally(() => {
          setLoadingAgentWallet(false)
        })
    }
  }, [selectedExchangeId, address, editingExchangeId])

  const handleCopyIP = async (ip: string) => {
    try {
      // 优先使用现代 Clipboard API
      if (navigator.clipboard && navigator.clipboard.writeText) {
        await navigator.clipboard.writeText(ip)
        setCopiedIP(true)
        setTimeout(() => setCopiedIP(false), 2000)
        toast.success(t('ipCopied', language))
      } else {
        // 降级方案: 使用传统的 execCommand 方法
        const textArea = document.createElement('textarea')
        textArea.value = ip
        textArea.style.position = 'fixed'
        textArea.style.left = '-999999px'
        textArea.style.top = '-999999px'
        document.body.appendChild(textArea)
        textArea.focus()
        textArea.select()

        try {
          const successful = document.execCommand('copy')
          if (successful) {
            setCopiedIP(true)
            setTimeout(() => setCopiedIP(false), 2000)
            toast.success(t('ipCopied', language))
          } else {
            throw new Error('复制命令执行失败')
          }
        } finally {
          document.body.removeChild(textArea)
        }
      }
    } catch (err) {
      console.error('复制失败:', err)
      // 显示错误提示
      toast.error(
        t('copyIPFailed', language) || `复制失败: ${ip}\n请手动复制此IP地址`
      )
    }
  }

  // 安全输入处理函数
  const secureInputContextLabel =
    secureInputTarget === 'aster'
      ? t('asterExchangeName', language)
      : secureInputTarget === 'hyperliquid'
        ? t('hyperliquidExchangeName', language)
        : undefined

  const handleSecureInputCancel = () => {
    setSecureInputTarget(null)
  }

  const handleSecureInputComplete = ({
    value,
    obfuscationLog,
  }: TwoStageKeyModalResult) => {
    const trimmed = value.trim()
    if (secureInputTarget === 'hyperliquid') {
      setApiKey(trimmed)
    }
    if (secureInputTarget === 'aster') {
      setAsterPrivateKey(trimmed)
    }
    console.log('Secure input obfuscation log:', obfuscationLog)
    setSecureInputTarget(null)
  }

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!selectedExchangeId) return

    // 根据交易所类型验证不同字段
    if (selectedExchange?.id === 'binance') {
      if (!apiKey.trim() || !secretKey.trim()) return
      await onSave(selectedExchangeId, apiKey.trim(), secretKey.trim(), testnet)
    } else if (selectedExchange?.id === 'hyperliquid') {
      // 使用后端生成的 Agent Wallet（使用特殊标识 "BACKEND_AGENT"）
      if (!backendAgentWallet) {
        toast.error(
          language === 'zh'
            ? '请先创建 Agent Wallet'
            : 'Please create Agent Wallet first'
        )
        return
      }
      await onSave(
        selectedExchangeId,
        'BACKEND_AGENT:' + backendAgentWallet.agent_address,
        '',
        testnet,
        backendAgentWallet.main_wallet
      )
    } else if (selectedExchange?.id === 'aster') {
      if (!asterUser.trim() || !asterSigner.trim() || !asterPrivateKey.trim())
        return
      await onSave(
        selectedExchangeId,
        '',
        '',
        testnet,
        undefined,
        asterUser.trim(),
        asterSigner.trim(),
        asterPrivateKey.trim()
      )
    } else if (selectedExchange?.id === 'okx') {
      if (!apiKey.trim() || !secretKey.trim() || !passphrase.trim()) return
      await onSave(selectedExchangeId, apiKey.trim(), secretKey.trim(), testnet)
    } else {
      // 默认情况（其他CEX交易所）
      if (!apiKey.trim() || !secretKey.trim()) return
      await onSave(selectedExchangeId, apiKey.trim(), secretKey.trim(), testnet)
    }
  }

  // 可选择的交易所列表（所有支持的交易所）
  const availableExchanges = allExchanges || []

  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50 p-4 overflow-y-auto">
      <div
        className="bg-gray-800 rounded-lg w-full max-w-lg relative my-8"
        style={{
          background: '#1E2329',
          maxHeight: 'calc(100vh - 4rem)',
        }}
      >
        <div
          className="flex items-center justify-between p-6 pb-4 sticky top-0 z-10"
          style={{ background: '#1E2329' }}
        >
          <h3 className="text-xl font-bold" style={{ color: '#EAECEF' }}>
            {editingExchangeId
              ? t('editExchange', language)
              : t('addExchange', language)}
          </h3>
          <div className="flex items-center gap-2">
            {selectedExchange?.id === 'binance' && (
              <button
                type="button"
                onClick={() => setShowGuide(true)}
                className="px-3 py-2 rounded text-sm font-semibold transition-all hover:scale-105 flex items-center gap-2"
                style={{
                  background: 'rgba(240, 185, 11, 0.1)',
                  color: '#F0B90B',
                }}
              >
                <BookOpen className="w-4 h-4" />
                {t('viewGuide', language)}
              </button>
            )}
            {editingExchangeId && (
              <button
                type="button"
                onClick={() => onDelete(editingExchangeId)}
                className="p-2 rounded hover:bg-red-100 transition-colors"
                style={{
                  background: 'rgba(246, 70, 93, 0.1)',
                  color: '#F6465D',
                }}
                title={t('delete', language)}
              >
                <Trash2 className="w-4 h-4" />
              </button>
            )}
          </div>
        </div>

        <form onSubmit={handleSubmit} className="px-6 pb-6">
          <div
            className="space-y-4 overflow-y-auto"
            style={{ maxHeight: 'calc(100vh - 16rem)' }}
          >
            {!editingExchangeId && (
              <div className="space-y-3">
                <div className="space-y-2">
                  <div
                    className="text-xs font-semibold uppercase tracking-wide"
                    style={{ color: '#F0B90B' }}
                  >
                    {t('environmentSteps.checkTitle', language)}
                  </div>
                  <WebCryptoEnvironmentCheck
                    language={language}
                    variant="card"
                    onStatusChange={setWebCryptoStatus}
                  />
                </div>
                <div className="space-y-2">
                  <div
                    className="text-xs font-semibold uppercase tracking-wide"
                    style={{ color: '#F0B90B' }}
                  >
                    {t('environmentSteps.selectTitle', language)}
                  </div>
                  <select
                    value={selectedExchangeId}
                    onChange={(e) => setSelectedExchangeId(e.target.value)}
                    className="w-full px-3 py-2 rounded"
                    style={{
                      background: '#0B0E11',
                      border: '1px solid #2B3139',
                      color: '#EAECEF',
                    }}
                    aria-label={t('selectExchange', language)}
                    disabled={webCryptoStatus !== 'secure'}
                    required
                  >
                    <option value="">
                      {t('pleaseSelectExchange', language)}
                    </option>
                    {availableExchanges.map((exchange) => (
                      <option key={exchange.id} value={exchange.id}>
                        {getShortName(exchange.name)} (
                        {exchange.type.toUpperCase()})
                      </option>
                    ))}
                  </select>
                </div>
              </div>
            )}

            {selectedExchange && (
              <div
                className="p-4 rounded"
                style={{ background: '#0B0E11', border: '1px solid #2B3139' }}
              >
                <div className="flex items-center gap-3 mb-3">
                  <div className="w-8 h-8 flex items-center justify-center">
                    {getExchangeIcon(selectedExchange.id, {
                      width: 32,
                      height: 32,
                    })}
                  </div>
                  <div>
                    <div className="font-semibold" style={{ color: '#EAECEF' }}>
                      {getShortName(selectedExchange.name)}
                    </div>
                    <div className="text-xs" style={{ color: '#848E9C' }}>
                      {selectedExchange.type.toUpperCase()} •{' '}
                      {selectedExchange.id}
                    </div>
                  </div>
                </div>
              </div>
            )}

            {selectedExchange && (
              <>
                {/* Binance 和其他 CEX 交易所的字段 */}
                {(selectedExchange.id === 'binance' ||
                  selectedExchange.type === 'cex') &&
                  selectedExchange.id !== 'hyperliquid' &&
                  selectedExchange.id !== 'aster' && (
                    <>
                      {/* 币安用户配置提示 (D1 方案) */}
                      {selectedExchange.id === 'binance' && (
                        <div
                          className="mb-4 p-3 rounded cursor-pointer transition-colors"
                          style={{
                            background: '#1a3a52',
                            border: '1px solid #2b5278',
                          }}
                          onClick={() => setShowBinanceGuide(!showBinanceGuide)}
                        >
                          <div className="flex items-center justify-between">
                            <div className="flex items-center gap-2">
                              <span style={{ color: '#58a6ff' }}>ℹ️</span>
                              <span
                                className="text-sm font-medium"
                                style={{ color: '#EAECEF' }}
                              >
                                <strong>币安用户必读：</strong>
                                使用「现货与合约交易」API，不要用「统一账户
                                API」
                              </span>
                            </div>
                            <span style={{ color: '#8b949e' }}>
                              {showBinanceGuide ? '▲' : '▼'}
                            </span>
                          </div>

                          {/* 展开的详细说明 */}
                          {showBinanceGuide && (
                            <div
                              className="mt-3 pt-3"
                              style={{
                                borderTop: '1px solid #2b5278',
                                fontSize: '0.875rem',
                                color: '#c9d1d9',
                              }}
                              onClick={(e) => e.stopPropagation()}
                            >
                              <p className="mb-2" style={{ color: '#8b949e' }}>
                                <strong>原因：</strong>统一账户 API
                                权限结构不同，会导致订单提交失败
                              </p>

                              <p
                                className="font-semibold mb-1"
                                style={{ color: '#EAECEF' }}
                              >
                                正确配置步骤：
                              </p>
                              <ol
                                className="list-decimal list-inside space-y-1 mb-3"
                                style={{ paddingLeft: '0.5rem' }}
                              >
                                <li>
                                  登录币安 → 个人中心 →{' '}
                                  <strong>API 管理</strong>
                                </li>
                                <li>
                                  创建 API → 选择「
                                  <strong>系统生成的 API 密钥</strong>」
                                </li>
                                <li>
                                  勾选「<strong>现货与合约交易</strong>」（
                                  <span style={{ color: '#f85149' }}>
                                    不选统一账户
                                  </span>
                                  ）
                                </li>
                                <li>
                                  IP 限制选「<strong>无限制</strong>
                                  」或添加服务器 IP
                                </li>
                              </ol>

                              <p
                                className="mb-2 p-2 rounded"
                                style={{
                                  background: '#3d2a00',
                                  border: '1px solid #9e6a03',
                                }}
                              >
                                💡 <strong>多资产模式用户注意：</strong>
                                如果您开启了多资产模式，将强制使用全仓模式。建议关闭多资产模式以支持逐仓交易。
                              </p>

                              <a
                                href="https://www.binance.com/zh-CN/support/faq/how-to-create-api-keys-on-binance-360002502072"
                                target="_blank"
                                rel="noopener noreferrer"
                                className="inline-block text-sm hover:underline"
                                style={{ color: '#58a6ff' }}
                              >
                                📖 查看币安官方教程 ↗
                              </a>
                            </div>
                          )}
                        </div>
                      )}

                      <div>
                        <label
                          className="block text-sm font-semibold mb-2"
                          style={{ color: '#EAECEF' }}
                        >
                          {t('apiKey', language)}
                        </label>
                        <input
                          type="password"
                          value={apiKey}
                          onChange={(e) => setApiKey(e.target.value)}
                          placeholder={t('enterAPIKey', language)}
                          className="w-full px-3 py-2 rounded"
                          style={{
                            background: '#0B0E11',
                            border: '1px solid #2B3139',
                            color: '#EAECEF',
                          }}
                          required
                        />
                      </div>

                      <div>
                        <label
                          className="block text-sm font-semibold mb-2"
                          style={{ color: '#EAECEF' }}
                        >
                          {t('secretKey', language)}
                        </label>
                        <input
                          type="password"
                          value={secretKey}
                          onChange={(e) => setSecretKey(e.target.value)}
                          placeholder={t('enterSecretKey', language)}
                          className="w-full px-3 py-2 rounded"
                          style={{
                            background: '#0B0E11',
                            border: '1px solid #2B3139',
                            color: '#EAECEF',
                          }}
                          required
                        />
                      </div>

                      {selectedExchange.id === 'okx' && (
                        <div>
                          <label
                            className="block text-sm font-semibold mb-2"
                            style={{ color: '#EAECEF' }}
                          >
                            {t('passphrase', language)}
                          </label>
                          <input
                            type="password"
                            value={passphrase}
                            onChange={(e) => setPassphrase(e.target.value)}
                            placeholder={t('enterPassphrase', language)}
                            className="w-full px-3 py-2 rounded"
                            style={{
                              background: '#0B0E11',
                              border: '1px solid #2B3139',
                              color: '#EAECEF',
                            }}
                            required
                          />
                        </div>
                      )}

                      {/* Binance 白名单IP提示 */}
                      {selectedExchange.id === 'binance' && (
                        <div
                          className="p-4 rounded"
                          style={{
                            background: 'rgba(240, 185, 11, 0.1)',
                            border: '1px solid rgba(240, 185, 11, 0.2)',
                          }}
                        >
                          <div
                            className="text-sm font-semibold mb-2"
                            style={{ color: '#F0B90B' }}
                          >
                            {t('whitelistIP', language)}
                          </div>
                          <div
                            className="text-xs mb-3"
                            style={{ color: '#848E9C' }}
                          >
                            {t('whitelistIPDesc', language)}
                          </div>

                          {loadingIP ? (
                            <div
                              className="text-xs"
                              style={{ color: '#848E9C' }}
                            >
                              {t('loadingServerIP', language)}
                            </div>
                          ) : serverIP && serverIP.public_ip ? (
                            <div
                              className="flex items-center gap-2 p-2 rounded"
                              style={{ background: '#0B0E11' }}
                            >
                              <code
                                className="flex-1 text-sm font-mono"
                                style={{ color: '#F0B90B' }}
                              >
                                {serverIP.public_ip}
                              </code>
                              <button
                                type="button"
                                onClick={() => handleCopyIP(serverIP.public_ip)}
                                className="px-3 py-1 rounded text-xs font-semibold transition-all hover:scale-105"
                                style={{
                                  background: 'rgba(240, 185, 11, 0.2)',
                                  color: '#F0B90B',
                                }}
                              >
                                {copiedIP
                                  ? t('ipCopied', language)
                                  : t('copyIP', language)}
                              </button>
                            </div>
                          ) : null}
                        </div>
                      )}
                    </>
                  )}

                {/* Aster 交易所的字段 */}
                {selectedExchange.id === 'aster' && (
                  <>
                    <div>
                      <label
                        className="block text-sm font-semibold mb-2 flex items-center gap-2"
                        style={{ color: '#EAECEF' }}
                      >
                        {t('user', language)}
                        <Tooltip content={t('asterUserDesc', language)}>
                          <HelpCircle
                            className="w-4 h-4 cursor-help"
                            style={{ color: '#F0B90B' }}
                          />
                        </Tooltip>
                      </label>
                      <input
                        type="text"
                        value={asterUser}
                        onChange={(e) => setAsterUser(e.target.value)}
                        placeholder={t('enterUser', language)}
                        className="w-full px-3 py-2 rounded"
                        style={{
                          background: '#0B0E11',
                          border: '1px solid #2B3139',
                          color: '#EAECEF',
                        }}
                        required
                      />
                    </div>

                    <div>
                      <label
                        className="block text-sm font-semibold mb-2 flex items-center gap-2"
                        style={{ color: '#EAECEF' }}
                      >
                        {t('signer', language)}
                        <Tooltip content={t('asterSignerDesc', language)}>
                          <HelpCircle
                            className="w-4 h-4 cursor-help"
                            style={{ color: '#F0B90B' }}
                          />
                        </Tooltip>
                      </label>
                      <input
                        type="text"
                        value={asterSigner}
                        onChange={(e) => setAsterSigner(e.target.value)}
                        placeholder={t('enterSigner', language)}
                        className="w-full px-3 py-2 rounded"
                        style={{
                          background: '#0B0E11',
                          border: '1px solid #2B3139',
                          color: '#EAECEF',
                        }}
                        required
                      />
                    </div>

                    <div>
                      <label
                        className="block text-sm font-semibold mb-2 flex items-center gap-2"
                        style={{ color: '#EAECEF' }}
                      >
                        {t('privateKey', language)}
                        <Tooltip content={t('asterPrivateKeyDesc', language)}>
                          <HelpCircle
                            className="w-4 h-4 cursor-help"
                            style={{ color: '#F0B90B' }}
                          />
                        </Tooltip>
                      </label>
                      <input
                        type="password"
                        value={asterPrivateKey}
                        onChange={(e) => setAsterPrivateKey(e.target.value)}
                        placeholder={t('enterPrivateKey', language)}
                        className="w-full px-3 py-2 rounded"
                        style={{
                          background: '#0B0E11',
                          border: '1px solid #2B3139',
                          color: '#EAECEF',
                        }}
                        required
                      />
                    </div>
                  </>
                )}

                {/* Hyperliquid 交易所的字段 */}
                {selectedExchange.id === 'hyperliquid' && (
                  <>
                    {/* 安全提示 banner */}
                    <div
                      className="p-3 rounded mb-4"
                      style={{
                        background: 'rgba(240, 185, 11, 0.1)',
                        border: '1px solid rgba(240, 185, 11, 0.3)',
                      }}
                    >
                      <div className="flex items-start gap-2">
                        <span style={{ color: '#F0B90B', fontSize: '16px' }}>
                          🔐
                        </span>
                        <div className="flex-1">
                          <div
                            className="text-sm font-semibold mb-1"
                            style={{ color: '#F0B90B' }}
                          >
                            {t('hyperliquidAgentWalletTitle', language)}
                          </div>
                          <div
                            className="text-xs"
                            style={{ color: '#848E9C', lineHeight: '1.5' }}
                          >
                            {t('hyperliquidAgentWalletDesc', language)}
                          </div>
                        </div>
                      </div>
                    </div>

                    {/* 后端 Agent Wallet 状态 */}
                    {!editingExchangeId && (
                      <>
                        {/* 未连接钱包提示 */}
                        {!address && (
                          <div
                            className="p-4 rounded mb-4"
                            style={{
                              background: 'rgba(248, 81, 73, 0.1)',
                              border: '1px solid rgba(248, 81, 73, 0.3)',
                            }}
                          >
                            <div
                              className="text-sm font-semibold mb-2"
                              style={{ color: '#F85149' }}
                            >
                              {language === 'zh'
                                ? '⚠️ 需要连接钱包'
                                : '⚠️ Wallet Connection Required'}
                            </div>
                            <div
                              className="text-xs mb-3"
                              style={{ color: '#848E9C' }}
                            >
                              {language === 'zh'
                                ? '请先访问 Agent Wallet 页面并连接您的 Web3 钱包（MetaMask、WalletConnect 等），然后创建 Agent Wallet。'
                                : 'Please visit the Agent Wallet page and connect your Web3 wallet (MetaMask, WalletConnect, etc.), then create an Agent Wallet.'}
                            </div>
                            <a
                              href="/agent-wallet"
                              target="_blank"
                              rel="noopener noreferrer"
                              className="inline-block px-3 py-2 rounded text-sm font-semibold transition-all hover:scale-105"
                              style={{
                                background: '#F85149',
                                color: '#FFF',
                              }}
                            >
                              {language === 'zh'
                                ? '前往 Agent Wallet 页面 →'
                                : 'Go to Agent Wallet →'}
                            </a>
                          </div>
                        )}

                        {loadingAgentWallet && (
                          <div
                            className="flex items-center gap-2 mb-4 text-xs"
                            style={{ color: '#848E9C' }}
                          >
                            <RefreshCw className="w-3 h-3 animate-spin" />
                            {language === 'zh'
                              ? '正在加载 Agent Wallet...'
                              : 'Loading Agent Wallet...'}
                          </div>
                        )}
                        {!loadingAgentWallet &&
                          !backendAgentWallet &&
                          address && (
                            <div
                              className="p-4 rounded mb-4"
                              style={{
                                background: 'rgba(240, 185, 11, 0.1)',
                                border: '1px solid rgba(240, 185, 11, 0.3)',
                              }}
                            >
                              <div
                                className="text-sm font-semibold mb-2"
                                style={{ color: '#F0B90B' }}
                              >
                                {language === 'zh'
                                  ? '⚠️ 需要创建 Agent Wallet'
                                  : '⚠️ Agent Wallet Required'}
                              </div>
                              <div
                                className="text-xs mb-3"
                                style={{ color: '#848E9C' }}
                              >
                                {language === 'zh'
                                  ? 'Hyperliquid 需要使用后端生成的 Agent Wallet。请先前往 Agent Wallet 页面创建并授权。'
                                  : 'Hyperliquid requires a backend-generated Agent Wallet. Please create and authorize one in the Agent Wallet page.'}
                              </div>
                              <a
                                href="/agent-wallet"
                                target="_blank"
                                rel="noopener noreferrer"
                                className="inline-block px-3 py-2 rounded text-sm font-semibold transition-all hover:scale-105"
                                style={{
                                  background: '#F0B90B',
                                  color: '#000',
                                }}
                              >
                                {language === 'zh'
                                  ? '前往创建 Agent Wallet →'
                                  : 'Go to Agent Wallet →'}
                              </a>
                            </div>
                          )}
                      </>
                    )}

                    {/* 显示 Agent Wallet 信息 */}
                    {backendAgentWallet ? (
                      /* 使用后端生成的 Agent Wallet */
                      <>
                        <div
                          className="p-4 rounded mb-4"
                          style={{
                            background: 'rgba(14, 203, 129, 0.1)',
                            border: '1px solid rgba(14, 203, 129, 0.2)',
                          }}
                        >
                          <div className="space-y-2">
                            <div>
                              <span
                                className="text-xs"
                                style={{ color: '#848E9C' }}
                              >
                                {language === 'zh'
                                  ? 'Agent 地址：'
                                  : 'Agent Address:'}
                              </span>
                              <code
                                className="block mt-1 px-2 py-1 rounded font-mono text-xs"
                                style={{
                                  background: '#0B0E11',
                                  color: '#0ECB81',
                                }}
                              >
                                {backendAgentWallet.agent_address}
                              </code>
                            </div>
                            <div>
                              <span
                                className="text-xs"
                                style={{ color: '#848E9C' }}
                              >
                                {language === 'zh'
                                  ? 'Main Wallet：'
                                  : 'Main Wallet:'}
                              </span>
                              <code
                                className="block mt-1 px-2 py-1 rounded font-mono text-xs"
                                style={{
                                  background: '#0B0E11',
                                  color: '#0ECB81',
                                }}
                              >
                                {backendAgentWallet.main_wallet}
                              </code>
                            </div>
                            <div>
                              <span
                                className="text-xs"
                                style={{ color: '#848E9C' }}
                              >
                                {language === 'zh' ? '状态：' : 'Status:'}
                              </span>
                              <span
                                className="ml-2 px-2 py-0.5 rounded text-xs font-medium"
                                style={{
                                  background: backendAgentWallet.status === 'ACTIVE'
                                    ? 'rgba(14, 203, 129, 0.2)'
                                    : 'rgba(240, 185, 11, 0.2)',
                                  color: backendAgentWallet.status === 'ACTIVE' ? '#0ECB81' : '#F0B90B',
                                }}
                              >
                                {backendAgentWallet.status}
                              </span>
                            </div>
                          </div>

                          {/* 如果状态不是 ACTIVE，显示授权提示 */}
                          {backendAgentWallet.status !== 'ACTIVE' && (
                            <div
                              className="mt-3 p-3 rounded"
                              style={{
                                background: 'rgba(240, 185, 11, 0.1)',
                                border: '1px solid rgba(240, 185, 11, 0.3)',
                              }}
                            >
                              <p className="text-xs mb-2" style={{ color: '#F0B90B' }}>
                                <strong>⚠️ {language === 'zh' ? '需要授权' : 'Authorization Required'}</strong>
                              </p>
                              <p className="text-xs mb-3" style={{ color: '#848E9C' }}>
                                {language === 'zh'
                                  ? 'Agent 钱包需要完成授权才能使用。点击下方按钮前往授权页面（需要 2 次签名）。'
                                  : 'Agent wallet needs to be authorized before use. Click the button below to go to the authorization page (2 signatures required).'}
                              </p>
                              <a
                                href="/agent-wallet-backend"
                                target="_blank"
                                rel="noopener noreferrer"
                                className="inline-flex items-center gap-2 px-4 py-2 rounded text-xs font-semibold transition-all"
                                style={{
                                  background: '#F0B90B',
                                  color: '#000',
                                  textDecoration: 'none',
                                }}
                                onMouseEnter={(e) => {
                                  e.currentTarget.style.transform = 'translateY(-1px)'
                                  e.currentTarget.style.boxShadow = '0 4px 12px rgba(240, 185, 11, 0.3)'
                                }}
                                onMouseLeave={(e) => {
                                  e.currentTarget.style.transform = 'translateY(0)'
                                  e.currentTarget.style.boxShadow = 'none'
                                }}
                              >
                                🔐 {language === 'zh' ? '前往授权 Agent 钱包' : 'Go to Authorize Agent Wallet'}
                              </a>
                            </div>
                          )}
                        </div>

                        {/* Main Wallet Address (自动填充) */}
                        <div>
                          <label
                            className="block text-sm font-semibold mb-2"
                            style={{ color: '#EAECEF' }}
                          >
                            {t('hyperliquidMainWalletAddress', language)}
                          </label>
                          <input
                            type="text"
                            value={backendAgentWallet.main_wallet}
                            readOnly
                            className="w-full px-3 py-2 rounded"
                            style={{
                              background: '#0B0E11',
                              border: '1px solid #2B3139',
                              color: '#848E9C',
                              cursor: 'not-allowed',
                            }}
                          />
                          <div
                            className="text-xs mt-1"
                            style={{ color: '#848E9C' }}
                          >
                            {language === 'zh'
                              ? '使用后端托管的 Agent Wallet，无需手动输入私钥'
                              : 'Using backend-hosted Agent Wallet, no private key input needed'}
                          </div>
                        </div>
                      </>
                    ) : null}
                  </>
                )}
              </>
            )}
          </div>

          <div
            className="flex gap-3 mt-6 pt-4 sticky bottom-0"
            style={{ background: '#1E2329' }}
          >
            <button
              type="button"
              onClick={onClose}
              className="flex-1 px-4 py-2 rounded text-sm font-semibold"
              style={{ background: '#2B3139', color: '#848E9C' }}
            >
              {t('cancel', language)}
            </button>
            <button
              type="submit"
              disabled={
                !selectedExchange ||
                (selectedExchange.id === 'binance' &&
                  (!apiKey.trim() || !secretKey.trim())) ||
                (selectedExchange.id === 'okx' &&
                  (!apiKey.trim() ||
                    !secretKey.trim() ||
                    !passphrase.trim())) ||
                (selectedExchange.id === 'hyperliquid' &&
                  (!backendAgentWallet ||
                    backendAgentWallet.status !== 'ACTIVE')) || // 验证后端 Agent Wallet 存在且已授权
                (selectedExchange.id === 'aster' &&
                  (!asterUser.trim() ||
                    !asterSigner.trim() ||
                    !asterPrivateKey.trim())) ||
                (selectedExchange.type === 'cex' &&
                  selectedExchange.id !== 'hyperliquid' &&
                  selectedExchange.id !== 'aster' &&
                  selectedExchange.id !== 'binance' &&
                  selectedExchange.id !== 'okx' &&
                  (!apiKey.trim() || !secretKey.trim()))
              }
              className="flex-1 px-4 py-2 rounded text-sm font-semibold disabled:opacity-50"
              style={{ background: '#F0B90B', color: '#000' }}
            >
              {t('saveConfig', language)}
            </button>
          </div>
        </form>
      </div>

      {/* Binance Setup Guide Modal */}
      {showGuide && (
        <div
          className="fixed inset-0 bg-black bg-opacity-75 flex items-center justify-center z-50 p-4"
          onClick={() => setShowGuide(false)}
        >
          <div
            className="bg-gray-800 rounded-lg p-6 w-full max-w-4xl relative"
            style={{ background: '#1E2329' }}
            onClick={(e) => e.stopPropagation()}
          >
            <div className="flex items-center justify-between mb-4">
              <h3
                className="text-xl font-bold flex items-center gap-2"
                style={{ color: '#EAECEF' }}
              >
                <BookOpen className="w-6 h-6" style={{ color: '#F0B90B' }} />
                {t('binanceSetupGuide', language)}
              </h3>
              <button
                onClick={() => setShowGuide(false)}
                className="px-4 py-2 rounded text-sm font-semibold transition-all hover:scale-105"
                style={{ background: '#2B3139', color: '#848E9C' }}
              >
                {t('closeGuide', language)}
              </button>
            </div>
            <div className="overflow-y-auto max-h-[80vh]">
              <img
                src="/images/guide.png"
                alt={t('binanceSetupGuide', language)}
                className="w-full h-auto rounded"
              />
            </div>
          </div>
        </div>
      )}

      {/* Two Stage Key Modal */}
      <TwoStageKeyModal
        isOpen={secureInputTarget !== null}
        language={language}
        contextLabel={secureInputContextLabel}
        expectedLength={64}
        onCancel={handleSecureInputCancel}
        onComplete={handleSecureInputComplete}
      />
    </div>
  )
}
