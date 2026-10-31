export type MosqueTimezone = 'Asia/Jakarta' | 'Asia/Makassar' | 'Asia/Jayapura'

export type MosqueProfile = {
  name: string
  address: string
  contact_phone: string
  latitude: number
  longitude: number
  timezone: MosqueTimezone
}
