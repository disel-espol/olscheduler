const { spawnCluster } = require('./cluster.js')
const { createClient } = require('./client.js')

let client, cluster;

const waitMilis = milis => new Promise((resolve, reject) => {
  setTimeout(() => resolve(), milis)
})

describe('least-loaded balancer', () => {

  beforeAll(async () => {
    cluster = await spawnCluster({
      balancer: 'least-loaded',
      port: 9040,
      workers: [
        'http://localhost:9041', 
        'http://localhost:9042', 
        'http://localhost:9043', 
        'http://localhost:9044'
      ],
      workerDelay: 1
    })    
    client = createClient(9040)
  })

  afterAll(() => {
    cluster.kill();
  })

  it('4 simultaneous requests should use 4 different workers', async () => {
    const req = { name: 'bar' };
    const responses = await Promise.all([
      client.sendRequest(req),
      client.sendRequest(req),
      client.sendRequest(req),
      client.sendRequest(req),
    ]);
    const responseTexts = responses.map(res => res.text).sort()

    expect(responseTexts).toEqual([
      "Request handled by worker at 9041",
      "Request handled by worker at 9042",
      "Request handled by worker at 9043",
      "Request handled by worker at 9044"
    ]);
  });
});


