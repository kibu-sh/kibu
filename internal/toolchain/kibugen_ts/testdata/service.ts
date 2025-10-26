import type { HTTPClient } from './client'

type CheckStatusRequest = {
  status: string
}

type CheckStatusResponse = {
  status: string
}

export function createService(c: HTTPClient) {
  return {
    async checkStatus(req: CheckStatusRequest): Promise<CheckStatusResponse> {
      return c.request({
        method: 'GET',
        pathname: '/status',
        data: req,
      })
    },
  }
}
