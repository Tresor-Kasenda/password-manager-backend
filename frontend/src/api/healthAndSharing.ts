import { apiClient } from './apiClient'

export const healthApi = {
    getReport: () => apiClient('/health/report'),
    analyzePassword: (password: string) => apiClient('/health/analyze', {
        method: 'POST',
        body: { password }
    }),
    checkBreach: (password: string) => apiClient('/password/check-breach', {
        method: 'POST',
        body: { password }
    })
}

export const sharingApi = {
    share: (vaultId: string, email: string, options: any = {}) => apiClient('/share', {
        method: 'POST',
        body: {
            vault_id: vaultId,
            recipient_email: email,
            ...options
        }
    }),
    listShares: () => apiClient('/shared'),
    getSharedPassword: (token: string, sharePassword?: string) => apiClient(`/shared/${token}`, {
        params: sharePassword ? { share_password: sharePassword } : {}
    }),
    revokeShare: (token: string) => apiClient(`/shared/${token}/revoke`, {
        method: 'POST'
    })
}
