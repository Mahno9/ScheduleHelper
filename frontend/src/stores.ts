import { writable, derived, get } from 'svelte/store'
import type { User, Slot, Event } from './api'
import { api } from './api'

// Current logged-in user
export const currentUser = writable<User | null>(null)

// All registered users
export const users = writable<User[]>([])

// Current page tab
export type Tab = 'shared' | 'personal'
export const activeTab = writable<Tab>('shared')

// Slots store (keyed by id for fast updates)
export const slots = writable<Slot[]>([])

// Events store
export const events = writable<Event[]>([])

// Theme: 'light' | 'dark' | 'system'
export const theme = writable<string>('system')

// Undo history for personal calendar actions
export interface UndoAction {
	type: 'create_slot' | 'delete_slot' | 'update_slot'
	slot?: Slot
	prevSlot?: Slot
}
export const undoHistory = writable<UndoAction[]>([])

export function pushUndo(action: UndoAction) {
	undoHistory.update(h => {
		const next = [...h, action]
		return next.slice(-20) // keep last 20 actions
	})
}

// Load users from API
export async function loadUsers() {
	const data = await api.getUsers()
	users.set(data)
}

// Initialize theme from user preference or system
export function applyTheme(t: string) {
	const root = document.documentElement
	if (t === 'system') {
		const prefersDark = window.matchMedia('(prefers-color-scheme: dark)').matches
		root.setAttribute('data-theme', prefersDark ? 'dark' : 'light')
	} else {
		root.setAttribute('data-theme', t)
	}
}

// Derived: effective theme string
export const effectiveTheme = derived(theme, ($theme) => {
	if ($theme === 'system') {
		return window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light'
	}
	return $theme
})

// SSE connection
let sseSource: EventSource | null = null

export function connectSSE() {
	if (sseSource) return
	sseSource = new EventSource('/api/events/sse')

	sseSource.onmessage = (e) => {
		try {
			const msg = JSON.parse(e.data) as { type: string; payload: unknown }
			handleSSEMessage(msg)
		} catch { }
	}
	sseSource.onerror = () => {
		// Reconnect after 3 seconds
		sseSource?.close()
		sseSource = null
		setTimeout(connectSSE, 3000)
	}
}

export function disconnectSSE() {
	sseSource?.close()
	sseSource = null
}

function handleSSEMessage(msg: { type: string; payload: unknown }) {
	switch (msg.type) {
		case 'user_created':
		case 'user_updated': {
			const u = msg.payload as User
			users.update(us => {
				const idx = us.findIndex(x => x.id === u.id)
				if (idx >= 0) { us[idx] = u; return [...us] }
				return [...us, u].sort((a, b) => a.sort_order - b.sort_order)
			})
			break
		}
		case 'user_deleted': {
			const { id } = msg.payload as { id: number }
			users.update(us => us.filter(u => u.id !== id))
			break
		}
		case 'slot_created':
		case 'slot_updated': {
			const s = msg.payload as Slot
			slots.update(ss => {
				const idx = ss.findIndex(x => x.id === s.id)
				if (idx >= 0) { ss[idx] = s; return [...ss] }
				return [...ss, s]
			})
			break
		}
		case 'slot_deleted': {
			const { id } = msg.payload as { id: number }
			slots.update(ss => ss.filter(s => s.id !== id))
			break
		}
		case 'event_created':
		case 'event_updated': {
			const ev = msg.payload as Event
			events.update(evs => {
				const idx = evs.findIndex(x => x.id === ev.id)
				if (idx >= 0) { evs[idx] = ev; return [...evs] }
				return [...evs, ev]
			})
			break
		}
		case 'event_deleted': {
			const { id } = msg.payload as { id: number }
			events.update(evs => evs.filter(e => e.id !== id))
			break
		}
		case 'sort_order_updated': {
			loadUsers()
			break
		}
	}
}
