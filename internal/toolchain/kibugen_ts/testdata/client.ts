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
