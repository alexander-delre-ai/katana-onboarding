import { useState, useEffect } from 'react'

interface PendingRequest {
  engineer: string
  email: string
  ticket_key: string
  ticket_url: string
  status: string
  created: string
}

interface PendingResponse {
  success: boolean
  configured: boolean
  source: string
  requests: PendingRequest[]
}

function PendingKomatsu() {
  const [pending, setPending] = useState<PendingRequest[]>([])
  const [configured, setConfigured] = useState(true)
  const [source, setSource] = useState('')
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    const load = async () => {
      setLoading(true)
      try {
        const r = await fetch('/api/pending-access')
        const data = (await r.json()) as PendingResponse
        setPending(data.requests || [])
        setConfigured(data.configured)
        setSource(data.source)
      } catch {
        // Leave defaults; the empty state will render.
      } finally {
        setLoading(false)
      }
    }
    load()
  }, [])

  return (
    <div>
      <div className="mb-6">
        <h2 className="text-2xl font-bold text-gray-800">Pending Onboarding - Komatsu</h2>
        <p className="text-gray-600 mt-1">
          Engineers with open onboarding (terminal) tickets awaiting IT
        </p>
      </div>

      <div className="bg-white rounded-lg shadow p-6">
        <h3 className="font-medium text-gray-800 mb-4">
          Pending access{' '}
          {pending.length > 0 && <span className="text-gray-500">({pending.length})</span>}
        </h3>

        {loading ? (
          <div className="text-center text-gray-500 py-6">Loading...</div>
        ) : !configured ? (
          <div className="p-3 rounded-lg bg-amber-50 border border-amber-200 text-amber-800 text-sm">
            Data source not connected yet{source ? `: ${source}` : ''}. This view will populate once
            the backend is wired up.
          </div>
        ) : pending.length === 0 ? (
          <div className="text-center text-gray-500 py-6">No pending access requests.</div>
        ) : (
          <div className="overflow-x-auto border rounded-lg">
            <table className="w-full text-sm">
              <thead className="bg-gray-50 text-left text-gray-500">
                <tr>
                  <th className="px-4 py-2 font-medium">Engineer</th>
                  <th className="px-4 py-2 font-medium">Email</th>
                  <th className="px-4 py-2 font-medium">Status</th>
                  <th className="px-4 py-2 font-medium">Ticket</th>
                </tr>
              </thead>
              <tbody>
                {pending.map((r) => (
                  <tr key={r.ticket_key} className="border-t">
                    <td className="px-4 py-2">{r.engineer || '-'}</td>
                    <td className="px-4 py-2 text-gray-600">{r.email || '-'}</td>
                    <td className="px-4 py-2">{r.status || '-'}</td>
                    <td className="px-4 py-2">
                      <a
                        href={r.ticket_url}
                        target="_blank"
                        rel="noreferrer"
                        className="text-blue-600 hover:underline"
                      >
                        {r.ticket_key || 'View ticket'}
                      </a>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>
    </div>
  )
}

export default PendingKomatsu
