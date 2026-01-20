import { Outlet, useNavigate } from 'react-router-dom'
import { useAuth } from '../auth/AuthContext'
import { Button, Nav, NavItem, Page, Row } from '../ui/ui'

export function Layout() {
  const { user, logout } = useAuth()
  const nav = useNavigate()

  return (
    <div>
      <div style={{ background: 'white', borderBottom: '1px solid rgb(226 232 240)' }}>
        <Page>
          <Row>
            <div>
              <div style={{ fontWeight: 700, fontSize: '1.125rem' }}>Mini Banking</div>
              <div style={{ fontSize: '0.75rem', color: 'rgb(100 116 139)' }}>{user?.email ?? '—'}</div>
            </div>

            <Nav>
              <NavItem to="/" end>
                Dashboard
              </NavItem>
              <NavItem to="/transfer">Transfer</NavItem>
              <NavItem to="/exchange">Exchange</NavItem>
              <NavItem to="/transactions">History</NavItem>
              <Button
                type="button"
                onClick={async () => {
                  await logout()
                  nav('/login')
                }}
              >
                Logout
              </Button>
            </Nav>
          </Row>
        </Page>
      </div>

      <Page>
        <Outlet />
      </Page>
    </div>
  )
}

