import { Navigate, Route, Routes } from 'react-router-dom'
import { useEffect } from 'react'
import { useAuth } from './auth/AuthContext'
import { Layout } from './components/Layout'
import { DashboardPage } from './pages/DashboardPage'
import { ExchangePage } from './pages/ExchangePage'
import { LoginPage } from './pages/LoginPage'
import { RegisterPage } from './pages/RegisterPage'
import { TransactionsPage } from './pages/TransactionsPage'
import { TransferPage } from './pages/TransferPage'

function Protected(props: { children: React.ReactNode }) {
  const { token, isReady } = useAuth()
  if (!isReady) return <div className="p-6">Loading…</div>
  if (!token) return <Navigate to="/login" replace />
  return props.children
}

function App() {
  const { refreshMe } = useAuth()

  useEffect(() => {
    void refreshMe()
  }, [refreshMe])

  return (
    <Routes>
      <Route path="/login" element={<LoginPage />} />
      <Route path="/register" element={<RegisterPage />} />

      <Route
        path="/"
        element={
          <Protected>
            <Layout />
          </Protected>
        }
      >
        <Route index element={<DashboardPage />} />
        <Route path="transfer" element={<TransferPage />} />
        <Route path="exchange" element={<ExchangePage />} />
        <Route path="transactions" element={<TransactionsPage />} />
      </Route>

      <Route path="*" element={<Navigate to="/" replace />} />
    </Routes>
  )
}

export default App
