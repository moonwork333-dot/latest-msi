declare module 'node-windows' {
  export interface ServiceConfig {
    name: string;
    description: string;
    script: string;
    nodeArgs?: string[];
    env?: Array<{ name: string; value: string }>;
    execPath?: string;
    maxRetries?: number;
    [key: string]: any;
  }

  export class Service {
    constructor(config: ServiceConfig);
    install(): void;
    uninstall(): void;
    start(): void;
    stop(): void;
    on(event: string, listener: (...args: any[]) => void): this;
  }
}
