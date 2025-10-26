type HTTPMethod = 'GET' | 'POST' | 'PUT' | 'DELETE'

type HTTPRequest<T = unknown> = {
  method: HTTPMethod
  pathname: string
  searchParams?: URLSearchParams
  headers?: Headers
  data?: T
}

export interface HTTPClient {
  request<Req, Res>(request: HTTPRequest<Req>): Promise<Res>
}

export type FetchClientParams = {
  baseUrl: string
  headers?: Headers
  timeout?: number
}

export function createFetchClient(init: FetchClientParams): HTTPClient {
  return {
    async request<ReqT, ResT>(request: HTTPRequest<ReqT>): Promise<ResT> {
      const url = new URL(request.pathname, init.baseUrl)
      const headers = new Headers(init.headers)

      request.searchParams?.forEach((value, key) => {
        url.searchParams.append(key, value)
      })

      request.headers?.forEach((value, key) => {
        headers.append(key, value)
      })

      const res = await fetch(url.toString(), {
        method: request.method,
        headers: headers,
      })

      const data = await res.json()
      return data as ResT
    },
  }
}
