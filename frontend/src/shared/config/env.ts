export const config = {
  apiBaseUrl: import.meta.env.VITE_API_BASE_URL ?? '/api/v1',
  wsUrl: import.meta.env.VITE_WS_URL ?? '/api/v1/ws',
  livekitUrl: import.meta.env.VITE_LIVEKIT_URL ?? 'ws://localhost:7880',
} as const
