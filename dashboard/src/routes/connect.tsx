import { createFileRoute, redirect } from '@tanstack/react-router'
import { ConnectView } from '@/views/connect'

export const Route = createFileRoute('/connect')({
  beforeLoad({ context }) {
    if (!context.clusterStatus) return

    if (context.clusterStatus.connectionStatus === 'connected') {
      if (import.meta.env.PROD) {
        throw redirect({ to: '/' })
      }
    }
  },
  component: ConnectView
})
