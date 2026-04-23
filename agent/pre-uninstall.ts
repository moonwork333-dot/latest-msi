import { execSync } from 'child_process';

const SERVICE_NAME = 'RemoteAgent';

function log(msg: string) {
  const timestamp = new Date().toISOString();
  process.stdout.write(`[${timestamp}] ${msg}\n`);
}

function stopAndRemoveService() {
  // Stop service
  try {
    log(`Stopping service "${SERVICE_NAME}"...`);
    execSync(`sc stop "${SERVICE_NAME}"`, { encoding: 'utf-8' });
    // Give it a moment to stop cleanly
    execSync('timeout /t 3 /nobreak', { stdio: 'ignore' });
    log('Service stopped');
  } catch (_) {
    log('Service was not running or already stopped');
  }

  // Delete service
  try {
    log(`Removing service "${SERVICE_NAME}"...`);
    execSync(`sc delete "${SERVICE_NAME}"`, { encoding: 'utf-8' });
    log('Service removed');
  } catch (error) {
    log(`Could not remove service (may already be removed): ${error}`);
  }
}

// Main
try {
  log('=== Remote Agent Pre-Uninstall ===');
  stopAndRemoveService();
  log('=== Pre-uninstall complete ===');
  process.exit(0);
} catch (error) {
  log(`FATAL: ${error}`);
  // Exit 0 even on error so uninstall is not blocked
  process.exit(0);
}
