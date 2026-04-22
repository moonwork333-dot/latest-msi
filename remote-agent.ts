import * as WebSocket from 'ws';
import * as os from 'os';
import { HubConnection, HubConnectionBuilder, LogLevel } from '@microsoft/signalr';

interface AgentConfig {
  serverUrl: string;
  authToken: string;
  machineId: string;
}

class RemoteAgent {
  private config: AgentConfig;
  private connection: HubConnection | null = null;
  private isConnected = false;

  constructor(config: AgentConfig) {
    this.config = config;
  }

  async connect(): Promise<void> {
    try {
      console.log('[RemoteAgent] Connecting to SignalR hub:', this.config.serverUrl);

      this.connection = new HubConnectionBuilder()
        .withUrl(this.config.serverUrl, {
          accessTokenFactory: () => this.config.authToken,
        })
        .withAutomaticReconnect([0, 1000, 5000, 10000])
        .configureLogging(LogLevel.Information)
        .build();

      // Set up handlers
      this.connection.on('GetScreenCapture', () => this.handleScreenCapture());
      this.connection.on('ExecuteMouseControl', (data: any) => this.handleMouseControl(data));
      this.connection.on('ExecuteKeyboardControl', (data: any) => this.handleKeyboardControl(data));
      this.connection.on('GetSystemInfo', () => this.handleSystemInfo());
      this.connection.on('GetProcessList', () => this.handleProcessList());

      this.connection.onreconnecting(() => {
        console.log('[RemoteAgent] Reconnecting...');
        this.isConnected = false;
      });

      this.connection.onreconnected(() => {
        console.log('[RemoteAgent] Reconnected');
        this.isConnected = true;
        this.reportOnline();
      });

      await this.connection.start();
      this.isConnected = true;
      console.log('[RemoteAgent] Connected successfully');

      // Report online
      await this.reportOnline();

      // Start heartbeat
      this.startHeartbeat();
    } catch (error) {
      console.error('[RemoteAgent] Connection error:', error);
      throw error;
    }
  }

  private async reportOnline(): Promise<void> {
    try {
      await this.connection?.invoke('ReportMachineOnline', {
        machineId: this.config.machineId,
        hostname: os.hostname(),
        platform: os.platform(),
        arch: os.arch(),
        timestamp: new Date().toISOString(),
      });
    } catch (error) {
      console.error('[RemoteAgent] Failed to report online:', error);
    }
  }

  private startHeartbeat(): void {
    setInterval(async () => {
      if (this.isConnected && this.connection?.state === 'Connected') {
        try {
          await this.connection.invoke('Heartbeat', {
            machineId: this.config.machineId,
            timestamp: new Date().toISOString(),
          });
        } catch (error) {
          console.error('[RemoteAgent] Heartbeat error:', error);
        }
      }
    }, 30000); // Every 30 seconds
  }

  private async handleScreenCapture(): Promise<void> {
    try {
      console.log('[RemoteAgent] Screen capture requested');
      // Placeholder - will use Windows APIs for actual implementation
      await this.connection?.invoke('ScreenCaptureTaken', {
        machineId: this.config.machineId,
        data: 'screenshot-base64-data',
      });
    } catch (error) {
      console.error('[RemoteAgent] Screen capture error:', error);
    }
  }

  private async handleMouseControl(data: any): Promise<void> {
    try {
      console.log('[RemoteAgent] Mouse control:', data);
      // Placeholder - will use Windows APIs for actual implementation
    } catch (error) {
      console.error('[RemoteAgent] Mouse control error:', error);
    }
  }

  private async handleKeyboardControl(data: any): Promise<void> {
    try {
      console.log('[RemoteAgent] Keyboard control:', data);
      // Placeholder - will use Windows APIs for actual implementation
    } catch (error) {
      console.error('[RemoteAgent] Keyboard control error:', error);
    }
  }

  private async handleSystemInfo(): Promise<void> {
    try {
      const systemInfo = {
        machineId: this.config.machineId,
        hostname: os.hostname(),
        platform: os.platform(),
        arch: os.arch(),
        cpus: os.cpus().length,
        totalMemory: os.totalmem(),
        freeMemory: os.freemem(),
        uptime: os.uptime(),
      };
      await this.connection?.invoke('SystemInfoReported', systemInfo);
    } catch (error) {
      console.error('[RemoteAgent] System info error:', error);
    }
  }

  private async handleProcessList(): Promise<void> {
    try {
      console.log('[RemoteAgent] Process list requested');
      // Placeholder - will use Windows APIs to get process list
      await this.connection?.invoke('ProcessListReported', {
        machineId: this.config.machineId,
        processes: [],
      });
    } catch (error) {
      console.error('[RemoteAgent] Process list error:', error);
    }
  }

  async disconnect(): Promise<void> {
    if (this.connection) {
      await this.connection.stop();
      this.isConnected = false;
    }
  }
}

// Main entry point
async function main() {
  const config: AgentConfig = {
    serverUrl: process.env.AGENT_SERVER_URL || 'ws://localhost:3001/hub',
    authToken: process.env.AGENT_AUTH_TOKEN || 'test-token',
    machineId: process.env.MACHINE_ID || 'default-machine-id',
  };

  const agent = new RemoteAgent(config);

  try {
    await agent.connect();
    console.log('[RemoteAgent] Agent running...');
  } catch (error) {
    console.error('[RemoteAgent] Failed to start agent:', error);
    process.exit(1);
  }
}

if (require.main === module) {
  main();
}

export default RemoteAgent;
