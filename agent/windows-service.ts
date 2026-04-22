import * as Service from 'node-windows';
import * as path from 'path';

const svc = new Service.Service({
  name: 'RemoteAgent',
  description: 'Remote Agent Service for RMM',
  script: path.join(__dirname, 'service-runner.js'),
  nodeArgs: ['--harmony'],
  env: [
    {
      name: 'AGENT_SERVER_URL',
      value: process.env.AGENT_SERVER_URL || 'ws://localhost:3001/hub',
    },
    {
      name: 'AGENT_AUTH_TOKEN',
      value: process.env.AGENT_AUTH_TOKEN || 'test-token',
    },
  ],
});

// Listen for the "install" event, which fires if the installation was successful
svc.on('install', () => {
  console.log('Service installed successfully');
  svc.start();
});

svc.on('start', () => {
  console.log('Service started');
});

svc.on('stop', () => {
  console.log('Service stopped');
});

svc.on('uninstall', () => {
  console.log('Service uninstalled');
});

svc.on('error', (error) => {
  console.error('Service error:', error);
});

// Install the script as a windows service
svc.install();
