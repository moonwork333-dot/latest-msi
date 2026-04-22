import { NextRequest, NextResponse } from 'next/server';
import * as fs from 'fs';
import * as path from 'path';
import * as crypto from 'crypto';

export async function POST(request: NextRequest) {
  try {
    const { installPath } = await request.json();

    if (!installPath) {
      return NextResponse.json({ error: 'Installation path required' }, { status: 400 });
    }

    // Read environment variables
    const serverUrl = process.env.AGENT_SERVER_URL || 'ws://localhost:3001/hub';
    const authToken = process.env.AGENT_AUTH_TOKEN || 'test-token';

    // Generate machine ID
    const machineId = crypto.randomBytes(8).toString('hex').toUpperCase();

    // Create configuration
    const config = {
      machineId,
      serverUrl,
      authToken,
      installPath,
      createdAt: new Date().toISOString(),
      autoStart: true,
    };

    // Write config file
    const configPath = path.join(installPath, 'config', 'agent-config.json');
    fs.writeFileSync(configPath, JSON.stringify(config, null, 2));

    return NextResponse.json({
      success: true,
      message: 'Configuration created',
      machineId,
      configPath,
    });
  } catch (error) {
    return NextResponse.json({ error: String(error) }, { status: 500 });
  }
}
