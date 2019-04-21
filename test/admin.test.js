const { spawnCluster, spawnWorkerProcess, wait } = require('./cluster.js')
const { createClient } = require('./client.js')
const cli = require('./cli.js')

let client, cluster;

describe('HTTP endpoints for admin', () => {
  beforeAll(async () => {
    cluster = await spawnCluster({
      balancer: 'pkg-aware',
      name: 'admin',
      port: 9030,
      workers: []
    })    
    client = createClient(9030)
  })

  afterAll(() => {
    cluster.kill();
  })

  it('can add and remove workers', async () => {
    const res1 = await client.addWorkers(['localhost:9031', 'localhost:9032', 'localhost:9033'])
    expect(res1.status).toBe(200)

    const res2 = await client.removeWorkers(['localhost:9033'])
    expect(res2.status).toBe(200)
  })
})

describe('Admin CLI', () => {
  beforeAll(async () => {
    cluster = await spawnCluster({
      balancer: 'round-robin',
      name: 'admin',
      port: 9030,
      workers: []
    })    
    client = createClient(9030)
  })

  afterAll(() => {
    cluster.kill();
  })

  it('can add and remove workers', async () => {
    // exec admin operations
    const workers = [9033, 9034].map(port => spawnWorkerProcess('0', port))
    await wait(1)
    cluster.addWorkers(workers);

    await cli.addWorkers(
      cluster.configPath, 
      [
        'localhost:9031', 
        'localhost:9032', 
        'localhost:9033', 
        'localhost:9034'
      ]
    )
    const { stdout } = await cli.removeWorkers(
      cluster.configPath, 
      [
        'localhost:9031', 
        'localhost:9032'
      ]
    )
    // now run the workers
    const requests = new Array(4).fill({ name: 'foo' })
    const responses = await client.sendRequestsSequentially(requests)
    const responseTexts = responses.map(res => res.text)

    expect(responseTexts).toEqual([
      "Request handled by worker at 9033",
      "Request handled by worker at 9034",
      "Request handled by worker at 9033",
      "Request handled by worker at 9034",
    ])
  })
})
