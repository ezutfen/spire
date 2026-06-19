interface RuntimeEnvShape {
  VITE_BACKEND_BASE_URL?: string
  VUE_APP_BACKEND_BASE_URL?: string
}

function getRuntimeEnv(): RuntimeEnvShape {
  return import.meta.env as RuntimeEnvShape
}

export const RuntimeEnv = {
  getBackendBaseUrl(): string {
    const env = getRuntimeEnv()
    return env.VITE_BACKEND_BASE_URL || env.VUE_APP_BACKEND_BASE_URL || window.location.origin
  },
}
