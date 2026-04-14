import { client } from './client'

export interface Execution {
  id: string
  command_id: string
  status: string
  error_message: string
  created_at: string
  started_at?: string
  completed_at?: string
  triggered_by_name: string
  triggered_by_email: string
  inputs?: Record<string, unknown>
}

export async function triggerExecution(
  slug: string,
  inputs: Record<string, unknown>,
): Promise<Execution> {
  const { data } = await client.post<Execution>(`/commands/${slug}/executions`, { inputs })
  return data
}

export async function fetchExecution(id: string): Promise<Execution> {
  const { data } = await client.get<Execution>(`/executions/${id}`)
  return data
}

export async function fetchExecutions(slug: string): Promise<Execution[]> {
  const { data } = await client.get<Execution[]>(`/commands/${slug}/executions`)
  return data
}
