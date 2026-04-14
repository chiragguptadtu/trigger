import { client } from './client'

export interface ImportError {
  filename: string
  error: string
  failed_at: string
}

export interface User {
  id: string
  email: string
  name: string
  is_admin: boolean
  is_active: boolean
  created_at: string
}

export interface Group {
  id: string
  name: string
  is_active: boolean
  created_at: string
}

export interface ConfigEntry {
  id: string
  key: string
  description: string
  created_at: string
  updated_at: string
}

export interface Permission {
  id: string
  command_id: string
  grantee_type: 'user' | 'group'
  grantee_id: string
}

export async function fetchImportErrors(): Promise<ImportError[]> {
  const { data } = await client.get<ImportError[]>('/commands/import-errors')
  return data
}

// Users
export async function fetchUsers(): Promise<User[]> {
  const { data } = await client.get<User[]>('/admin/users')
  return data
}
export async function createUser(payload: { email: string; name: string; password: string; is_admin: boolean }): Promise<User> {
  const { data } = await client.post<User>('/admin/users', payload)
  return data
}
export async function updateUser(id: string, payload: { name?: string; password?: string; is_admin?: boolean; is_active?: boolean }): Promise<User> {
  const { data } = await client.patch<User>(`/admin/users/${id}`, payload)
  return data
}
export async function deactivateUser(id: string): Promise<void> {
  await client.delete(`/admin/users/${id}`)
}

// Groups
export async function fetchGroups(): Promise<Group[]> {
  const { data } = await client.get<Group[]>('/admin/groups')
  return data
}
export async function createGroup(name: string): Promise<Group> {
  const { data } = await client.post<Group>('/admin/groups', { name })
  return data
}
export async function deleteGroup(id: string): Promise<void> {
  await client.delete(`/admin/groups/${id}`)
}
export async function fetchGroupMembers(id: string): Promise<User[]> {
  const { data } = await client.get<User[]>(`/admin/groups/${id}/members`)
  return data
}
export async function addGroupMember(groupId: string, userId: string): Promise<void> {
  await client.post(`/admin/groups/${groupId}/members`, { user_id: userId })
}
export async function removeGroupMember(groupId: string, userId: string): Promise<void> {
  await client.delete(`/admin/groups/${groupId}/members/${userId}`)
}

// Config
export async function fetchConfig(): Promise<ConfigEntry[]> {
  const { data } = await client.get<ConfigEntry[]>('/admin/config')
  return data
}
export async function createConfig(payload: { key: string; value: string; description: string }): Promise<ConfigEntry> {
  const { data } = await client.post<ConfigEntry>('/admin/config', payload)
  return data
}
export async function updateConfig(key: string, payload: { value: string; description: string }): Promise<ConfigEntry> {
  const { data } = await client.put<ConfigEntry>(`/admin/config/${key}`, payload)
  return data
}
export async function deleteConfig(key: string): Promise<void> {
  await client.delete(`/admin/config/${key}`)
}

// Permissions
export async function fetchPermissions(slug: string): Promise<Permission[]> {
  const { data } = await client.get<Permission[]>(`/admin/commands/${slug}/permissions`)
  return data
}
export async function grantPermission(slug: string, payload: { grantee_type: string; grantee_id: string }): Promise<Permission> {
  const { data } = await client.post<Permission>(`/admin/commands/${slug}/permissions`, payload)
  return data
}
export async function revokePermission(slug: string, granteeType: string, granteeId: string): Promise<void> {
  await client.delete(`/admin/commands/${slug}/permissions/${granteeType}/${granteeId}`)
}
