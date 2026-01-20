import { createContext, useCallback, useContext, useMemo, useState } from 'react'
import { apiFetch } from '../api/client'
import type { AuthResponse, User } from '../api/types'

type AuthState = {
  token: string | null
  user: User | null
  isReady: boolean
}

type AuthContextValue = AuthState & {
  login: (email: string, password: string) => Promise<void>
  register: (input: { email: string; password: string; firstName: string; lastName: string }) => Promise<void>
  logout: () => Promise<void>
  refreshMe: () => Promise<void>
}

const AuthContext = createContext<AuthContextValue | undefined>(undefined)

const TOKEN_KEY = 'banking_access_token'

export function AuthProvider(props: { children: React.ReactNode }) {
  const [token, setToken] = useState<string | null>(() => localStorage.getItem(TOKEN_KEY))
  const [user, setUser] = useState<User | null>(null)
  const [isReady, setIsReady] = useState<boolean>(false)

  const setSession = useCallback((auth: AuthResponse | null) => {
    if (!auth) {
      localStorage.removeItem(TOKEN_KEY)
      setToken(null)
      setUser(null)
      return
    }
    localStorage.setItem(TOKEN_KEY, auth.access_token)
    setToken(auth.access_token)
    setUser(auth.user)
  }, [])

  const refreshMe = useCallback(async () => {
    try {
      const t = localStorage.getItem(TOKEN_KEY)
      if (!t) {
        setIsReady(true)
        return
      }
      const me = await apiFetch<User>('/auth/me', { token: t })
      setToken(t)
      setUser(me)
    } finally {
      setIsReady(true)
    }
  }, [])

  const login = useCallback(
    async (email: string, password: string) => {
      const auth = await apiFetch<AuthResponse>('/auth/login', { method: 'POST', body: { email, password } })
      setSession(auth)
    },
    [setSession],
  )

  const register = useCallback(
    async (input: { email: string; password: string; firstName: string; lastName: string }) => {
      const auth = await apiFetch<AuthResponse>('/auth/register', {
        method: 'POST',
        body: { email: input.email, password: input.password, first_name: input.firstName, last_name: input.lastName },
      })
      setSession(auth)
    },
    [setSession],
  )

  const logout = useCallback(async () => {
    const t = localStorage.getItem(TOKEN_KEY)
    try {
      if (t) {
        const body = { access_token: t, refresh_token: '' }
        await apiFetch<void>('/auth/logout', { method: 'POST', token: t, body })
      }
    } catch {
    } finally {
      localStorage.removeItem(TOKEN_KEY)
      setToken(null)
      setUser(null)
    }
  }, [])

  const value = useMemo<AuthContextValue>(
    () => ({ token, user, isReady, login, register, logout, refreshMe }),
    [token, user, isReady, login, register, logout, refreshMe],
  )

  return <AuthContext.Provider value={value}>{props.children}</AuthContext.Provider>
}

export function useAuth() {
  const ctx = useContext(AuthContext)
  if (!ctx) throw new Error('useAuth must be used within AuthProvider')
  return ctx
}

