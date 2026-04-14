export function getApiError(err: unknown, fallback = 'An error occurred.'): string {
  return (err as { response?: { data?: { error?: string } } })?.response?.data?.error ?? fallback
}
