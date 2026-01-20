import { useEffect, useState } from 'react'
import { apiFetch, HttpError } from '../api/client'
import { useAuth } from '../auth/AuthContext'
import type { Transaction, TransactionType } from '../api/types'
import { centsToDecimal } from '../lib/money'

export function TransactionsPage() {
  const { token } = useAuth()
  const [type, setType] = useState<TransactionType | ''>('')
  const [page, setPage] = useState(1)
  const [limit, setLimit] = useState(20)
  const [items, setItems] = useState<Transaction[]>([])
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    let cancelled = false
    async function run() {
      if (!token) return
      setLoading(true)
      setError(null)
      try {
        const t = await apiFetch<Transaction[]>('/transactions', { token, query: { type, page, limit } })
        if (cancelled) return
        setItems(t)
      } catch (e: any) {
        if (cancelled) return
        setError(e instanceof HttpError ? e.message : 'load_failed')
      } finally {
        if (!cancelled) setLoading(false)
      }
    }
    void run()
    return () => {
      cancelled = true
    }
  }, [token, type, page, limit])

  return (
    <div className="space-y-4">
      <div className="flex flex-wrap items-end justify-between gap-3">
        <div>
          <h2 className="text-xl font-semibold">Transaction history</h2>
          <p className="text-sm text-slate-600">Filter and paginate.</p>
        </div>

        <div className="flex flex-wrap gap-2">
          <label className="block">
            <div className="text-xs font-medium text-slate-600">Type</div>
            <select
              className="mt-1 rounded border px-3 py-2 text-sm"
              value={type}
              onChange={(e) => {
                setPage(1)
                setType(e.target.value as any)
              }}
            >
              <option value="">All</option>
              <option value="transfer">transfer</option>
              <option value="exchange">exchange</option>
            </select>
          </label>
          <label className="block">
            <div className="text-xs font-medium text-slate-600">Limit</div>
            <select className="mt-1 rounded border px-3 py-2 text-sm" value={limit} onChange={(e) => setLimit(Number(e.target.value))}>
              <option value={10}>10</option>
              <option value={20}>20</option>
              <option value={50}>50</option>
            </select>
          </label>
          <div className="flex items-end gap-2">
            <button className="rounded border px-3 py-2 text-sm disabled:opacity-50" disabled={page <= 1 || loading} onClick={() => setPage((p) => p - 1)}>
              Prev
            </button>
            <button className="rounded border px-3 py-2 text-sm disabled:opacity-50" disabled={loading || items.length < limit} onClick={() => setPage((p) => p + 1)}>
              Next
            </button>
          </div>
        </div>
      </div>

      {error ? <div className="rounded border border-red-200 bg-red-50 p-3 text-sm text-red-700">{error}</div> : null}

      <div className="overflow-hidden rounded-lg border bg-white">
        <table className="w-full text-sm">
          <thead className="bg-slate-50 text-left text-xs text-slate-500">
            <tr>
              <th className="px-3 py-2">Time</th>
              <th className="px-3 py-2">Type</th>
              <th className="px-3 py-2">Amount</th>
              <th className="px-3 py-2">Description</th>
            </tr>
          </thead>
          <tbody>
            {items.map((t) => (
              <tr key={t.id} className="border-t">
                <td className="px-3 py-2">{new Date(t.created_at).toLocaleString()}</td>
                <td className="px-3 py-2">{t.type}</td>
                <td className="px-3 py-2">
                  {t.currency} {centsToDecimal(t.amount_cents)}
                  {t.converted_amount_cents != null ? ` → ${centsToDecimal(t.converted_amount_cents)}` : ''}
                </td>
                <td className="px-3 py-2 text-slate-600">{t.description}</td>
              </tr>
            ))}
            {loading ? (
              <tr>
                <td className="px-3 py-3 text-slate-500" colSpan={4}>
                  Loading…
                </td>
              </tr>
            ) : null}
            {!loading && items.length === 0 ? (
              <tr>
                <td className="px-3 py-3 text-slate-500" colSpan={4}>
                  No transactions found.
                </td>
              </tr>
            ) : null}
          </tbody>
        </table>
      </div>

      <div className="text-sm text-slate-600">Page: {page}</div>
    </div>
  )
}

