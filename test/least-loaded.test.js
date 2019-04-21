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
      port: 9020,
      workers: [9021, 9022, 9023, 9024],
      workerDelay: 1
    })    
    client = createClient(9020)
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
      "Request handled by worker at 9021",
      "Request handled by worker at 9022",
      "Request handled by worker at 9023",
      "Request handled by worker at 9024"
    ]);
  });
});


