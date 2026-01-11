const BASE_URL = 'http://localhost:8000/api/v1'

interface FetchOptions extends Omit<RequestInit, 'body'> {
    params?: Record<string, string>
    body?: any
}

export async function apiClient<T>(endpoint: string, options: FetchOptions = {}): Promise<T> {
    const { params, ...customConfig } = options

    const token = localStorage.getItem('token')

    const headers: Record<string, string> = {
        'Content-Type': 'application/json',
    }

    if (token) {
        headers.Authorization = `Bearer ${token}`
    }

    const config: RequestInit = {
        method: customConfig.method || 'GET',
        ...customConfig,
        headers: {
            ...headers,
            ...customConfig.headers,
        },
    }

    if (customConfig.body) {
        config.body = JSON.stringify(customConfig.body)
    }

    let url = `${BASE_URL}${endpoint}`
    if (params) {
        const searchParams = new URLSearchParams(params)
        url += `?${searchParams.toString()}`
    }

    const response = await fetch(url, config)

    if (response.status === 401) {
        localStorage.removeItem('token')
        window.location.href = '/login'
        throw new Error('Unauthorized')
    }

    const data = await response.json()

    if (response.ok) {
        return data
    } else {
        throw new Error(data.error || response.statusText)
    }
}
