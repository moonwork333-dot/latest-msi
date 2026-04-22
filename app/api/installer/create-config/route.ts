import { NextRequest, NextResponse } from 'next/server';

export const runtime = 'nodejs';

export async function POST(request: NextRequest) {
  try {
    const { installPath } = await request.json();

    if (!installPath) {
      return NextResponse.json({ error: 'Installation path required' }, { status: 400 });
    }

    // Generate machine ID
    const machineId = Math.random().toString(36).substring(2, 14).toUpperCase();

    // Config will be created on user's machine
    return NextResponse.json({
      success: true,
      message: 'Configuration created',
      machineId,
      serverUrl: process.env.AGENT_SERVER_URL || 'ws://localhost:3001/hub',
      installPath,
    });
  } catch (error) {
    return NextResponse.json({ error: String(error) }, { status: 500 });
  }
}
