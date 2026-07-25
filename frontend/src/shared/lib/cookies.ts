/** Read a cookie value from document.cookie (empty string if missing). */
export function getCookie(name: string): string {
  if (typeof document === 'undefined') {
    return ''
  }
  const prefix = `${name}=`
  for (const part of document.cookie.split(';')) {
    const trimmed = part.trim()
    if (trimmed.startsWith(prefix)) {
      return decodeURIComponent(trimmed.slice(prefix.length))
    }
  }
  return ''
}
