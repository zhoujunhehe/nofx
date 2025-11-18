import { http, createConfig } from 'wagmi'
import { mainnet, arbitrum } from 'wagmi/chains'
import { injected, walletConnect } from 'wagmi/connectors'

// Wagmi configuration for Web3 wallet connections
export const config = createConfig({
  chains: [mainnet, arbitrum],
  connectors: [
    injected(),
    walletConnect({
      projectId:
        import.meta.env.VITE_WALLETCONNECT_PROJECT_ID || 'default-project-id',
    }),
  ],
  transports: {
    [mainnet.id]: http(),
    [arbitrum.id]: http(),
  },
})
