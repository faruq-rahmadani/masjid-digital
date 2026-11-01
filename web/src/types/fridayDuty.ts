// Petugas Jumat (E5): who leads each Friday's prayer, shown on the TV's Jumat screen.

/** Petugas Jumat of one Friday. */
export type FridayDuty = {
  date: string /** "YYYY-MM-DD", a Friday */
  khatib: string
  imam: string
  muadzin: string
  theme: string /** tema khutbah */
  updated_by: number
  created_at: string
  updated_at: string
}

export type SaveFridayDutyRequest = Pick<FridayDuty, 'khatib' | 'imam' | 'muadzin' | 'theme'>
