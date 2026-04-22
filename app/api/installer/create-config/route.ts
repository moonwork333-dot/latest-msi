import { NextRequest, NextResponse } from 'next/server';
import * as fs from 'fs';
import * as path from 'path';
import * as crypto from 'crypto';

export const runtime = 'nodejs';

export async function POST(request: NextRequest) {
  try {
    const { installPath } = await request.json();

    if (!installPath) {
      return NextResponse.json({ error: 'Installation path required' }, { status: 400 });
    }

    // Generate unique machine ID
    const machineId = crypto.randomBytes(6).toString('hex').toUpperCase();

    // Read pre-configured environment variables
    const serverUrl = process.env.AGENT_SERVER_URL || '';
    const authToken = process.env.AGENT_AUTH_TOKEN || '';

    if (!serverUrl || !authToken) {
      return NextResponse.json(
        { error: 'Agent server configuration not found. Set AGENT_SERVER_URL and AGENT_AUTH_TOKEN environment variables.' },
        { status: 500 }
      );
    }

    const config = {
      machineId,
      serverUrl,
      authToken,
      installPath,
      createdAt: new Date().toISOString(),
    };

    const configPath = path.join(installPath, 'config', 'agent-config.json');
    fs.writeFileSync(configPath, JSON.stringify(config, null, 2), 'utf-8');

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
