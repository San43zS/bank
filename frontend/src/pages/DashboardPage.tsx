import { useEffect, useMemo, useState } from 'react'
import { apiFetch, HttpError } from '../api/client'
import { useAuth } from '../auth/AuthContext'
import type { Account, Transaction } from '../api/types'
import { centsToDecimal } from '../lib/money'
import { Card, ErrorBox, Grid2, Section, Subtitle, Title } from '../ui/ui'

export function DashboardPage() {
  const { token } = useAuth()
  const [accounts, setAccounts] = useState<Account[] | null>(null)
  const [txs, setTxs] = useState<Transaction[] | null>(null)
  const [error, setError] = useState<string | null>(null)

  const byCurrency = useMemo(() => {
    const map: Record<string, Account | undefined> = {}
    for (const a of accounts ?? []) map[a.currency] = a
    return map
  }, [accounts])

  useEffect(() => {
    let cancelled = false
    async function run() {
      if (!token) return
      setError(null)
      try {
        const [a, t] = await Promise.all([
          apiFetch<Account[]>('/accounts', { token }),
          apiFetch<Transaction[]>('/transactions', { token, query: { page: 1, limit: 5 } }),
        ])
        if (cancelled) return
        setAccounts(a)
        setTxs(t)
      } catch (e: any) {
        if (cancelled) return
        setError(e instanceof HttpError ? e.message : 'load_failed')
      }
    }
    void run()
    return () => {
      cancelled = true
    }
  }, [token])

  return (
    <Section>
      <div>
        <Title>Balances</Title>
        <Subtitle>Your USD and EUR wallets.</Subtitle>
      </div>

      {error ? <ErrorBox>{error}</ErrorBox> : null}

      <Grid2>
        <Card>
          <div style={{ fontSize: '0.75rem', fontWeight: 600, color: 'rgb(100 116 139)' }}>USD</div>
          <div style={{ marginTop: '0.25rem', fontSize: '1.5rem', fontWeight: 700 }}>
            ${centsToDecimal(byCurrency['USD']?.balance_cents ?? 0)}
          </div>
          <div style={{ marginTop: '0.25rem', fontSize: '0.75rem', color: 'rgb(100 116 139)' }}>
            Account: {byCurrency['USD']?.id ?? '—'}
          </div>
        </Card>
        <Card>
          <div style={{ fontSize: '0.75rem', fontWeight: 600, color: 'rgb(100 116 139)' }}>EUR</div>
          <div style={{ marginTop: '0.25rem', fontSize: '1.5rem', fontWeight: 700 }}>
            €{centsToDecimal(byCurrency['EUR']?.balance_cents ?? 0)}
          </div>
          <div style={{ marginTop: '0.25rem', fontSize: '0.75rem', color: 'rgb(100 116 139)' }}>
            Account: {byCurrency['EUR']?.id ?? '—'}
          </div>
        </Card>
      </Grid2>

      <div>
        <Title>Last 5 transactions</Title>
        <Subtitle>Recent transfers and exchanges.</Subtitle>
      </div>

      <div style={{ overflow: 'hidden', borderRadius: '0.75rem', border: '1px solid rgb(226 232 240)', background: 'white' }}>
        <table style={{ width: '100%', fontSize: '0.875rem' }}>
          <thead style={{ background: 'rgb(248 250 252)', textAlign: 'left', fontSize: '0.75rem', color: 'rgb(100 116 139)' }}>
            <tr>
              <th style={{ padding: '0.5rem 0.75rem' }}>Time</th>
              <th style={{ padding: '0.5rem 0.75rem' }}>Type</th>
              <th style={{ padding: '0.5rem 0.75rem' }}>Amount</th>
              <th style={{ padding: '0.5rem 0.75rem' }}>Details</th>
            </tr>
          </thead>
          <tbody>
            {(txs ?? []).map((t) => (
              <tr key={t.id} style={{ borderTop: '1px solid rgb(226 232 240)' }}>
                <td style={{ padding: '0.5rem 0.75rem' }}>{new Date(t.created_at).toLocaleString()}</td>
                <td style={{ padding: '0.5rem 0.75rem' }}>{t.type}</td>
                <td style={{ padding: '0.5rem 0.75rem' }}>
                  {t.currency} {centsToDecimal(t.amount_cents)}
                  {t.converted_amount_cents != null ? ` → ${centsToDecimal(t.converted_amount_cents)}` : ''}
                </td>
                <td style={{ padding: '0.5rem 0.75rem', color: 'rgb(71 85 105)' }}>{t.description}</td>
              </tr>
            ))}
            {txs && txs.length === 0 ? (
              <tr>
                <td style={{ padding: '0.75rem', color: 'rgb(100 116 139)' }} colSpan={4}>
                  No transactions yet.
                </td>
              </tr>
            ) : null}
          </tbody>
        </table>
      </div>
    </Section>
  )
}

