declare module 'ws' {
  export interface ClientOptions {
    accessTokenFactory?: () => string;
    [key: string]: any;
  }

  export class WebSocket {
    constructor(url: string, options?: ClientOptions);
    on(event: string, listener: (...args: any[]) => void): this;
    send(data: string | Buffer, callback?: (err?: Error) => void): void;
    close(): void;
  }
}
