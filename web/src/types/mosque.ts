export type MosqueTimezone = 'Asia/Jakarta' | 'Asia/Makassar' | 'Asia/Jayapura'

export type MosqueProfile = {
  name: string
  address: string
  contact_phone: string
  latitude: number
  longitude: number
  timezone: MosqueTimezone
}

/** PUT /v1/mosque-profile replaces every field, so it takes the whole profile. */
export type UpdateMosqueProfileRequest = MosqueProfile

export type MosqueLogo = {
  /** `data:<content type>;base64,...`, usable as an <img> src. */
  url: string
}
