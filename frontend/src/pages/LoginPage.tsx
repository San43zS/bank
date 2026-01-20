import { useState } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { HttpError } from '../api/client'
import { useAuth } from '../auth/AuthContext'
import { Button, Card, ErrorBox, Field, FieldLabel, Input, Page, Subtitle, Title } from '../ui/ui'

export function LoginPage() {
  const { login } = useAuth()
  const nav = useNavigate()
  const [email, setEmail] = useState('user1@test.com')
  const [password, setPassword] = useState('password123')
  const [error, setError] = useState<string | null>(null)
  const [loading, setLoading] = useState(false)

  return (
    <Page style={{ maxWidth: '28rem' }}>
      <Title>Login</Title>
      <Subtitle>Use seeded users or register a new one.</Subtitle>

      <Card as="form"
        style={{ marginTop: '1.5rem', display: 'flex', flexDirection: 'column', gap: '0.75rem' }}
        onSubmit={async (e: any) => {
          e.preventDefault()
          setError(null)
          setLoading(true)
          try {
            await login(email, password)
            nav('/')
          } catch (e: any) {
            setError(e instanceof HttpError ? e.message : 'login_failed')
          } finally {
            setLoading(false)
          }
        }}
      >
        <Field>
          <FieldLabel>Email</FieldLabel>
          <Input value={email} onChange={(e) => setEmail(e.target.value)} autoComplete="email" />
        </Field>
        <Field>
          <FieldLabel>Password</FieldLabel>
          <Input type="password" value={password} onChange={(e) => setPassword(e.target.value)} autoComplete="current-password" />
        </Field>

        {error ? <ErrorBox>{error}</ErrorBox> : null}

        <Button style={{ width: '100%' }} disabled={loading}>
          {loading ? 'Signing in…' : 'Sign in'}
        </Button>
      </Card>

      <div style={{ marginTop: '0.75rem', fontSize: '0.875rem', color: 'rgb(71 85 105)' }}>
        No account?{' '}
        <Link style={{ fontWeight: 600, color: 'rgb(15 23 42)', textDecoration: 'underline' }} to="/register">
          Register
        </Link>
      </div>
    </Page>
  )
}

