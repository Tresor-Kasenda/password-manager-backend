import { apiClient } from './apiClient'

export const twofaApi = {
    enable: (masterPassword: string) => apiClient('/2fa/enable', {
        method: 'POST',
        body: { master_password: masterPassword }
    }),
    verifyAndEnable: (token: string) => apiClient('/2fa/verify-and-enable', {
        method: 'POST',
        body: { token }
    }),
    verify: (token: string) => apiClient('/2fa/verify', {
        method: 'POST',
        body: { token }
    }),
    disable: (masterPassword: string) => apiClient('/2fa/disable', {
        method: 'POST',
        body: { master_password: masterPassword }
    })
}
