export async function handleResult(resp: Response) {
  if (resp.status !== 200) {
    let reason = await resp.text()
    if (/json/i.test(resp.headers.get('Content-Type') || '')) {
      const data = JSON.parse(reason)
      reason = data.error || reason
    }
    if (!reason)
      reason = resp.statusText
    console.error(reason)
    return Promise.reject(reason)
  }
  return await resp.json()
}

export async function sendReq(method: string, url: RequestInfo, data?: any) {
  const resp = await fetch(url, {
    method,
    body: JSON.stringify(data || {}),
    headers: new Headers({
      'Content-Type': 'application/json',
    }),
  })
  return await handleResult(resp)
}

class Request {
  async delete(url: RequestInfo, data?: any) {
    return await sendReq('DELETE', url, data)
  }

  async put(url: RequestInfo, data?: any) {
    return await sendReq('PUT', url, data)
  }

  async patch(url: RequestInfo, data?: any) {
    return await sendReq('PATCH', url, data)
  }

  async post(url: RequestInfo, data?: any) {
    return await sendReq('POST', url, data)
  }

  async get(url: RequestInfo) {
    const resp = await fetch(url)
    return await handleResult(resp)
  }
}

const request = new Request()
export default request
