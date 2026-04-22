import { NextRequest, NextResponse } from 'next/server';
import * as fs from 'fs';
import * as path from 'path';

export async function POST(request: NextRequest) {
  try {
    const { installPath } = await request.json();

    if (!installPath) {
      return NextResponse.json({ error: 'Installation path required' }, { status: 400 });
    }

    // Create main installation directory
    if (!fs.existsSync(installPath)) {
      fs.mkdirSync(installPath, { recursive: true });
    }

    // Create subdirectories
    const subdirs = ['bin', 'config', 'logs', 'data'];
    for (const subdir of subdirs) {
      const dirPath = path.join(installPath, subdir);
      if (!fs.existsSync(dirPath)) {
        fs.mkdirSync(dirPath, { recursive: true });
      }
    }

    return NextResponse.json({
      success: true,
      message: 'Directories created successfully',
      installPath,
    });
  } catch (error) {
    return NextResponse.json({ error: String(error) }, { status: 500 });
  }
}
