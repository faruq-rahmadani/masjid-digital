import { beforeEach, describe, expect, it, vi } from 'vitest'
import { DEFAULT_INQUIRY_URL, inquiryUrl, setupInquiryForm } from './inquiry'

const render = () => {
  document.body.innerHTML = `
    <form data-inquiry>
      <input name="full_name" value=" Ahmad Fauzi " />
      <input name="phone" value="0812-3456-7890" />
      <input name="domicile" value="Banjarbaru" />
      <textarea name="reason">Ingin pasang di TV masjid</textarea>
      <select name="source"><option value="Media sosial" selected>Media sosial</option></select>
      <input type="checkbox" name="consent" checked />
      <input name="website" value="" />
      <p data-inquiry-status hidden></p>
      <button type="submit">Kirim</button>
    </form>
    <div data-inquiry-success hidden tabindex="-1">Terima kasih, <span data-inquiry-name></span>.</div>`
}

const form = () => document.querySelector<HTMLFormElement>('form')!
const status = () => document.querySelector<HTMLElement>('[data-inquiry-status]')!
const success = () => document.querySelector<HTMLElement>('[data-inquiry-success]')!
const button = () => document.querySelector<HTMLButtonElement>('button')!

const submit = async () => {
  form().dispatchEvent(new Event('submit', { cancelable: true }))
  // let the awaited fetch and json settle
  await new Promise((resolve) => setTimeout(resolve, 0))
}

const respond = (status: number, body: unknown) => {
  const headers = { 'Content-Type': 'application/json' }
  return vi.fn().mockResolvedValue(new Response(JSON.stringify(body), { status, headers }))
}

describe('inquiryUrl', () => {
  it('defaults to the API on the same origin', () => {
    expect(inquiryUrl({})).toBe(DEFAULT_INQUIRY_URL)
    expect(inquiryUrl({ VITE_INQUIRY_URL: '  ' })).toBe(DEFAULT_INQUIRY_URL)
    expect(inquiryUrl({ VITE_INQUIRY_URL: 'https://api.example.com/x' })).toBe(
      'https://api.example.com/x',
    )
  })
})

describe('setupInquiryForm', () => {
  beforeEach(render)

  it('posts the fields as JSON and shows the thank-you note', async () => {
    const send = respond(201, { message: 'ok' })
    setupInquiryForm(document, '/api/public/v1/inquiries', send)
    await submit()

    expect(send).toHaveBeenCalledOnce()
    const [url, init] = send.mock.calls[0] as [string, RequestInit]
    expect(url).toBe('/api/public/v1/inquiries')
    expect(init.method).toBe('POST')
    expect(JSON.parse(String(init.body))).toEqual({
      full_name: ' Ahmad Fauzi ',
      phone: '0812-3456-7890',
      domicile: 'Banjarbaru',
      reason: 'Ingin pasang di TV masjid',
      source: 'Media sosial',
      consent: true,
      website: '',
    })
    expect(form().hidden).toBe(true)
    expect(success().hidden).toBe(false)
    expect(success().textContent).toContain('Terima kasih, Ahmad Fauzi.')
  })

  it('shows the API message, capitalized, and keeps the form', async () => {
    setupInquiryForm(
      document,
      '/x',
      respond(429, {
        code: 'TOO_MANY_REQUESTS',
        message: 'terlalu banyak percobaan, silakan coba lagi nanti',
      }),
    )
    await submit()

    expect(status().hidden).toBe(false)
    expect(status().textContent).toBe('Terlalu banyak percobaan, silakan coba lagi nanti')
    expect(form().hidden).toBe(false)
    expect(button().disabled).toBe(false)
    expect(button().textContent).toBe('Kirim')
  })

  it('falls back to a generic message when the answer is not JSON', async () => {
    setupInquiryForm(
      document,
      '/x',
      vi.fn().mockResolvedValue(new Response('<html>502</html>', { status: 502 })),
    )
    await submit()

    expect(status().textContent).toMatch(/belum terkirim/)
  })

  it('tells the visitor when the network fails', async () => {
    setupInquiryForm(document, '/x', vi.fn().mockRejectedValue(new TypeError('Failed to fetch')))
    await submit()

    expect(status().textContent).toMatch(/Tidak bisa terhubung/)
  })

  it('sends consent false when unchecked', async () => {
    document.querySelector<HTMLInputElement>('[name="consent"]')!.checked = false
    const send = respond(400, { message: 'centang persetujuan untuk dihubungi' })
    setupInquiryForm(document, '/x', send)
    await submit()

    const [, init] = send.mock.calls[0] as [string, RequestInit]
    expect(JSON.parse(String(init.body)).consent).toBe(false)
  })
})
