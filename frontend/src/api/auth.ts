import { apiClient } from './apiClient'

export const authApi = {
    register: (data: any) => {
        return apiClient('/auth/register', {
            method: 'POST',
            body: data,
        })
    },

    login: (data: any) => {
        return apiClient('/auth/login', {
            method: 'POST',
            body: data,
        })
    },

    getProfile: () => {
        // In many APIs there's a /me or /profile endpoint, 
        // but the backend router doesn't show one explicitly for now.
        // We'll stick to login/register for now as requested.
        return apiClient('/vault') // Testing token with a protected route
    }
}
