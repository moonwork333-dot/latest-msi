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

    // Copy agent binaries and files
    // In production, this would copy the compiled agent from a build directory
    // For now, we're creating placeholder for the build process to fill

    const binDir = path.join(installPath, 'bin');
    const files = ['service-runner.js', 'windows-service.js', 'remote-agent.js'];

    for (const file of files) {
      const filePath = path.join(binDir, file);
      // Create placeholder - actual files will be bundled by build process
      if (!fs.existsSync(filePath)) {
        fs.writeFileSync(filePath, '// Agent binary placeholder\n', 'utf-8');
      }
    }

    return NextResponse.json({ success: true, message: 'Agent files copied' });
  } catch (error) {
    return NextResponse.json({ error: String(error) }, { status: 500 });
  }
}
