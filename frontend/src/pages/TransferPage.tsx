import { useState } from 'react'
import { apiFetch, HttpError } from '../api/client'
import { useAuth } from '../auth/AuthContext'
import type { Currency, Transaction } from '../api/types'
import { decimalToCents } from '../lib/money'

export function TransferPage() {
  const { token } = useAuth()
  const [toEmail, setToEmail] = useState('')
  const [currency, setCurrency] = useState<Currency>('USD')
  const [amount, setAmount] = useState('10.00')
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [success, setSuccess] = useState<Transaction | null>(null)

  return (
    <div className="max-w-xl space-y-4">
      <div>
        <h2 className="text-xl font-semibold">Transfer</h2>
        <p className="text-sm text-slate-600">Send money to another user in the same currency.</p>
      </div>

      <form
        className="space-y-3 rounded-lg border bg-white p-4"
        onSubmit={async (e) => {
          e.preventDefault()
          setError(null)
          setSuccess(null)
          setLoading(true)
          try {
            const amountCents = decimalToCents(amount)
            const tx = await apiFetch<Transaction>('/transactions/transfer', {
              method: 'POST',
              token,
              body: { to_user_email: toEmail.trim(), currency, amount_cents: amountCents },
            })
            setSuccess(tx)
          } catch (e: any) {
            setError(e instanceof HttpError ? e.message : 'transfer_failed')
          } finally {
            setLoading(false)
          }
        }}
      >
        <label className="block">
          <div className="text-xs font-medium text-slate-600">Recipient email</div>
          <input className="mt-1 w-full rounded border px-3 py-2" value={toEmail} onChange={(e) => setToEmail(e.target.value)} />
        </label>

        <div className="grid grid-cols-2 gap-3">
          <label className="block">
            <div className="text-xs font-medium text-slate-600">Currency</div>
            <select className="mt-1 w-full rounded border px-3 py-2" value={currency} onChange={(e) => setCurrency(e.target.value as Currency)}>
              <option value="USD">USD</option>
              <option value="EUR">EUR</option>
            </select>
          </label>
          <label className="block">
            <div className="text-xs font-medium text-slate-600">Amount</div>
            <input className="mt-1 w-full rounded border px-3 py-2" value={amount} onChange={(e) => setAmount(e.target.value)} placeholder="10.00" />
          </label>
        </div>

        {error ? <div className="rounded bg-red-50 p-2 text-sm text-red-700">{error}</div> : null}
        {success ? <div className="rounded bg-green-50 p-2 text-sm text-green-700">Transfer created: {success.id}</div> : null}

        <button className="rounded bg-slate-900 px-3 py-2 text-sm font-medium text-white disabled:opacity-50" disabled={loading}>
          {loading ? 'Sending…' : 'Send'}
        </button>
      </form>
    </div>
  )
}

