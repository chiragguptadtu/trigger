import { client } from './client'

export interface CommandInput {
  name: string
  label: string
  type: 'open' | 'closed'
  options?: string[]
  multi: boolean
  required: boolean
}

export interface Command {
  id: string
  slug: string
  name: string
  description: string
  inputs: CommandInput[]
}

export async function fetchCommands(): Promise<Command[]> {
  const { data } = await client.get<Command[]>('/commands')
  return data
}

export async function fetchCommand(slug: string): Promise<Command> {
  const { data } = await client.get<Command>(`/commands/${slug}`)
  return data
}
