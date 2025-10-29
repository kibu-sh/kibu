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
  integrity?: string
  keepalive?: boolean
  redirect?: 'error' | 'follow' | 'manual'
  credentials?: 'include' | 'omit' | 'same-origin'
  mode?: 'cors' | 'navigate' | 'no-cors' | 'same-origin'
  cache?:
    | 'default'
    | 'force-cache'
    | 'no-cache'
    | 'no-store'
    | 'only-if-cached'
    | 'reload'
  referrerPolicy?:
    | ''
    | 'no-referrer'
    | 'no-referrer-when-downgrade'
    | 'origin'
    | 'origin-when-cross-origin'
    | 'same-origin'
    | 'strict-origin'
    | 'strict-origin-when-cross-origin'
    | 'unsafe-url'
}

function cannotHaveBody(method: HTTPMethod) {
  return ['GET', 'HEAD'].includes(method)
}

function hasSuccessStatusCode(status: number) {
  return status >= 200 && status < 300
}

function asJSON(request: HTTPRequest) {
  if (cannotHaveBody(request.method)) return undefined
  return request.data ? JSON.stringify(request.data) : undefined
}

const applicationJSON = 'application/json'

export function createFetchClient(init: FetchClientParams): HTTPClient {
  return {
    async request<ReqT, ResT>(request: HTTPRequest<ReqT>): Promise<ResT> {
      const url = new URL(init.baseUrl)
      url.pathname = url.pathname + request.pathname
      url.pathname = url.pathname.replaceAll('//', '/')
      const headers = new Headers(init.headers)
      headers.set('Accept', applicationJSON)
      headers.set('Content-Type', applicationJSON)

      request.searchParams?.forEach((value, key) => {
        url.searchParams.append(key, value)
      })

      request.headers?.forEach((value, key) => {
        headers.append(key, value)
      })

      const res = await fetch(url.toString(), {
        method: request.method,
        headers: headers,
        credentials: init.credentials,
        cache: init.cache,
        mode: init.mode,
        keepalive: init.keepalive,
        integrity: init.integrity,
        redirect: init.redirect,
        referrerPolicy: init.referrerPolicy,
        body: asJSON(request),
      })

      const data = await res.json()

      if (!res.ok) {
        throw new Error(
          `${request.method} ${url} ${res.status} ${JSON.stringify(data)}`,
        )
      }

      return data as ResT
    },
  }
}
