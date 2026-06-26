import { useState, useEffect, createContext, useContext } from 'react'
import Sidebar from './components/Sidebar'
import OnboardedKomatsu from './components/OnboardedKomatsu'
import PendingKomatsu from './components/PendingKomatsu'

export type Tab = 'onboarded' | 'pending'

interface BotStatus {
  ready: boolean
  message: string
}

const StatusContext = createContext<BotStatus>({ ready: false, message: 'Loading...' })

export function useStatus() {
  return useContext(StatusContext)
}

function App() {
  const [activeTab, setActiveTab] = useState<Tab>('onboarded')
  const [status, setStatus] = useState<BotStatus>({ ready: false, message: 'Loading...' })

  useEffect(() => {
    if (status.ready) return // Stop polling once initialized

    const fetchStatus = async () => {
      try {
        const response = await fetch('/slack/status')
        const data = await response.json()
        setStatus({ ready: data.ready, message: data.message })
      } catch {
        setStatus({ ready: false, message: 'Failed to connect to server' })
      }
    }

    fetchStatus()
    const interval = setInterval(fetchStatus, 5000) // Poll every 5s until ready
    return () => clearInterval(interval)
  }, [status.ready])

  const renderContent = () => {
    switch (activeTab) {
      case 'onboarded':
        return <OnboardedKomatsu />
      case 'pending':
        return <PendingKomatsu />
      default:
        return <OnboardedKomatsu />
    }
  }

  return (
    <StatusContext.Provider value={status}>
      <div className="flex h-screen bg-gray-100">
        <Sidebar activeTab={activeTab} setActiveTab={setActiveTab} />
        <main className="flex-1 overflow-auto p-6">
          {renderContent()}
        </main>
      </div>
    </StatusContext.Provider>
  )
}

export default App
