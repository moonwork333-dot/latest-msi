import * as Service from 'node-windows';
import * as path from 'path';

// Get config path from environment or default
const installPath = process.env.AGENT_INSTALL_PATH || 'C:\\Program Files\\RemoteAgent';
const configPath = path.join(installPath, 'config', 'agent-config.json');

// Create service
const svc = new Service.Service({
  name: 'RemoteAgent',
  description: 'Remote Agent Service',
  script: path.join(__dirname, 'service-runner.js'),
  env: [
    {
      name: 'AGENT_CONFIG_PATH',
      value: configPath,
    },
    {
      name: 'NODE_ENV',
      value: 'production',
    },
  ],
  execPath: process.execPath,
  maxRetries: 3,
});

// Listen for the "install" event
svc.on('install', () => {
  console.log('Service installed');
  svc.start();
});

// Listen for the "start" event
svc.on('start', () => {
  console.log('Service started');
});

// Listen for the "error" event
svc.on('error', (error: Error) => {
  console.error('Service error:', error);
});

// Check if it's an install request
if (process.argv[2] === 'install') {
  svc.install();
} else if (process.argv[2] === 'uninstall') {
  svc.uninstall();
} else if (process.argv[2] === 'start') {
  svc.start();
} else if (process.argv[2] === 'stop') {
  svc.stop();
} else {
  console.log('Usage: node windows-service.ts [install|uninstall|start|stop]');
}
