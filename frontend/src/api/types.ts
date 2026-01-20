export type Currency = 'USD' | 'EUR'
export type TransactionType = 'transfer' | 'exchange'

export type User = {
  id: string
  email: string
  first_name: string
  last_name: string
  created_at: string
  updated_at: string
}

export type AuthResponse = {
  access_token: string
  refresh_token: string
  user: User
}

export type TokenResponse = {
  access_token: string
  refresh_token: string
}

export type Account = {
  id: string
  currency: Currency
  balance_cents: number
}

export type Transaction = {
  id: string
  type: TransactionType
  from_account_id?: string
  to_account_id: string
  amount_cents: number
  currency: Currency
  exchange_rate?: number
  converted_amount_cents?: number
  description: string
  created_at: string
  from_user_email?: string
  to_user_email?: string
}

export type ApiError =
  | { error: string }
  | { error: string; fields: Array<{ field: string; message: string }> }

