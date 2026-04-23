import * as fs from 'fs';
import * as path from 'path';
import { execSync } from 'child_process';

const SERVICE_NAME = 'RemoteAgent';
const SERVICE_DISPLAY_NAME = 'Remote Agent Service';
const SERVICE_DESCRIPTION = 'Remote control and monitoring agent';

function log(msg: string) {
  const timestamp = new Date().toISOString();
  const line = `[${timestamp}] ${msg}\n`;
  process.stdout.write(line);
  try {
    const logPath = path.join(getInstallDir(), 'logs', 'install.log');
    fs.appendFileSync(logPath, line);
  } catch (_) {}
}

function getInstallDir(): string {
  // When run as a CustomAction by MSI, the exe lives in the install folder
  return path.dirname(process.execPath);
}

function writeConfig() {
  const installDir = getInstallDir();
  const configDir = path.join(installDir, 'config');
  const configPath = path.join(configDir, 'agent-config.json');

  if (!fs.existsSync(configDir)) {
    fs.mkdirSync(configDir, { recursive: true });
  }

  // Read server URL and auth token from environment (set by MSI properties)
  // or fall back to defaults that the user can edit post-install
  const serverUrl = process.env.AGENT_SERVER_URL || 'wss://your-rmm-server.com/agent';
  const authToken = process.env.AGENT_AUTH_TOKEN || '';

  // Generate a unique machine ID
  const machineId = generateMachineId();

  const config = {
    machineId,
    serverUrl,
    authToken,
    installPath: installDir,
    reconnectInterval: 5000,
    screenshotQuality: 80,
    logLevel: 'info',
  };

  fs.writeFileSync(configPath, JSON.stringify(config, null, 2), 'utf-8');
  log(`Config written to: ${configPath}`);
  log(`Machine ID: ${machineId}`);
}

function generateMachineId(): string {
  try {
    // Use Windows hostname + MAC address for a stable unique ID
    const hostname = execSync('hostname', { encoding: 'utf-8' }).trim();
    const macOutput = execSync(
      'getmac /fo csv /nh',
      { encoding: 'utf-8' }
    ).trim();
    const mac = macOutput.split('\n')[0].replace(/"/g, '').split(',')[0].replace(/-/g, '');
    return `${hostname}-${mac}`.toLowerCase().replace(/[^a-z0-9-]/g, '-');
  } catch (_) {
    // Fallback to random ID
    return `agent-${Math.random().toString(36).substr(2, 12)}`;
  }
}

function registerService() {
  const installDir = getInstallDir();
  const exePath = path.join(installDir, 'RemoteAgent.exe');

  try {
    // Check if service already exists and remove it first
    try {
      execSync(`sc query "${SERVICE_NAME}"`, { stdio: 'ignore' });
      log('Existing service found, removing it first...');
      execSync(`sc stop "${SERVICE_NAME}"`, { stdio: 'ignore' });
      execSync(`sc delete "${SERVICE_NAME}"`, { stdio: 'ignore' });
      // Wait for deletion to complete
      execSync('timeout /t 2 /nobreak', { stdio: 'ignore' });
    } catch (_) {
      // Service doesn't exist yet, that's fine
    }

    // Create the Windows service
    execSync(
      `sc create "${SERVICE_NAME}" binPath= "${exePath}" start= auto DisplayName= "${SERVICE_DISPLAY_NAME}"`,
      { encoding: 'utf-8' }
    );
    log(`Service "${SERVICE_NAME}" created`);

    // Set description
    execSync(
      `sc description "${SERVICE_NAME}" "${SERVICE_DESCRIPTION}"`,
      { encoding: 'utf-8' }
    );

    // Configure service to restart automatically on failure
    execSync(
      `sc failure "${SERVICE_NAME}" reset= 60 actions= restart/5000/restart/10000/restart/30000`,
      { encoding: 'utf-8' }
    );
    log('Service failure recovery configured');

    // Start the service
    execSync(`sc start "${SERVICE_NAME}"`, { encoding: 'utf-8' });
    log(`Service "${SERVICE_NAME}" started successfully`);

  } catch (error) {
    log(`ERROR registering service: ${error}`);
    process.exit(1);
  }
}

function createLogsDir() {
  const logsDir = path.join(getInstallDir(), 'logs');
  if (!fs.existsSync(logsDir)) {
    fs.mkdirSync(logsDir, { recursive: true });
  }
}

// Main
try {
  log('=== Remote Agent Post-Install ===');
  createLogsDir();
  writeConfig();
  registerService();
  log('=== Post-install complete ===');
  process.exit(0);
} catch (error) {
  log(`FATAL: ${error}`);
  process.exit(1);
}
