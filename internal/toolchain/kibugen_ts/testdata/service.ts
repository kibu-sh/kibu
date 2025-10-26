import type { Client } from './client'

type CheckStatusRequest = {
  status: string
}

type CheckStatusResponse = {
  status: string
}

export function createService(c: Client) {
  return {
    async checkStatus(req: CheckStatusRequest): Promise<CheckStatusResponse> {
      return c.execAsJSON({
        method: 'GET',
        pathname: '/status',
        data: req,
      })
    },
  }
}
