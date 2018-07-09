const request = require('superagent')

const createClient = port => {
  const sendRequest = (overridenAttrs = {}) =>
    request.post(`localhost:${port}/runLambda/foo`)
      .ok(res => res.status)
      .send({ pkgs: ['pkg0', 'pkg1'], param0: 'value0', ...overridenAttrs })

  const sendRequestsSequentially = async overrideAttrsArray => {
      const results = []
      for (let i = 0; i < overrideAttrsArray.length; i++) {
        const res = await sendRequest(overrideAttrsArray[i])
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
