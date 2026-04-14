import { client } from './client'

export async function login(email: string, password: string): Promise<string> {
  const { data } = await client.post<{ token: string }>('/auth/login', { email, password })
  return data.token
}
