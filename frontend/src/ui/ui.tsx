import styled from 'styled-components'
import { NavLink } from 'react-router-dom'

export const Page = styled.div`
  max-width: 64rem;
  margin: 0 auto;
  padding: 1.5rem 1rem;
`

export const Section = styled.div`
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
`

export const Title = styled.h1`
  font-size: 1.5rem;
  font-weight: 700;
  margin: 0;
`

export const Subtitle = styled.p`
  margin: 0.25rem 0 0;
  font-size: 0.875rem;
  color: rgb(71 85 105);
`

export const Card = styled.div`
  border: 1px solid rgb(226 232 240);
  background: white;
  border-radius: 0.75rem;
  padding: 1rem;
`

export const Row = styled.div`
  display: flex;
  gap: 0.75rem;
  flex-wrap: wrap;
  align-items: center;
  justify-content: space-between;
`

export const Grid2 = styled.div`
  display: grid;
  gap: 0.75rem;
  grid-template-columns: repeat(1, minmax(0, 1fr));

  @media (min-width: 768px) {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
`

export const Field = styled.label`
  display: block;
`

export const FieldLabel = styled.div`
  font-size: 0.75rem;
  font-weight: 600;
  color: rgb(71 85 105);
`

export const Input = styled.input`
  margin-top: 0.25rem;
  width: 100%;
  border-radius: 0.5rem;
  border: 1px solid rgb(226 232 240);
  padding: 0.5rem 0.75rem;
  font-size: 0.875rem;
  outline: none;
`

export const Select = styled.select`
  margin-top: 0.25rem;
  width: 100%;
  border-radius: 0.5rem;
  border: 1px solid rgb(226 232 240);
  padding: 0.5rem 0.75rem;
  font-size: 0.875rem;
  outline: none;
  background: white;
`

export const Button = styled.button`
  border-radius: 0.5rem;
  background: rgb(15 23 42);
  color: white;
  padding: 0.5rem 0.75rem;
  font-size: 0.875rem;
  font-weight: 600;
  border: none;
  cursor: pointer;

  &:disabled {
    opacity: 0.5;
    cursor: default;
  }
`

export const GhostButton = styled.button`
  border-radius: 0.5rem;
  background: rgb(241 245 249);
  color: rgb(15 23 42);
  padding: 0.5rem 0.75rem;
  font-size: 0.875rem;
  font-weight: 600;
  border: 1px solid rgb(226 232 240);
  cursor: pointer;
`

export const ErrorBox = styled.div`
  border-radius: 0.5rem;
  background: rgb(254 242 242);
  color: rgb(185 28 28);
  padding: 0.5rem 0.75rem;
  font-size: 0.875rem;
`

export const SuccessBox = styled.div`
  border-radius: 0.5rem;
  background: rgb(240 253 244);
  color: rgb(21 128 61);
  padding: 0.5rem 0.75rem;
  font-size: 0.875rem;
`

export const Nav = styled.nav`
  display: flex;
  gap: 0.5rem;
  flex-wrap: wrap;
  align-items: center;
`

export const NavItem = styled(NavLink)`
  border-radius: 0.5rem;
  padding: 0.5rem 0.75rem;
  font-size: 0.875rem;
  font-weight: 600;
  text-decoration: none;
  color: rgb(51 65 85);

  &:hover {
    background: rgb(226 232 240);
  }

  &.active {
    background: rgb(15 23 42);
    color: white;
  }
`

