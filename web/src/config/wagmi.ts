import { connectorsForWallets } from '@rainbow-me/rainbowkit'
import {
  metaMaskWallet,
  walletConnectWallet,
  coinbaseWallet,
  okxWallet,
  phantomWallet,
  binanceWallet,
  bybitWallet,
  trustWallet,
} from '@rainbow-me/rainbowkit/wallets'
import { createConfig, http } from 'wagmi'
import { mainnet, arbitrum, arbitrumSepolia, base, optimism, polygon } from 'wagmi/chains'

// WalletConnect 项目 ID - 需要从 https://cloud.walletconnect.com/ 获取
// 使用一个临时的默认 ID，但强烈建议替换为自己的项目 ID
// 如果没有有效 ID，RainbowKit 会回退到仅支持注入式钱包（如 MetaMask）
const projectId =
  import.meta.env.VITE_WALLETCONNECT_PROJECT_ID ||
  'adb9696ee8fbfbf7f434696cf1717699'

// 自定义钱包列表
const connectors = connectorsForWallets(
  [
    {
      groupName: 'Recommended',
      wallets: [
        metaMaskWallet,
        walletConnectWallet,
        coinbaseWallet,
      ],
    },
    {
      groupName: 'Popular',
      wallets: [
        okxWallet, // OKX 钱包
        binanceWallet, // Binance 币安钱包
        bybitWallet, // Bybit 钱包
        phantomWallet, // Phantom (幻影钱包)
        trustWallet,
      ],
    },
  ],
  {
    appName: 'NOFX Trading OS',
    projectId,
  }
)

// 支持的链
const chains = [mainnet, arbitrum, arbitrumSepolia, base, optimism, polygon] as const

export const config = createConfig({
  connectors,
  chains,
  transports: {
    [mainnet.id]: http(),
    [arbitrum.id]: http(),
    [arbitrumSepolia.id]: http(),
    [base.id]: http(),
    [optimism.id]: http(),
    [polygon.id]: http(),
  },
  ssr: false,
})
