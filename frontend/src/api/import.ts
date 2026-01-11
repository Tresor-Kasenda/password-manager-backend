import { apiClient } from './apiClient'

export const importApi = {
    getSupportedFormats: () => apiClient('/import/supported-formats'),
    upload: (content: string, filename: string, source: string) => apiClient('/import/upload', {
        method: 'POST',
        body: { content, filename, source }
    }),
    confirm: (sessionId: string, masterPassword: string, mergeStrategy: string = 'skip') => apiClient(`/import/confirm/${sessionId}`, {
        method: 'POST',
        body: { master_password: masterPassword, merge_strategy: mergeStrategy }
    })
}
