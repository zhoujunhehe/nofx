/**
 * Hyperliquid Builder Fee Authorization
 *
 * This module handles the ApproveBuilderFee action to authorize a builder
 * to charge a fee on user's trades (up to 0.1% for perps, 1% for spot).
 */

import { type WalletClient } from 'viem'

// EIP-712 Domain for Hyperliquid (matches official documentation)
const HYPERLIQUID_DOMAIN = {
  name: 'HyperliquidSignTransaction',
  version: '1',
  chainId: 421614, // Hyperliquid L1 (0x66eee) - used for user-signed actions
  verifyingContract: '0x0000000000000000000000000000000000000000' as `0x${string}`,
}

// EIP-712 Types for ApproveBuilderFee (matches Python SDK - order matters!)
const APPROVE_BUILDER_FEE_TYPES = {
  'HyperliquidTransaction:ApproveBuilderFee': [
    { name: 'hyperliquidChain', type: 'string' },
    { name: 'maxFeeRate', type: 'string' },
    { name: 'builder', type: 'address' },
    { name: 'nonce', type: 'uint64' },
  ],
} as const

interface ApproveBuilderFeeParams {
  builderAddress: string
  maxFeeRate?: number // in tenths of basis points (10 = 1bp = 0.01%, 100 = 0.1%)
  hyperliquidChain?: 'Mainnet' | 'Testnet'
}

/**
 * Convert hex signature to {r, s, v} format (Hyperliquid API expects this format)
 */
function signatureToRSV(signature: string): { r: string; s: string; v: number } {
  // Remove 0x prefix if present
  const sig = signature.startsWith('0x') ? signature.slice(2) : signature

  // Extract r, s, v from signature (each 32 bytes = 64 hex chars, v is 1 byte = 2 hex chars)
  const r = '0x' + sig.slice(0, 64)
  const s = '0x' + sig.slice(64, 128)
  const v = parseInt(sig.slice(128, 130), 16)

  return { r, s, v }
}

/**
 * Generate signature for ApproveBuilderFee action
 * Returns both signature object and nonce to ensure they match
 */
async function signApproveBuilderFee(
  walletClient: WalletClient,
  params: ApproveBuilderFeeParams
): Promise<{ signature: { r: string; s: string; v: number }; nonce: number }> {
  if (!walletClient.account) {
    throw new Error('Wallet not connected')
  }

  const { builderAddress, maxFeeRate = 10, hyperliquidChain = 'Mainnet' } = params

  const nonce = Date.now()

  // Convert maxFeeRate from tenths of basis points to percentage string
  // 100 (tenths of bp) = 10 bp = 0.1% = "0.1%"
  // 10 = 1 bp = 0.01% = "0.01%"
  // 1000 = 100 bp = 1% = "1%"
  const maxFeeRatePercentage = (maxFeeRate / 1000).toString() + '%'

  const message = {
    hyperliquidChain,
    maxFeeRate: maxFeeRatePercentage,
    builder: builderAddress as `0x${string}`,
    nonce: BigInt(nonce),
  }

  // Sign using EIP-712 (matches Python SDK)
  const signatureHex = await walletClient.signTypedData({
    account: walletClient.account,
    domain: HYPERLIQUID_DOMAIN,
    types: APPROVE_BUILDER_FEE_TYPES,
    primaryType: 'HyperliquidTransaction:ApproveBuilderFee',
    message,
  })

  // Convert to {r, s, v} format (required by Hyperliquid API)
  const signature = signatureToRSV(signatureHex)

  return { signature, nonce }
}

/**
 * Approve builder fee on Hyperliquid
 *
 * @param maxFeeRate - Maximum fee rate in tenths of basis points
 *                     10 = 1 basis point = 0.01%
 *                     100 = 10 basis points = 0.1% (recommended for perps)
 *                     1000 = 100 basis points = 1% (max for spot)
 */
export async function approveHyperliquidBuilderFee(
  walletClient: WalletClient,
  params: ApproveBuilderFeeParams
): Promise<{ success: boolean; message: string }> {
  try {
    const { builderAddress, maxFeeRate = 10, hyperliquidChain = 'Mainnet' } = params

    if (!walletClient.account) {
      throw new Error('钱包未连接')
    }

    // Validate fee rate
    const isSpot = false // Currently only for perps
    const maxAllowed = isSpot ? 1000 : 100 // 1% for spot, 0.1% for perps
    if (maxFeeRate > maxAllowed) {
      throw new Error(
        `最大费率超出限制。永续合约最高 ${maxAllowed / 10} 个基点 (${maxAllowed / 1000}%)`
      )
    }

    // Step 1: Generate signature (returns both signature and nonce to ensure they match)
    const { signature, nonce } = await signApproveBuilderFee(walletClient, params)

    // Convert maxFeeRate to percentage format for action payload
    const maxFeeRatePercentage = (maxFeeRate / 1000).toString() + '%'

    // Step 2: Build action payload (use the SAME nonce from signature)
    const action = {
      type: 'approveBuilderFee',
      signatureChainId: '0x66eee', // Must match the value in message (Hyperliquid L1 chainId)
      hyperliquidChain,
      builder: builderAddress,
      maxFeeRate: maxFeeRatePercentage,
      nonce,
    }

    // Step 3: Send to Hyperliquid API
    const apiUrl =
      hyperliquidChain === 'Mainnet'
        ? 'https://api.hyperliquid.xyz/exchange'
        : 'https://api.hyperliquid-testnet.xyz/exchange'

    const requestBody = {
      action,
      nonce,  // Outer nonce (same as action.nonce, required by API)
      signature,
    }

    console.log('🔧 ApproveBuilderFee Request Body:', requestBody)

    const response = await fetch(apiUrl, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(requestBody),
    })

    if (!response.ok) {
      const errorData = await response.json().catch(() => ({}))
      console.error('Hyperliquid Builder Fee API Error:', {
        status: response.status,
        statusText: response.statusText,
        errorData,
        requestBody: { action, signature }  // Only log what we actually sent
      })
      throw new Error(errorData.error || `HTTP ${response.status}: ${response.statusText}`)
    }

    const result = await response.json()

    console.log('✅ ApproveBuilderFee Response:', result)

    // Check if authorization was successful
    if (result.status === 'ok' || result.type === 'default') {
      const feePercentage = (maxFeeRate / 1000).toFixed(2)
      return {
        success: true,
        message: `授权成功！Builder 可收取最高 ${feePercentage}% 的交易手续费`,
      }
    } else if (result.status === 'err') {
      // Handle specific error cases
      if (result.response && result.response.includes('Must deposit')) {
        throw new Error(
          `Builder 地址需要在 Hyperliquid 存款才能接收手续费。\n` +
          `提示：Builder Fee 是可选的，不影响 Agent 交易功能。\n` +
          `如不需要 Builder Fee，可跳过此步骤。`
        )
      }
      console.error('❌ Unexpected response format:', result)
      throw new Error(result.response || result.error || '授权失败，请稍后重试')
    } else {
      console.error('❌ Unexpected response format:', result)
      throw new Error(result.error || '授权失败，请稍后重试')
    }
  } catch (error: any) {
    console.error('Hyperliquid builder fee approval failed:', error)
    return {
      success: false,
      message: error.message || '授权失败，请检查网络连接',
    }
  }
}

/**
 * Revoke builder fee authorization (set maxFeeRate to 0)
 */
export async function revokeHyperliquidBuilderFee(
  walletClient: WalletClient,
  builderAddress: string,
  hyperliquidChain: 'Mainnet' | 'Testnet' = 'Mainnet'
): Promise<{ success: boolean; message: string }> {
  try {
    if (!walletClient.account) {
      throw new Error('钱包未连接')
    }

    // Revoke by setting maxFeeRate to 0
    const result = await approveHyperliquidBuilderFee(walletClient, {
      builderAddress,
      maxFeeRate: 0, // 0 = revoke authorization
      hyperliquidChain,
    })

    if (result.success) {
      return {
        success: true,
        message: '撤销成功！Builder Fee 授权已移除',
      }
    }

    return result
  } catch (error: any) {
    console.error('Revoke builder fee failed:', error)
    return {
      success: false,
      message: error.message || '撤销失败，请检查网络连接',
    }
  }
}

/**
 * Query user's USDC balance on Hyperliquid
 */
export async function queryHyperliquidBalance(
  userAddress: string,
  hyperliquidChain: 'Mainnet' | 'Testnet' = 'Mainnet'
): Promise<{ balance: number; error?: string }> {
  try {
    const apiUrl =
      hyperliquidChain === 'Mainnet'
        ? 'https://api.hyperliquid.xyz/info'
        : 'https://api.hyperliquid-testnet.xyz/info'

    const response = await fetch(apiUrl, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        type: 'clearinghouseState',
        user: userAddress,
      }),
    })

    if (!response.ok) {
      throw new Error(`HTTP ${response.status}`)
    }

    const result = await response.json()

    // Extract USDC balance from marginSummary
    if (result && result.marginSummary && result.marginSummary.accountValue) {
      const balance = parseFloat(result.marginSummary.accountValue)
      return { balance }
    }

    return { balance: 0 }
  } catch (error) {
    console.error('Query Hyperliquid balance failed:', error)
    return { balance: 0, error: error instanceof Error ? error.message : 'Unknown error' }
  }
}

/**
 * Query user's account status on Hyperliquid
 */
export async function queryAccountStatus(
  userAddress: string,
  hyperliquidChain: 'Mainnet' | 'Testnet' = 'Mainnet'
): Promise<{
  hasDeposit: boolean
  balance: number
  accountValue: number
  error?: string
}> {
  try {
    const apiUrl =
      hyperliquidChain === 'Mainnet'
        ? 'https://api.hyperliquid.xyz/info'
        : 'https://api.hyperliquid-testnet.xyz/info'

    const response = await fetch(apiUrl, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        type: 'clearinghouseState',
        user: userAddress,
      }),
    })

    if (!response.ok) {
      throw new Error(`HTTP ${response.status}`)
    }

    const result = await response.json()

    // Extract account information
    if (result && result.marginSummary) {
      const accountValue = parseFloat(result.marginSummary.accountValue || '0')
      return {
        hasDeposit: accountValue > 0,
        balance: accountValue,
        accountValue,
      }
    }

    return {
      hasDeposit: false,
      balance: 0,
      accountValue: 0,
    }
  } catch (error) {
    console.error('Query account status failed:', error)
    return {
      hasDeposit: false,
      balance: 0,
      accountValue: 0,
      error: error instanceof Error ? error.message : 'Unknown error',
    }
  }
}

/**
 * Query approved builder fee for a user
 */
export async function queryBuilderFee(
  userAddress: string,
  builderAddress: string,
  hyperliquidChain: 'Mainnet' | 'Testnet' = 'Mainnet'
): Promise<{ approved: boolean; maxFeeRate: number }> {
  try {
    const apiUrl =
      hyperliquidChain === 'Mainnet'
        ? 'https://api.hyperliquid.xyz/info'
        : 'https://api.hyperliquid-testnet.xyz/info'

    const response = await fetch(apiUrl, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        type: 'maxBuilderFee',
        user: userAddress,
        builder: builderAddress,
      }),
    })

    if (!response.ok) {
      throw new Error(`HTTP ${response.status}`)
    }

    const result = await response.json()

    // If result is a number, that's the max fee rate
    if (typeof result === 'string' || typeof result === 'number') {
      const feeRate = parseInt(result.toString())
      return {
        approved: feeRate > 0,
        maxFeeRate: feeRate,
      }
    }

    return {
      approved: false,
      maxFeeRate: 0,
    }
  } catch (error) {
    console.error('Query builder fee failed:', error)
    return {
      approved: false,
      maxFeeRate: 0,
    }
  }
}
