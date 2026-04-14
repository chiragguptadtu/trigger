export interface TokenClaims {
  uid: string
  admin: boolean
}

export function decodeClaims(): TokenClaims | null {
  const token = localStorage.getItem('token')
  if (!token) return null
  try {
    const payload = token.split('.')[1]
    return JSON.parse(atob(payload)) as TokenClaims
  } catch {
    return null
  }
}
