import { useMemo, useState } from 'react'
import { apiFetch, HttpError } from '../api/client'
import { useAuth } from '../auth/AuthContext'
import type { Currency, Transaction } from '../api/types'
import { centsToDecimal, decimalToCents } from '../lib/money'

const USD_TO_EUR = 0.92

export function ExchangePage() {
  const { token } = useAuth()
  const [fromCurrency, setFromCurrency] = useState<Currency>('USD')
  const [amount, setAmount] = useState('10.00')
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [success, setSuccess] = useState<Transaction | null>(null)

  const toCurrency: Currency = fromCurrency === 'USD' ? 'EUR' : 'USD'

  const preview = useMemo(() => {
    try {
      const cents = decimalToCents(amount)
      const converted =
        fromCurrency === 'USD' ? Math.round(cents * USD_TO_EUR) : Math.round(cents / USD_TO_EUR)
      return { cents, converted, rate: fromCurrency === 'USD' ? USD_TO_EUR : 1 / USD_TO_EUR }
    } catch {
      return null
    }
  }, [amount, fromCurrency])

  return (
    <div className="max-w-xl space-y-4">
      <div>
        <h2 className="text-xl font-semibold">Exchange</h2>
        <p className="text-sm text-slate-600">Convert between USD and EUR (fixed rate).</p>
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
            const tx = await apiFetch<Transaction>('/transactions/exchange', {
              method: 'POST',
              token,
              body: { from_currency: fromCurrency, to_currency: toCurrency, amount_cents: amountCents },
            })
            setSuccess(tx)
          } catch (e: any) {
            setError(e instanceof HttpError ? e.message : 'exchange_failed')
          } finally {
            setLoading(false)
          }
        }}
      >
        <label className="block">
          <div className="text-xs font-medium text-slate-600">From currency</div>
          <select className="mt-1 w-full rounded border px-3 py-2" value={fromCurrency} onChange={(e) => setFromCurrency(e.target.value as Currency)}>
            <option value="USD">USD</option>
            <option value="EUR">EUR</option>
          </select>
        </label>

        <label className="block">
          <div className="text-xs font-medium text-slate-600">Amount</div>
          <input className="mt-1 w-full rounded border px-3 py-2" value={amount} onChange={(e) => setAmount(e.target.value)} placeholder="10.00" />
        </label>

        <div className="rounded border bg-slate-50 p-3 text-sm text-slate-700">
          <div className="font-medium">Preview</div>
          {preview ? (
            <div className="mt-1">
              Rate: {preview.rate.toFixed(6)} <br />
              You pay: {fromCurrency} {centsToDecimal(preview.cents)} <br />
              You receive: {toCurrency} {centsToDecimal(preview.converted)}
            </div>
          ) : (
            <div className="mt-1 text-slate-500">Enter a valid amount.</div>
          )}
        </div>

        {error ? <div className="rounded bg-red-50 p-2 text-sm text-red-700">{error}</div> : null}
        {success ? <div className="rounded bg-green-50 p-2 text-sm text-green-700">Exchange created: {success.id}</div> : null}

        <button className="rounded bg-slate-900 px-3 py-2 text-sm font-medium text-white disabled:opacity-50" disabled={loading}>
          {loading ? 'Exchanging…' : 'Exchange'}
        </button>
      </form>
    </div>
  )
}

