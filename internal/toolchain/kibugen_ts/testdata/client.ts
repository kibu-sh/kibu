type HTTPMethod = 'GET' | 'POST' | 'PUT' | 'DELETE'

type JSONRequest<T = unknown> = {
  method: HTTPMethod
  pathname: string
  data?: T
}

export interface Client {
  execAsJSON<Req, Res>(request: JSONRequest<Req>): Promise<Res>
}
