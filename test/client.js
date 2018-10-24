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

  return {
    sendRequest,
    sendRequestsSequentially  
  }
}

module.exports = { createClient };
