import type { Activity } from '@/types/home'

// Not in the API yet (E2-05 activity log; the PRD shows the latest 4 on Home). When it lands:
//   const resp = await apiGet<DataResponse<Activity[]>>('/v1/activity-logs?limit=4')
//   return resp.data

export async function listRecentActivity(): Promise<Activity[]> {
  return mockActivity()
}

// --- Mock ----------------------------------------------------------------------------------

function mockActivity(): Activity[] {
  const minutesAgo = (m: number) => new Date(Date.now() - m * 60_000).toISOString()
  return [
    {
      id: 4,
      actor_name: 'Siti Aminah',
      summary: 'menambah pemasukan Rp 750.000',
      created_at: minutesAgo(25),
    },
    {
      id: 3,
      actor_name: 'Hasan Basri',
      summary: 'mengubah iqamah Maghrib jadi 7 mnt',
      created_at: minutesAgo(160),
    },
    {
      id: 2,
      actor_name: 'Hasan Basri',
      summary: 'mengaktifkan slide Kajian Ahad',
      created_at: minutesAgo(26 * 60),
    },
    {
      id: 1,
      actor_name: 'Ahmad Fauzi',
      summary: 'menambah pengguna Bendahara',
      created_at: minutesAgo(6 * 24 * 60),
    },
  ]
}
