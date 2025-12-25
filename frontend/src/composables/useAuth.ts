import { ref, onMounted } from 'vue'

export interface AuthState {
  authenticated: boolean
  username: string
  checking: boolean
}

export function useAuth() {
  const authState = ref<AuthState>({
    authenticated: false,
    username: '',
    checking: true
  })

  const checkAuthStatus = async () => {
    try {
      const res = await fetch('/auth/status')
      const data = await res.json()
      authState.value = {
        authenticated: data.authenticated,
        username: data.username || '',
        checking: false
      }
    } catch (error) {
      console.error('Auth check failed:', error)
      authState.value = {
        authenticated: false,
        username: '',
        checking: false
      }
    }
  }

  const login = async (username: string, password: string): Promise<{ success: boolean; message?: string }> => {
    const formData = new URLSearchParams()
    formData.append('username', username)
    formData.append('password', password)

    try {
      const res = await fetch('/login', {
        method: 'POST',
        headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
        body: formData
      })
      const data = await res.json()
      if (data.success) {
        authState.value = {
          authenticated: true,
          username: data.username,
          checking: false
        }
      }
      return data
    } catch (error) {
      return { success: false, message: 'Network error' }
    }
  }

  const logout = async () => {
    try {
      await fetch('/logout', { method: 'POST' })
      authState.value = {
        authenticated: false,
        username: '',
        checking: false
      }
    } catch (error) {
      console.error('Logout failed:', error)
    }
  }

  onMounted(() => {
    checkAuthStatus()
  })

  return {
    authState,
    login,
    logout
  }
}

