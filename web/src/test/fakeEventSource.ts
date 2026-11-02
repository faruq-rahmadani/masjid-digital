/**
 * A stand-in for the browser's EventSource (jsdom has none), so TV tests can play the API's side
 * of the live connection: open it, then push events.
 */
export class FakeEventSource {
  static CONNECTING = 0
  static OPEN = 1
  static CLOSED = 2
  static instances: FakeEventSource[] = []

  readyState = FakeEventSource.CONNECTING
  onopen: (() => void) | null = null
  onerror: (() => void) | null = null
  private listeners = new Map<string, ((e: MessageEvent) => void)[]>()

  readonly url: string

  constructor(url: string) {
    this.url = url
    FakeEventSource.instances.push(this)
  }

  addEventListener(type: string, listener: (e: MessageEvent) => void) {
    this.listeners.set(type, [...(this.listeners.get(type) ?? []), listener])
  }

  close() {
    this.readyState = FakeEventSource.CLOSED
  }

  /** The API accepted the connection. */
  open() {
    this.readyState = FakeEventSource.OPEN
    this.onopen?.()
  }

  /** The API sent an event, `data` as it would be JSON-encoded. */
  emit(type: string, data: unknown = null) {
    const event = new MessageEvent(type, { data: JSON.stringify(data) })
    for (const listener of this.listeners.get(type) ?? []) listener(event)
  }

  /** The network dropped; the browser would retry. */
  drop() {
    this.readyState = FakeEventSource.CONNECTING
    this.onerror?.()
  }

  /** The last connection a test opened. */
  static get latest() {
    return FakeEventSource.instances.at(-1)
  }
}
