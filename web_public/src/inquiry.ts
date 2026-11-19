export type InquiryEnv = { VITE_INQUIRY_URL?: string }

/** The API's inquiry route, on the site's own origin unless VITE_INQUIRY_URL says otherwise. */
export const DEFAULT_INQUIRY_URL = '/api/public/v1/inquiries'

export function inquiryUrl(env: InquiryEnv): string {
  return env.VITE_INQUIRY_URL?.trim() || DEFAULT_INQUIRY_URL
}

const NETWORK_ERROR = 'Tidak bisa terhubung. Periksa internet Anda, lalu coba lagi.'
const UNKNOWN_ERROR = 'Pesan belum terkirim. Silakan coba lagi beberapa saat lagi.'

/** The API answers {message} in Indonesian, ready to show; anything else gets a generic line. */
async function errorMessage(res: Response): Promise<string> {
  try {
    const body: unknown = await res.json()
    if (
      body &&
      typeof body === 'object' &&
      'message' in body &&
      typeof body.message === 'string' &&
      body.message
    ) {
      const text = body.message
      return text.charAt(0).toUpperCase() + text.slice(1)
    }
  } catch {
    // not JSON, e.g. a proxy's error page
  }
  return UNKNOWN_ERROR
}

/**
 * Sends `form[data-inquiry]` as JSON to url. The browser checks the required fields first; the
 * API's message is shown in `[data-inquiry-status]`, and on success the form is replaced by
 * `[data-inquiry-success]` (its `[data-inquiry-name]` gets the sender's name).
 */
export function setupInquiryForm(root: ParentNode, url: string, send: typeof fetch = fetch) {
  const form = root.querySelector<HTMLFormElement>('form[data-inquiry]')
  if (!form) return
  const status = form.querySelector<HTMLElement>('[data-inquiry-status]')
  const button = form.querySelector<HTMLButtonElement>('button[type="submit"]')
  const success = root.querySelector<HTMLElement>('[data-inquiry-success]')
  const idleLabel = button?.textContent ?? ''

  const showError = (text: string) => {
    if (!status) return
    status.textContent = text
    status.hidden = false
  }

  form.addEventListener('submit', async (event) => {
    event.preventDefault()
    if (button?.disabled) return

    const data = new FormData(form)
    const text = (name: string) => String(data.get(name) ?? '')
    const body = {
      full_name: text('full_name'),
      phone: text('phone'),
      domicile: text('domicile'),
      reason: text('reason'),
      source: text('source'),
      consent: data.get('consent') !== null,
      website: text('website'),
    }

    if (status) status.hidden = true
    if (button) {
      button.disabled = true
      button.textContent = 'Mengirim…'
    }

    try {
      const res = await send(url, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(body),
      })
      if (!res.ok) {
        showError(await errorMessage(res))
        return
      }

      form.hidden = true
      if (success) {
        const name = success.querySelector<HTMLElement>('[data-inquiry-name]')
        if (name) name.textContent = body.full_name.trim()
        success.hidden = false
        success.focus()
      }
    } catch {
      showError(NETWORK_ERROR)
    } finally {
      if (button) {
        button.disabled = false
        button.textContent = idleLabel
      }
    }
  })
}
