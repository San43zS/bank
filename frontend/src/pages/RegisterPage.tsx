import { useState } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { HttpError } from '../api/client'
import { useAuth } from '../auth/AuthContext'

export function RegisterPage() {
  const { register } = useAuth()
  const nav = useNavigate()
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [firstName, setFirstName] = useState('')
  const [lastName, setLastName] = useState('')
  const [error, setError] = useState<string | null>(null)
  const [loading, setLoading] = useState(false)

  return (
    <div className="mx-auto flex min-h-full max-w-md flex-col justify-center px-4 py-10">
      <h1 className="text-2xl font-semibold">Register</h1>
      <p className="mt-1 text-sm text-slate-600">Creates USD and EUR accounts automatically.</p>

      <form
        className="mt-6 space-y-3 rounded-lg border bg-white p-4"
        onSubmit={async (e) => {
          e.preventDefault()
          setError(null)
          setLoading(true)
          try {
            await register({ email, password, firstName, lastName })
            nav('/')
          } catch (e: any) {
            setError(e instanceof HttpError ? e.message : 'register_failed')
          } finally {
            setLoading(false)
          }
        }}
      >
        <div className="grid grid-cols-2 gap-3">
          <label className="block">
            <div className="text-xs font-medium text-slate-600">First name</div>
            <input className="mt-1 w-full rounded border px-3 py-2" value={firstName} onChange={(e) => setFirstName(e.target.value)} />
          </label>
          <label className="block">
            <div className="text-xs font-medium text-slate-600">Last name</div>
            <input className="mt-1 w-full rounded border px-3 py-2" value={lastName} onChange={(e) => setLastName(e.target.value)} />
          </label>
        </div>

        <label className="block">
          <div className="text-xs font-medium text-slate-600">Email</div>
          <input className="mt-1 w-full rounded border px-3 py-2" value={email} onChange={(e) => setEmail(e.target.value)} autoComplete="email" />
        </label>
        <label className="block">
          <div className="text-xs font-medium text-slate-600">Password</div>
          <input
            className="mt-1 w-full rounded border px-3 py-2"
            type="password"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            autoComplete="new-password"
          />
        </label>

        {error ? <div className="rounded bg-red-50 p-2 text-sm text-red-700">{error}</div> : null}

        <button
          className="w-full rounded bg-slate-900 px-3 py-2 text-sm font-medium text-white disabled:opacity-50"
          disabled={loading}
        >
          {loading ? 'Creating…' : 'Create account'}
        </button>
      </form>

      <div className="mt-3 text-sm text-slate-600">
        Already have an account?{' '}
        <Link className="font-medium text-slate-900 underline" to="/login">
          Login
        </Link>
      </div>
    </div>
  )
}

