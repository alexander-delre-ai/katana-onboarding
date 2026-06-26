import { useState, useEffect } from 'react'

interface KomatsuUser {
  username: string
  name: string
  email: string
  has_aws_account: boolean
}

interface UsersResponse {
  success: boolean
  configured: boolean
  source: string
  users: KomatsuUser[]
}

function OnboardedKomatsu() {
  const [users, setUsers] = useState<KomatsuUser[]>([])
  const [configured, setConfigured] = useState(true)
  const [source, setSource] = useState('')
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    const load = async () => {
      setLoading(true)
      try {
        const r = await fetch('/api/komatsu-users')
        const data = (await r.json()) as UsersResponse
        setUsers(data.users || [])
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
        <h2 className="text-2xl font-bold text-gray-800">Onboarded - Komatsu</h2>
        <p className="text-gray-600 mt-1">ext-komatsu engineers and their AWS account status</p>
      </div>

      <div className="bg-white rounded-lg shadow p-6">
        <h3 className="font-medium text-gray-800 mb-4">
          Engineers {users.length > 0 && <span className="text-gray-500">({users.length})</span>}
        </h3>

        {loading ? (
          <div className="text-center text-gray-500 py-6">Loading...</div>
        ) : !configured ? (
          <div className="p-3 rounded-lg bg-amber-50 border border-amber-200 text-amber-800 text-sm">
            Data source not connected yet{source ? `: ${source}` : ''}. This view will populate once
            the backend is wired up.
          </div>
        ) : users.length === 0 ? (
          <div className="text-center text-gray-500 py-6">No engineers found.</div>
        ) : (
          <div className="overflow-x-auto border rounded-lg">
            <table className="w-full text-sm">
              <thead className="bg-gray-50 text-left text-gray-500">
                <tr>
                  <th className="px-4 py-2 font-medium">Engineer</th>
                  <th className="px-4 py-2 font-medium">Email</th>
                  <th className="px-4 py-2 font-medium">AWS account</th>
                </tr>
              </thead>
              <tbody>
                {users.map((u) => (
                  <tr key={u.username} className="border-t">
                    <td className="px-4 py-2">{u.name || u.username}</td>
                    <td className="px-4 py-2 text-gray-600">{u.email || '-'}</td>
                    <td className="px-4 py-2">
                      {u.has_aws_account ? (
                        <span className="inline-flex items-center px-2 py-0.5 rounded-full text-xs font-medium bg-green-100 text-green-800">
                          Yes
                        </span>
                      ) : (
                        <span className="inline-flex items-center px-2 py-0.5 rounded-full text-xs font-medium bg-gray-100 text-gray-600">
                          No
                        </span>
                      )}
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

export default OnboardedKomatsu
