/**
 * Agent Wallet Backend API
 * 后端生成和托管 Agent 钱包
 */

const API_BASE_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080'

// Helper function to get auth headers
function getAuthHeaders(): Record<string, string> {
  const token = localStorage.getItem('auth_token')
  const headers: Record<string, string> = {
    'Content-Type': 'application/json',
  }

  if (token) {
    headers['Authorization'] = `Bearer ${token}`
  }

  return headers
}

export interface AgentWallet {
  id: number
  main_wallet: string
  agent_address: string
  authorization_signature?: string
  status: 'INIT' | 'ACTIVE' | 'REVOKED'
  hyperliquid_chain: 'Mainnet' | 'Testnet'
  builder_fee_authorized: boolean
  builder_fee_max_rate: number
  builder_fee_authorized_at?: string
  created_at: string
  updated_at: string
}

export interface CreateAgentWalletRequest {
  main_wallet: string
  hyperliquid_chain?: 'Mainnet' | 'Testnet'
}

export interface CreateAgentWalletResponse {
  success: boolean
  message: string
  agent_address?: string
  main_wallet?: string
  status?: string
}

export interface GetAgentWalletResponse {
  success: boolean
  message?: string
  data?: AgentWallet
}

export interface AuthorizeAgentRequest {
  main_wallet: string
  signature: string
  agent_name?: string
  nonce: number
  signature_rsv: {
    r: string
    s: string
    v: number
  }
}

export interface AuthorizeAgentResponse {
  success: boolean
  message: string
  status?: string
}

/**
 * 创建 Agent 钱包（后端生成私钥）
 */
export async function createAgentWallet(
  mainWallet: string,
  hyperliquidChain: 'Mainnet' | 'Testnet' = 'Mainnet'
): Promise<CreateAgentWalletResponse> {
  const response = await fetch(`${API_BASE_URL}/api/agent/create`, {
    method: 'POST',
    headers: getAuthHeaders(),
    body: JSON.stringify({
      main_wallet: mainWallet.toLowerCase(),
      hyperliquid_chain: hyperliquidChain,
    } as CreateAgentWalletRequest),
  })

  if (!response.ok) {
    const errorData = await response.json().catch(() => ({}))
    throw new Error(errorData.message || `HTTP ${response.status}`)
  }

  return await response.json()
}

/**
 * 查询 Agent 钱包状态
 */
export async function getAgentWallet(
  mainWallet: string
): Promise<GetAgentWalletResponse> {
  const response = await fetch(
    `${API_BASE_URL}/api/agent/status?main_wallet=${mainWallet.toLowerCase()}`,
    {
      method: 'GET',
      headers: getAuthHeaders(),
    }
  )

  if (!response.ok) {
    const errorData = await response.json().catch(() => ({}))
    throw new Error(errorData.message || `HTTP ${response.status}`)
  }

  return await response.json()
}

/**
 * 授权 Agent 钱包（提交 MainWallet 签名到后端，后端转发到 Hyperliquid）
 */
export async function authorizeAgent(
  request: AuthorizeAgentRequest
): Promise<AuthorizeAgentResponse> {
  const response = await fetch(`${API_BASE_URL}/api/agent/authorize`, {
    method: 'POST',
    headers: getAuthHeaders(),
    body: JSON.stringify(request),
  })

  if (!response.ok) {
    const errorData = await response.json().catch(() => ({}))
    throw new Error(errorData.message || `HTTP ${response.status}`)
  }

  return await response.json()
}

export interface ConfirmBuilderFeeRequest {
  main_wallet: string
  max_fee_rate: number
}

export interface ConfirmBuilderFeeResponse {
  success: boolean
  message: string
}

/**
 * 确认 Builder Fee 授权（前端在成功授权后调用，通知后端）
 */
export async function confirmBuilderFee(
  mainWallet: string,
  maxFeeRate: number
): Promise<ConfirmBuilderFeeResponse> {
  const response = await fetch(
    `${API_BASE_URL}/api/agent/confirm-builder-fee`,
    {
      method: 'POST',
      headers: getAuthHeaders(),
      body: JSON.stringify({
        main_wallet: mainWallet.toLowerCase(),
        max_fee_rate: maxFeeRate,
      } as ConfirmBuilderFeeRequest),
    }
  )

  if (!response.ok) {
    const errorData = await response.json().catch(() => ({}))
    throw new Error(errorData.message || `HTTP ${response.status}`)
  }

  return await response.json()
}
