const querystring = require('querystring')
const request = require('superagent')

const createClient = port => {
  const sendRequest = ({ name, ...overrideAttrs }) =>
    request.post(`localhost:${port}/runLambda/${name}`)
      .ok(res => res.status)
      .send({ param0: 'value0', ...(overrideAttrs || []) })

  const sendRequestsSequentially = async paramsArray => {
    const results = []
    for (let i = 0; i < paramsArray.length; i++) {
      const res = await sendRequest(paramsArray[i])
      results.push(res)
    }
    return results
  }

  const addWorkers = async workers => {
    const q = querystring.stringify({ workers })
    return request.post(`localhost:${port}/admin/workers/add?${q}`)
      .ok(res => res.status)
      .send()
  }

  const removeWorkers = async workers => {
    const q = querystring.stringify({ workers })
    return request.post(`localhost:${port}/admin/workers/remove?${q}`)
      .ok(res => res.status)
      .send()
  }
  return {
    addWorkers,
    removeWorkers,
    sendRequest,
    sendRequestsSequentially  
  }
}

module.exports = { createClient };
