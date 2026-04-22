import { NextRequest, NextResponse } from 'next/server';
import * as fs from 'fs';
import * as path from 'path';

export async function POST(request: NextRequest) {
  try {
    const { installPath } = await request.json();

    if (!installPath) {
      return NextResponse.json({ error: 'Installation path required' }, { status: 400 });
    }

    // Copy agent files from dist to installation path
    const distPath = path.join(process.cwd(), 'dist', 'agent');
    const binPath = path.join(installPath, 'bin');

    if (fs.existsSync(distPath)) {
      const files = fs.readdirSync(distPath);
      for (const file of files) {
        const src = path.join(distPath, file);
        const dest = path.join(binPath, file);
        fs.copyFileSync(src, dest);
      }
    }

    return NextResponse.json({
      success: true,
      message: 'Files copied successfully',
    });
  } catch (error) {
    return NextResponse.json({ error: String(error) }, { status: 500 });
  }
}
