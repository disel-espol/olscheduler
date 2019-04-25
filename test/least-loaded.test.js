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

  it('should use a different worker node if the package lists are the same', async () => {
    const req = { name: 'bar' };
    const responses = await Promise.all([
      client.sendRequest(req),
      waitMilis(30).then(() => client.sendRequest(req)),
      waitMilis(60).then(() => client.sendRequest(req)),
      waitMilis(90).then(() => client.sendRequest(req)),
    ]);
    const responseTexts = responses.map(res => res.text)

    expect(responseTexts).toEqual([
      "Request handled by worker at 9041",
      "Request handled by worker at 9042",
      "Request handled by worker at 9043",
      "Request handled by worker at 9044"
    ]);
  });
});


