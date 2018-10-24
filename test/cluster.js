const fs = require('fs')
const { promisify } = require('util') 
const { spawn } = require('child_process');

const writeFile = promisify(fs.writeFile)
const OL_BIN = require.resolve('../bin/olscheduler')
const WORKER_JS = require.resolve('./worker.js')

const abortOnErrorHandler = err => {
  if (err)
    console.error('Failed to initialize test cluster: ', err)
}

const writeJSONFile = (filePath, obj) => {
  const configText = JSON.stringify(obj)

  return writeFile(filePath, configText)
    .then(() => Promise.resolve(filePath))
}

const createRegistryConfig = () => {
  const entries = [
    {
      handle: 'foo',
      pkgs: [
        'pkg0',
        'pkg1'
      ]
    },
    {
      handle: 'bar',
      pkgs: [
        'pkg7',
        'pkg8'
      ]
    }
  ]  
  const filePath = '/tmp/olscheduler-registry.json';
  return writeJSONFile(filePath, entries);
}


const createOlschedulerConfig = async overridenOpts => {
  const baseConfig = {
    host: 'localhost',
    port: 9080,
    ['load-threshold']: 3,
    registry: await createRegistryConfig()
  }
  const filePath = `/tmp/olscheduler-${overridenOpts.balancer}.json`
  return writeJSONFile(filePath, { ...baseConfig, ...overridenOpts })
}

const spawnOlschedulerProcess = async overridenOpts => {
  const configPath = await createOlschedulerConfig(overridenOpts)
  const cp = spawn(OL_BIN, ['start', '-c', configPath])

  if (process.env.DEBUG) 
    cp.stderr.on('data', data => console.log('[OLS]: ' + data.toString()));

  return cp
}

const spawnWorkerProcess = workerPort => {
  const cp = spawn('node', [WORKER_JS, '0', workerPort])
  return cp
}

const wait = seconds => new Promise((resolve, reject) => {
  setTimeout(() => resolve(), seconds * 1000)
})

const createOlschedulerOverridenOpts = args => {
  const workers = args.workers
    .map((port) => ['localhost:' + port, "1"])
    .reduce((acc, array) => acc.concat(array), [])
  return { ...args, workers }
}

const spawnCluster = async opts => {
  const workerProcesses = opts.workers.map(workerPort => spawnWorkerProcess(workerPort))

  const overridenOpts = createOlschedulerOverridenOpts(opts)
  const olProcess = await spawnOlschedulerProcess(overridenOpts)

  await wait(1)

  return {
    kill: () => {
      olProcess.kill()
      workerProcesses.forEach(w => w.kill())
    }
  }
}

module.exports = { spawnCluster }
