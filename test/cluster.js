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
  // pretty print the JSON object
  const configText = JSON.stringify(obj, null, 4)

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
        'z17922!',
        'pkg2'
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
  const name = overridenOpts.name || overridenOpts.balancer
  const filePath = `/tmp/olscheduler-${name}.json`
  return writeJSONFile(filePath, { ...baseConfig, ...overridenOpts })
}

const spawnOlschedulerProcess = async overridenOpts => {
  const configPath = await createOlschedulerConfig(overridenOpts)
  const cp = spawn(OL_BIN, ['start', '-c', configPath])
  if (process.env.DEBUG) 
    cp.stderr.on('data', data => console.log('[OLS]: ' + data.toString()));
  return { cp, configPath }
}

const spawnWorkerProcess = (delay, port) => {
  const cp = spawn('node', [WORKER_JS, delay.toString(), port])
  return cp
}

const wait = seconds => new Promise((resolve, reject) => {
  setTimeout(() => resolve(), seconds * 1000)
})

const spawnCluster = async opts => {
  const { workerDelay, ...overridenOpts } = opts;
  const workerProcesses = opts.workers
    .map(workerUrlString => new URL(workerUrlString).port)
    .map(workerPort => spawnWorkerProcess(workerDelay || '0', workerPort))

  const { 
    cp: olProcess, 
    configPath 
  } = await spawnOlschedulerProcess(opts)

  // wait 2 seconds for the server to launch
  await wait(2)

  return {
    configPath,
    kill: () => {
      olProcess.kill()
      workerProcesses.forEach(w => w.kill())
    },
    addWorkers: newWorkers => {
      workerProcesses.push.apply(workerProcesses, newWorkers)
    }
  }
}

module.exports = { 
  spawnCluster, 
  spawnWorkerProcess,
  wait
}
