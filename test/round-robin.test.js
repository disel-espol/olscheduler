const { spawnCluster } = require('./cluster.js')
const { createClient } = require('./client.js')

let client, cluster;


describe('round-robin balancer', () => {

  beforeAll(async () => {
    cluster = await spawnCluster({
      balancer: 'round-robin',
      port: 9010,
      workers: [9011, 9012, 9013, 9014]
    })    
    client = createClient(9010)
  })

  afterAll(() => {
    cluster.kill();
  })

  it('should distribute load evenly between all workers', async () => {
    const requests = new Array(8).fill({ name: 'foo' })
    const responses = await client.sendRequestsSequentially(requests)
    const responseTexts = responses.map(res => res.text)

    expect(responseTexts).toEqual([
      "Request handled by worker at 9011",
      "Request handled by worker at 9012",
      "Request handled by worker at 9013",
      "Request handled by worker at 9014",
      "Request handled by worker at 9011",
      "Request handled by worker at 9012",
      "Request handled by worker at 9013",
      "Request handled by worker at 9014"
    ])
  })
})
