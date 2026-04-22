import { NextRequest, NextResponse } from 'next/server';
import * as fs from 'fs';
import * as path from 'path';

export const runtime = 'nodejs';

export async function POST(request: NextRequest) {
  try {
    const { installPath } = await request.json();

    if (!installPath) {
      return NextResponse.json({ error: 'Installation path required' }, { status: 400 });
    }

    // Create main directories
    const dirs = [
      path.join(installPath, 'bin'),
      path.join(installPath, 'config'),
      path.join(installPath, 'logs'),
      path.join(installPath, 'data'),
    ];

    for (const dir of dirs) {
      if (!fs.existsSync(dir)) {
        fs.mkdirSync(dir, { recursive: true });
      }
    }

    return NextResponse.json({ success: true, message: 'Directories created' });
  } catch (error) {
    return NextResponse.json({ error: String(error) }, { status: 500 });
  }
}
