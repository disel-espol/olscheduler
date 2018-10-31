const cp = require('child_process')
const path = require('path')
const { promisify } = require('util')
const exec = promisify(cp.exec)

const OL_BIN = require.resolve('../bin/olscheduler')

const addWorkers = (configPath, urls) => {
  const args = urls.join(' ')
  return exec(`${OL_BIN} workers add -c ${configPath} ${args}`)
}

const removeWorkers = (configPath, urls) => {
  const args = urls.join(' ')
  return exec(`${OL_BIN} workers remove -c ${configPath} ${args}`)
}

module.exports = { addWorkers, removeWorkers }
