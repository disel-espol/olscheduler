const { spawnCluster, spawnWorkerProcess, wait } = require('./cluster.js')
const { createClient } = require('./client.js')
const cli = require('./cli.js')

let client, cluster;

describe('HTTP endpoints for admin', () => {
  beforeAll(async () => {
    cluster = await spawnCluster({
      balancer: 'pkg-aware',
      name: 'admin',
      port: 9080,
      workers: []
    })    
    client = createClient(9080)
  })

  afterAll(() => {
    cluster.kill();
  }) 
  it('can add and remove workers', async () => {
    const res1 = await client.addWorkers([
      'http://localhost:9081', 
      'http://localhost:9082', 
      'http://localhost:9083'
    ]);
    expect(res1.status).toBe(200);

    const res2 = await client.removeWorkers(['http://localhost:9083']);
    expect(res2.status).toBe(200);
  })
})

describe('Admin CLI', () => {
  beforeAll(async () => {
    cluster = await spawnCluster({
      balancer: 'round-robin',
      name: 'admin',
      port: 9080,
      workers: []
    })    
    client = createClient(9080)
  })

  afterAll(() => {
    cluster.kill();
  })

  it('can add and remove workers', async () => {
    // exec admin operations
    const workers = [9083, 9084].map(port => spawnWorkerProcess('0', port))
    await wait(1)
    cluster.addWorkers(workers);

    await cli.addWorkers(
      cluster.configPath, 
      [
        'http://localhost:9081', 
        'http://localhost:9082', 
        'http://localhost:9083', 
        'http://localhost:9084'
      ]
    )
    const { stdout } = await cli.removeWorkers(
      cluster.configPath, 
      [
        'http://localhost:9081', 
        'http://localhost:9082'
      ]
    )
    // now run the workers
    const requests = new Array(4).fill({ name: 'foo' })
    const responses = await Promise.all(requests.map(req => client.sendRequest(req)));
    const responseTexts = responses.map(res => res.text).sort()

    expect(responseTexts).toEqual([
      "Request handled by worker at 9083",
      "Request handled by worker at 9083",
      "Request handled by worker at 9084",
      "Request handled by worker at 9084",
    ])
  })
})
