import { apiClient } from './apiClient'

export const vaultApi = {
    list: () => apiClient('/vault'),

    get: (id: string, masterPassword: string) => {
        return apiClient(`/vault/${id}`, {
            params: { master_password: masterPassword }
        })
    },

    create: (data: any) => {
        return apiClient('/vault', {
            method: 'POST',
            body: data,
        })
    },

    update: (id: string, data: any) => {
        return apiClient(`/vault/${id}`, {
            method: 'PUT',
            body: data,
        })
    },

    delete: (id: string) => {
        return apiClient(`/vault/${id}`, {
            method: 'DELETE',
        })
    },

    generatePassword: (length: number = 20, useSpecial: boolean = true) => {
        return apiClient('/vault/generate-password', {
            method: 'POST',
            params: {
                length: length.toString(),
                use_special: useSpecial.toString(),
            }
        }) as Promise<{ password: string }>
    }
}
