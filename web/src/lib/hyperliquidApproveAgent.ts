/**
 * Hyperliquid ApproveAgent Authorization
 *
 * This module handles the ApproveAgent action to authorize an agent wallet
 * to trade on behalf of the main wallet.
 */

import { type WalletClient } from 'viem'

// EIP-712 Domain for Hyperliquid (matches official documentation)
const HYPERLIQUID_DOMAIN = {
  name: 'HyperliquidSignTransaction',
  version: '1',
  chainId: 421614, // Hyperliquid L1 (0x66eee) - used for user-signed actions
  verifyingContract: '0x0000000000000000000000000000000000000000' as `0x${string}`,
}

// EIP-712 Types for ApproveAgent (matches Hyperliquid API - order matters!)
const APPROVE_AGENT_TYPES = {
  'HyperliquidTransaction:ApproveAgent': [
    { name: 'hyperliquidChain', type: 'string' },
    { name: 'agentAddress', type: 'string' },
    { name: 'agentName', type: 'string' },
    { name: 'nonce', type: 'uint64' },
  ],
} as const

interface ApproveAgentParams {
  agentAddress: string
  agentName?: string
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
 * Generate signature for ApproveAgent action
 * Returns both signature object, hex signature, and nonce to ensure they match
 */
export async function signApproveAgent(
  walletClient: WalletClient,
  params: ApproveAgentParams
): Promise<{
  signature: { r: string; s: string; v: number }
  signatureHex: string
  nonce: number
}> {
  if (!walletClient.account) {
    throw new Error('Wallet not connected')
  }

  const { agentAddress, agentName = '', hyperliquidChain = 'Mainnet' } = params

  const nonce = Date.now()

  const message = {
    hyperliquidChain,
    agentAddress: agentAddress.toLowerCase(),
    agentName,
    nonce: BigInt(nonce),
  }

  // Sign using EIP-712 (matches Hyperliquid API)
  const signatureHex = await walletClient.signTypedData({
    account: walletClient.account,
    domain: HYPERLIQUID_DOMAIN,
    types: APPROVE_AGENT_TYPES,
    primaryType: 'HyperliquidTransaction:ApproveAgent',
    message,
  })

  // Convert to {r, s, v} format (required by Hyperliquid API)
  const signature = signatureToRSV(signatureHex)

  return { signature, signatureHex, nonce }
}
