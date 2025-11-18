/**
 * Agent Wallet Backend Generation Page
 * 后端生成 Agent 钱包（后端托管私钥）
 *
 * 统一设计风格：Binance 黑色主题
 * 整合两次签名：ApproveAgent + ApproveBuilderFee (0.1%)
 */

import { useState, useEffect } from 'react'
import { Wallet, RefreshCw, Check, AlertCircle, ArrowRight, Shield, Key } from 'lucide-react'
import { useAccount, useWalletClient } from 'wagmi'
import {
  createAgentWallet,
  getAgentWallet,
  authorizeAgent,
  confirmBuilderFee,
  type AgentWallet,
} from '../lib/agentWalletBackend'
import { signApproveAgent } from '../lib/hyperliquidApproveAgent'
import { approveHyperliquidBuilderFee } from '../lib/hyperliquidBuilderFee'
import { useLanguage } from '../contexts/LanguageContext'

const BUILDER_ADDRESS = '0x891dc6f05ad47a3c1a05da55e7a7517971faaf0d' // NOFX Builder Address

export function AgentWalletBackendPage() {
  const { language } = useLanguage()
  const { address, isConnected } = useAccount()
  const { data: walletClient } = useWalletClient()

  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [authStatus, setAuthStatus] = useState<string | null>(null)
  const [agentWallet, setAgentWallet] = useState<AgentWallet | null>(null)

  // 查询现有的 Agent 钱包
  useEffect(() => {
    if (address && isConnected) {
      loadAgentWallet()
    }
  }, [address, isConnected])

  const loadAgentWallet = async () => {
    if (!address) return

    try {
      setLoading(true)
      const response = await getAgentWallet(address)
      if (response.success && response.data) {
        setAgentWallet(response.data)
      }
    } catch (err: any) {
      // 404 means no agent wallet exists yet, which is fine
      if (!err.message.includes('404')) {
        console.error('Failed to load agent wallet:', err)
      }
    } finally {
      setLoading(false)
    }
  }

  const handleCreateAgent = async () => {
    if (!address || !isConnected) {
      setError(language === 'zh' ? '请先连接钱包' : 'Please connect wallet first')
      return
    }

    try {
      setLoading(true)
      setError(null)
      setAuthStatus(null)

      const response = await createAgentWallet(address, 'Mainnet')

      if (response.success) {
        setAuthStatus(
          language === 'zh'
            ? `成功创建 Agent 钱包！地址：${response.agent_address}`
            : `Agent wallet created! Address: ${response.agent_address}`
        )
        // 重新加载
        await loadAgentWallet()
      } else {
        setError(response.message)
      }
    } catch (err: any) {
      console.error('Create agent wallet failed:', err)
      setError(err.message || 'Unknown error')
    } finally {
      setLoading(false)
    }
  }

  const handleAuthorizeAgent = async () => {
    if (!address || !isConnected || !walletClient || !agentWallet) {
      setError(language === 'zh' ? '请先连接钱包並创建 Agent' : 'Please connect wallet and create Agent first')
      return
    }

    try {
      setLoading(true)
      setError(null)
      setAuthStatus(null)

      // ===== 步骤 1: 用户签名 ApproveAgent 消息 =====
      setAuthStatus(language === 'zh' ? '步骤 1/2: 授权 Agent 钱包...' : 'Step 1/2: Authorizing Agent Wallet...')

      const { signature, signatureHex, nonce } = await signApproveAgent(walletClient, {
        agentAddress: agentWallet.agent_address,
        agentName: '', // 可选：可以添加 UI 讓用户输入 Agent 名稱
        hyperliquidChain: agentWallet.hyperliquid_chain as 'Mainnet' | 'Testnet',
      })

      // 提交到后端
      const response = await authorizeAgent({
        main_wallet: address,
        signature: signatureHex,
        agent_name: '',
        nonce,
        signature_rsv: signature,
      })

      if (!response.success) {
        setError(response.message)
        return
      }

      console.log('✅ Step 1/2: Agent Wallet authorized')

      // ===== 步骤 2: 授权 Builder Fee (0.1%) =====
      setAuthStatus(language === 'zh' ? '步骤 2/2: 授权平台费率 (0.1%)...' : 'Step 2/2: Authorizing Platform Fee (0.1%)...')

      const builderFeeResult = await approveHyperliquidBuilderFee(walletClient, {
        builderAddress: BUILDER_ADDRESS,
        maxFeeRate: 100, // 固定 0.1%
        hyperliquidChain: agentWallet.hyperliquid_chain as 'Mainnet' | 'Testnet',
      })

      if (!builderFeeResult.success) {
        // Agent 授权成功但 Builder Fee 失敗
        setAuthStatus(
          language === 'zh'
            ? `✅ Agent 授权成功\n⚠️ Builder Fee 授权失敗: ${builderFeeResult.message}`
            : `✅ Agent authorized\n⚠️ Builder Fee failed: ${builderFeeResult.message}`
        )
        await loadAgentWallet()
        return
      }

      console.log('✅ Step 2/2: Builder Fee authorized')

      // ===== 步骤 3: 通知后端 Builder Fee 已授权 =====
      setAuthStatus(language === 'zh' ? '步骤 3/3: 保存授权状态...' : 'Step 3/3: Saving authorization status...')

      try {
        await confirmBuilderFee(address, 100) // 100 基点 = 0.1%
        console.log('✅ Step 3/3: Builder Fee confirmed in backend')
      } catch (confirmError: any) {
        console.warn('⚠️  Builder Fee confirmation failed (non-critical):', confirmError)
        // 不阻塞流程，因为链上授权已成功
      }

      // ===== 三个步骤都完成 =====
      setAuthStatus(
        language === 'zh'
          ? '✅ 完整授权成功！Agent 钱包和平台费率 (0.1%) 已授权。'
          : '✅ Full authorization completed! Agent wallet and platform fee (0.1%) authorized.'
      )

      // 重新加载状态
      await loadAgentWallet()

      // 3秒后清除消息
      setTimeout(() => setAuthStatus(null), 3000)
    } catch (err: any) {
      console.error('Authorize agent wallet failed:', err)
      setError(err.message || 'Unknown error')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="min-h-screen p-8" style={{ background: '#000000', color: '#EAECEF' }}>
      <div className="max-w-4xl mx-auto">
        {/* Header */}
        <div className="mb-8">
          <h1 className="text-3xl font-bold mb-2" style={{ color: '#EAECEF' }}>
            {language === 'zh' ? 'Agent 钱包（后端生成）' : 'Agent Wallet (Backend Generated)'}
          </h1>
          <p style={{ color: '#848E9C' }}>
            {language === 'zh'
              ? '后端生成並托管 Agent 钱包，用户无需保存私钥'
              : 'Backend generates and hosts Agent wallet, no private key management needed'}
          </p>
        </div>

        {/* Connection Status */}
        <div
          className="rounded-xl p-6 mb-6"
          style={{ background: '#0a0a0a', border: '1px solid #2b3139' }}
        >
          <div className="flex items-center gap-3 mb-4">
            <Wallet className="h-5 w-5" style={{ color: '#60a5fa' }} />
            <h2 className="text-lg font-semibold" style={{ color: '#EAECEF' }}>
              {language === 'zh' ? '钱包状态' : 'Wallet Status'}
            </h2>
          </div>

          {isConnected && address ? (
            <div className="space-y-2">
              <div className="flex items-center gap-2">
                <Check className="h-4 w-4" style={{ color: '#0ECB81' }} />
                <span className="text-sm" style={{ color: '#EAECEF' }}>
                  {language === 'zh' ? '已连接：' : 'Connected: '}
                  <code
                    className="ml-1 px-2 py-1 rounded text-sm"
                    style={{ background: '#0B0E11', color: '#60a5fa' }}
                  >
                    {address.slice(0, 10)}...{address.slice(-8)}
                  </code>
                </span>
              </div>
            </div>
          ) : (
            <div className="flex items-center gap-2" style={{ color: '#F0B90B' }}>
              <AlertCircle className="h-4 w-4" />
              <span className="text-sm">
                {language === 'zh' ? '请使用右上角按钮连接钱包' : 'Please connect wallet using button above'}
              </span>
            </div>
          )}
        </div>

        {/* Agent Wallet Status */}
        {agentWallet ? (
          <div
            className="rounded-xl p-6 mb-6"
            style={{ background: 'rgba(14, 203, 129, 0.1)', border: '1px solid rgba(14, 203, 129, 0.2)' }}
          >
            <div className="flex items-center gap-3 mb-4">
              <Shield className="h-5 w-5" style={{ color: '#0ECB81' }} />
              <h2 className="text-lg font-semibold" style={{ color: '#0ECB81' }}>
                {language === 'zh' ? 'Agent 钱包已创建' : 'Agent Wallet Created'}
              </h2>
            </div>

            <div className="space-y-3">
              <div>
                <span className="text-sm" style={{ color: '#848E9C' }}>
                  {language === 'zh' ? 'Agent 地址：' : 'Agent Address:'}
                </span>
                <code
                  className="block mt-1 px-3 py-2 rounded font-mono text-sm"
                  style={{ background: '#0B0E11', color: '#0ECB81' }}
                >
                  {agentWallet.agent_address}
                </code>
              </div>

              <div>
                <span className="text-sm" style={{ color: '#848E9C' }}>
                  {language === 'zh' ? '状态：' : 'Status:'}
                </span>
                <span
                  className="ml-2 px-2 py-1 rounded text-xs font-medium"
                  style={{
                    background:
                      agentWallet.status === 'ACTIVE' ? 'rgba(14, 203, 129, 0.2)' : 'rgba(240, 185, 11, 0.2)',
                    color: agentWallet.status === 'ACTIVE' ? '#0ECB81' : '#F0B90B',
                  }}
                >
                  {agentWallet.status}
                </span>
              </div>

              <div>
                <span className="text-sm" style={{ color: '#848E9C' }}>
                  {language === 'zh' ? '创建时间：' : 'Created At:'}
                </span>
                <span className="ml-2 text-sm" style={{ color: '#EAECEF' }}>
                  {new Date(agentWallet.created_at).toLocaleString()}
                </span>
              </div>

              {/* Authorization Button - only show when status is INIT */}
              {agentWallet.status === 'INIT' && (
                <div className="mt-4 pt-4" style={{ borderTop: '1px solid #2b3139' }}>
                  {/* 授权说明 */}
                  <div
                    className="rounded-lg p-4 mb-4"
                    style={{ background: '#0B0E11', border: '1px solid #2B3139' }}
                  >
                    <p className="text-sm mb-2" style={{ color: '#EAECEF' }}>
                      <strong>{language === 'zh' ? '授权包含 2 个步骤：' : 'Authorization includes 2 steps:'}</strong>
                    </p>
                    <ul className="text-sm space-y-1" style={{ color: '#848E9C' }}>
                      <li>
                        1️⃣ {language === 'zh' ? 'Agent 钱包授权 - 允许 AI 代理交易' : 'Agent Wallet Authorization - Enable AI trading'}
                      </li>
                      <li>
                        2️⃣ {language === 'zh' ? '平台费率授权 (0.1%) - 确保平台稳定运营' : 'Platform Fee (0.1%) - Ensure platform stability'}
                      </li>
                    </ul>
                  </div>

                  <p className="text-sm mb-3" style={{ color: '#848E9C' }}>
                    {language === 'zh'
                      ? 'Agent 钱包已创建，但尚未在 Hyperliquid 上授权。点击下方按钮进行授权（2次签名）。'
                      : 'Agent wallet created but not yet authorized on Hyperliquid. Click below to authorize (2 signatures required).'}
                  </p>
                  <button
                    onClick={handleAuthorizeAgent}
                    disabled={loading}
                    className="flex items-center gap-2 px-4 py-2 rounded-lg transition-colors text-white font-medium text-sm"
                    style={{
                      background: loading
                        ? 'linear-gradient(135deg, #6B7280 0%, #4B5563 100%)'
                        : 'linear-gradient(135deg, #0ECB81 0%, #0aa66a 100%)',
                      cursor: loading ? 'not-allowed' : 'pointer',
                      opacity: loading ? 0.7 : 1,
                    }}
                  >
                    {loading ? (
                      <>
                        <RefreshCw className="h-4 w-4 animate-spin" />
                        {language === 'zh' ? '授权中...' : 'Authorizing...'}
                      </>
                    ) : (
                      <>
                        <Key className="h-4 w-4" />
                        {language === 'zh' ? '立即授权（2次签名）' : 'Authorize Now (2 Signatures)'}
                      </>
                    )}
                  </button>
                </div>
              )}
            </div>
          </div>
        ) : (
          <div
            className="rounded-xl p-6 mb-6"
            style={{ background: '#0a0a0a', border: '1px solid #2b3139' }}
          >
            <div className="flex items-center gap-3 mb-4">
              <AlertCircle className="h-5 w-5" style={{ color: '#F0B90B' }} />
              <h2 className="text-lg font-semibold" style={{ color: '#EAECEF' }}>
                {language === 'zh' ? '尚未创建 Agent 钱包' : 'No Agent Wallet Yet'}
              </h2>
            </div>

            <p className="text-sm mb-4" style={{ color: '#848E9C' }}>
              {language === 'zh'
                ? '点击下方按钮创建 Agent 钱包。私钥将由后端生成并加密存储，您无需保存任何私钥。'
                : 'Click the button below to create an Agent wallet. The private key will be generated and encrypted by the backend, no manual key management needed.'}
            </p>

            <button
              onClick={handleCreateAgent}
              disabled={!isConnected || loading}
              className="flex items-center gap-2 px-6 py-3 rounded-lg transition-colors font-medium disabled:opacity-50 disabled:cursor-not-allowed"
              style={{
                background: loading
                  ? 'linear-gradient(135deg, #6B7280 0%, #4B5563 100%)'
                  : 'linear-gradient(135deg, #F0B90B 0%, #d9a309 100%)',
                color: '#000000',
              }}
            >
              {loading ? (
                <>
                  <RefreshCw className="h-4 w-4 animate-spin" />
                  {language === 'zh' ? '创建中...' : 'Creating...'}
                </>
              ) : (
                <>
                  <ArrowRight className="h-4 w-4" />
                  {language === 'zh' ? '创建 Agent 钱包' : 'Create Agent Wallet'}
                </>
              )}
            </button>
          </div>
        )}

        {/* Error/Status Messages */}
        {error && (
          <div
            className="rounded-xl p-4 mb-6"
            style={{ background: 'rgba(246, 70, 93, 0.1)', border: '1px solid rgba(246, 70, 93, 0.2)' }}
          >
            <div className="flex items-center gap-2" style={{ color: '#F6465D' }}>
              <AlertCircle className="h-4 w-4" />
              <span className="text-sm">{error}</span>
            </div>
          </div>
        )}

        {authStatus && (
          <div
            className="rounded-xl p-4 mb-6"
            style={{ background: 'rgba(14, 203, 129, 0.1)', border: '1px solid rgba(14, 203, 129, 0.2)' }}
          >
            <div className="flex items-center gap-2" style={{ color: '#0ECB81' }}>
              <Check className="h-4 w-4" />
              <span className="text-sm whitespace-pre-line">{authStatus}</span>
            </div>
          </div>
        )}

        {/* Technical Details */}
        <details
          className="rounded-xl p-6"
          style={{ background: '#0a0a0a', border: '1px solid #2b3139' }}
        >
          <summary
            className="cursor-pointer font-semibold transition-colors"
            style={{ color: '#EAECEF' }}
            onMouseEnter={(e) => (e.currentTarget.style.color = '#F0B90B')}
            onMouseLeave={(e) => (e.currentTarget.style.color = '#EAECEF')}
          >
            {language === 'zh' ? '技术详情' : 'Technical Details'}
          </summary>
          <div className="mt-4 space-y-3 text-sm" style={{ color: '#848E9C' }}>
            <p>
              <strong style={{ color: '#EAECEF' }}>
                {language === 'zh' ? '后端生成流程：' : 'Backend Generation Process:'}
              </strong>
            </p>
            <ol className="list-decimal list-inside space-y-2 ml-4">
              <li>
                {language === 'zh'
                  ? '后端使用 Go crypto/ecdsa 生成私钥'
                  : 'Backend generates private key using Go crypto/ecdsa'}
              </li>
              <li>
                {language === 'zh'
                  ? '使用 AES-256-GCM 加密私钥'
                  : 'Encrypts private key with AES-256-GCM'}
              </li>
              <li>
                {language === 'zh'
                  ? '加密后的私钥存储在数据库'
                  : 'Stores encrypted private key in database'}
              </li>
              <li>
                {language === 'zh'
                  ? '返回 Agent 地址给前端'
                  : 'Returns Agent address to frontend'}
              </li>
            </ol>

            <p className="mt-4">
              <strong style={{ color: '#EAECEF' }}>
                {language === 'zh' ? '安全优势：' : 'Security Advantages:'}
              </strong>
            </p>
            <ul className="list-disc list-inside space-y-1 ml-4">
              <li>{language === 'zh' ? '用户无需保存私钥' : 'No private key management needed'}</li>
              <li>
                {language === 'zh' ? '私钥不暴露在前端' : 'Private key never exposed to frontend'}
              </li>
              <li>
                {language === 'zh' ? '适合托管服务场景' : 'Suitable for custodial service scenarios'}
              </li>
              <li>
                {language === 'zh' ? '降低用户使用门槛' : 'Lower barrier for users'}
              </li>
            </ul>
          </div>
        </details>
      </div>
    </div>
  )
}
