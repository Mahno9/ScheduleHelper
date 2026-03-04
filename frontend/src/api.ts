// API Client for ScheduleHelper backend

export interface User {
	id: number
	name: string
	emoji: string
	color: string
	theme: string
	timezone: string
	auto_timezone: boolean
	sort_order: number
	created_at: string
}

export interface Slot {
	id: number
	user_id: number
	start_time: string
	end_time: string
	comment: string
	created_at: string
	updated_at: string
}

export interface Event {
	id: number
	title: string
	description: string
	icon: string
	color: string
	start_time: string
	end_time: string
	created_by: number
	created_at: string
	updated_at: string
}

export interface CalendarData {
	users: User[]
	slots: Slot[]
	events: Event[]
}

const BASE = '/api'

async function request<T>(method: string, path: string, body?: unknown): Promise<T> {
	const res = await fetch(BASE + path, {
		method,
		headers: body ? { 'Content-Type': 'application/json' } : {},
		body: body ? JSON.stringify(body) : undefined,
	})
	if (res.status === 204) return undefined as T
	if (!res.ok) {
		const err = await res.json().catch(() => ({ error: res.statusText }))
		throw new Error(err.error || res.statusText)
	}
	return res.json()
}

// Users
export const api = {
	getUsers: () => request<User[]>('GET', '/users'),
	register: (data: { name: string; emoji: string; color: string; timezone: string; auto_timezone: boolean }) =>
		request<User>('POST', '/register', data),
	login: (user_id: number) => request<User>('POST', '/login', { user_id }),
	getProfile: (id: number) => request<User>('GET', `/profile/${id}`),
	updateProfile: (id: number, data: Partial<User>) => request<User>('PUT', `/profile/${id}`, data),
	deleteProfile: (id: number) => request<void>('DELETE', `/profile/${id}`),
	updateSortOrder: (order: number[]) => request<void>('PUT', '/users/sort', { order }),

	// Slots
	getSlots: (params?: { user_id?: number; from?: string; to?: string }) => {
		const qs = params ? '?' + Object.entries(params).filter(([, v]) => v != null).map(([k, v]) => `${k}=${encodeURIComponent(String(v))}`).join('&') : ''
		return request<Slot[]>('GET', '/slots' + qs)
	},
	createSlot: (data: { user_id: number; start_time: string; end_time: string; comment: string }) =>
		request<Slot>('POST', '/slots', data),
	updateSlot: (id: number, data: { start_time: string; end_time: string; comment: string }) =>
		request<Slot>('PUT', `/slots/${id}`, data),
	deleteSlot: (id: number) => request<void>('DELETE', `/slots/${id}`),

	// Events
	getEvents: (params?: { from?: string; to?: string }) => {
		const qs = params ? '?' + Object.entries(params).filter(([, v]) => v != null).map(([k, v]) => `${k}=${encodeURIComponent(String(v))}`).join('&') : ''
		return request<Event[]>('GET', '/events' + qs)
	},
	createEvent: (data: Omit<Event, 'id' | 'created_at' | 'updated_at'>) =>
		request<Event>('POST', '/events', data),
	updateEvent: (id: number, data: Partial<Event>) => request<Event>('PUT', `/events/${id}`, data),
	deleteEvent: (id: number) => request<void>('DELETE', `/events/${id}`),

	// Calendar
	getCalendar: (params?: { from?: string; to?: string }) => {
		const qs = params ? '?' + Object.entries(params).filter(([, v]) => v != null).map(([k, v]) => `${k}=${encodeURIComponent(String(v))}`).join('&') : ''
		return request<CalendarData>('GET', '/calendar' + qs)
	},
}
