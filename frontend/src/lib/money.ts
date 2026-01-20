export function centsToDecimal(cents: number): string {
  const sign = cents < 0 ? '-' : ''
  const abs = Math.abs(cents)
  const whole = Math.floor(abs / 100)
  const frac = abs % 100
  return `${sign}${whole}.${String(frac).padStart(2, '0')}`
}

export function decimalToCents(input: string): number {
  const raw = input.trim()
  if (!raw) throw new Error('amount is required')

  const m = raw.match(/^([+-])?(\d+)(?:\.(\d{0,2}))?$/)
  if (!m) throw new Error('amount must be a number with up to 2 decimals')

  const sign = m[1] === '-' ? -1 : 1
  const whole = Number(m[2])
  const fracRaw = (m[3] ?? '').padEnd(2, '0')
  const frac = fracRaw ? Number(fracRaw) : 0
  return sign * (whole * 100 + frac)
}

