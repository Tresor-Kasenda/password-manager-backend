import { reactive, toRefs } from 'vue'

interface AuthState {
  isAuthenticated: boolean
  user: any | null
  token: string | null
}

const state = reactive<AuthState>({
  isAuthenticated: !!localStorage.getItem('token'),
  user: JSON.parse(localStorage.getItem('user') || 'null'),
  token: localStorage.getItem('token'),
})

export const useAuthStore = () => {
  const login = (userData: any, token: string) => {
    state.isAuthenticated = true
    state.user = userData
    state.token = token
    localStorage.setItem('token', token)
    localStorage.setItem('user', JSON.stringify(userData))
  }

  const logout = () => {
    state.isAuthenticated = false
    state.user = null
    state.token = null
    localStorage.removeItem('token')
    localStorage.removeItem('user')
  }

  return {
    ...toRefs(state),
    login,
    logout,
  }
}
