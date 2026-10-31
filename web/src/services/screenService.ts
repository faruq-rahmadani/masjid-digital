import type { Screen } from '@/types/home'

// Not in the API yet (E9-02 heartbeat). When it lands:
//   const resp = await apiGet<DataResponse<Screen[]>>('/v1/screens')
//   return resp.data

export async function listScreens(): Promise<Screen[]> {
  return mockScreens()
}

// --- Mock ----------------------------------------------------------------------------------

function mockScreens(): Screen[] {
  const minutesAgo = (m: number) => new Date(Date.now() - m * 60_000).toISOString()
  return [
    { id: 1, name: 'TV Ruang Utama', status: 'online', last_seen_at: minutesAgo(1) },
    { id: 2, name: 'TV Teras', status: 'offline', last_seen_at: minutesAgo(125) },
  ]
}
