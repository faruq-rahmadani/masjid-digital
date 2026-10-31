import '@testing-library/jest-dom/vitest'
import { cleanup } from '@testing-library/react'
import { afterEach, vi } from 'vitest'
import { useSessionStore } from '@/stores/sessionStore'
import { resetFakeApi } from './fakeApi'

// jsdom has no layout engine; ScrollRestoration calls this on every navigation.
window.scrollTo = () => {}

// jsdom doesn't implement <dialog> modality; enough of it for Sheet to open and close.
HTMLDialogElement.prototype.showModal = function showModal(this: HTMLDialogElement) {
  this.open = true
}
HTMLDialogElement.prototype.close = function close(this: HTMLDialogElement) {
  this.open = false
  this.dispatchEvent(new Event('close'))
}

// Tests run against the strict fakes in ./fakeApi, never the network or the quick mocks
// inside src/services, so they pass whether a service is mocked or live.
// (vi.mock is hoisted above imports, so the factories load the fakes themselves.)
vi.mock('@/services/authService', async () => (await import('./fakeApi')).fakeAuthService)
vi.mock('@/services/userService', async () => (await import('./fakeApi')).fakeUserService)
vi.mock(
  '@/services/mosqueProfileService',
  async () => (await import('./fakeApi')).fakeMosqueProfileService,
)
vi.mock(
  '@/services/prayerTimeService',
  async () => (await import('./fakeApi')).fakePrayerTimeService,
)
vi.mock('@/services/screenService', async () => (await import('./fakeApi')).fakeScreenService)
vi.mock('@/services/activityService', async () => (await import('./fakeApi')).fakeActivityService)

afterEach(() => {
  cleanup()
  resetFakeApi()
  localStorage.clear()
  sessionStorage.clear()
  useSessionStore.setState({ session: null })
})
