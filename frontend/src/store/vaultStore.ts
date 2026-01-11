import { reactive, toRefs } from 'vue'
import { vaultApi } from '../api/vault'

interface VaultState {
    items: any[]
    loading: boolean
    error: string | null
}

const state = reactive<VaultState>({
    items: [],
    loading: false,
    error: null,
})

export const useVaultStore = () => {
    const fetchVaults = async () => {
        state.loading = true
        state.error = null
        try {
            state.items = await vaultApi.list() as any[]
        } catch (err: any) {
            state.error = err.message || 'Failed to fetch vaults'
        } finally {
            state.loading = false
        }
    }

    const addItem = (item: any) => {
        state.items.unshift(item)
    }

    const updateItem = (updatedItem: any) => {
        const index = state.items.findIndex(i => i.id === updatedItem.id)
        if (index !== -1) {
            state.items[index] = updatedItem
        }
    }

    const removeItem = (id: string) => {
        state.items = state.items.filter(i => i.id !== id)
    }

    return {
        ...toRefs(state),
        fetchVaults,
        addItem,
        updateItem,
        removeItem,
    }
}
