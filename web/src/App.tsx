import { RouterProvider } from 'react-router-dom'
import { WagmiProvider } from 'wagmi'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { RainbowKitProvider, darkTheme } from '@rainbow-me/rainbowkit'
import '@rainbow-me/rainbowkit/styles.css'
import { LanguageProvider } from './contexts/LanguageContext'
import { AuthProvider } from './contexts/AuthContext'
import { ConfirmDialogProvider } from './components/ConfirmDialog'
import { router } from './routes'
import { config } from './config/wagmi'

const queryClient = new QueryClient()

function AppContent() {
  // Don't show loading screen, let pages handle their own loading states
  // The global loading was causing unnecessary delays on landing pages

  return <RouterProvider router={router} />
}

export default function App() {
  return (
    <WagmiProvider config={config}>
      <QueryClientProvider client={queryClient}>
        <RainbowKitProvider
          theme={darkTheme({
            accentColor: '#F0B90B', // NOFX 品牌黄色
            accentColorForeground: '#0B0E11', // 深色文字
            borderRadius: 'medium',
            overlayBlur: 'small',
          })}
          modalSize="compact"
          appInfo={{
            appName: 'NOFX',
          }}
        >
          <LanguageProvider>
            <AuthProvider>
              <ConfirmDialogProvider>
                <AppContent />
              </ConfirmDialogProvider>
            </AuthProvider>
          </LanguageProvider>
        </RainbowKitProvider>
      </QueryClientProvider>
    </WagmiProvider>
  )
}
