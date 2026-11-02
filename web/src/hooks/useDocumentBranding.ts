import { useEffect } from 'react'
import { usePublicMosqueLogo, usePublicMosqueProfile } from '@/hooks/useMosqueProfile'

const DEFAULT_TITLE = 'Masjid Digital'

/**
 * Browser tab: the mosque's name as the title and its logo as the favicon, from the public
 * endpoints so the login screen and the TV get them too. Falls back to "Masjid Digital" and the
 * browser's default icon until the profile or logo is saved.
 */
export function useDocumentBranding() {
  const name = usePublicMosqueProfile({ live: false }).data?.name
  const logo = usePublicMosqueLogo({ live: false }).data

  useEffect(() => {
    document.title = name || DEFAULT_TITLE
  }, [name])

  useEffect(() => {
    let link = document.querySelector<HTMLLinkElement>('link[rel="icon"]')
    if (!logo) {
      link?.remove()
      return
    }
    if (!link) {
      link = document.createElement('link')
      link.rel = 'icon'
      document.head.append(link)
    }
    link.href = logo
  }, [logo])
}
