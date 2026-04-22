import RemoteAgent from './remote-agent';

const config = {
  serverUrl: process.env.AGENT_SERVER_URL || 'ws://localhost:3001/hub',
  authToken: process.env.AGENT_AUTH_TOKEN || 'test-token',
  machineId: process.env.MACHINE_ID || 'default-machine-id',
};

const agent = new RemoteAgent(config);

agent.connect().catch((error) => {
  console.error('Failed to connect agent:', error);
  process.exit(1);
});

// Handle graceful shutdown
process.on('SIGINT', async () => {
  console.log('Shutting down agent...');
  await agent.disconnect();
  process.exit(0);
});

process.on('SIGTERM', async () => {
  console.log('Shutting down agent...');
  await agent.disconnect();
  process.exit(0);
});
