import { NextRequest, NextResponse } from 'next/server';
import * as path from 'path';
import { spawn } from 'child_process';

export async function POST(request: NextRequest) {
  try {
    const { installPath } = await request.json();

    if (!installPath) {
      return NextResponse.json({ error: 'Installation path required' }, { status: 400 });
    }

    // Start the Windows service
    const serviceFile = path.join(installPath, 'bin', 'windows-service.js');

    return new Promise((resolve, reject) => {
      const proc = spawn('node', [serviceFile, 'start'], {
        cwd: installPath,
        stdio: 'pipe',
      });

      let output = '';
      let errorOutput = '';

      proc.stdout?.on('data', (data) => {
        output += data.toString();
      });

      proc.stderr?.on('data', (data) => {
        errorOutput += data.toString();
      });

      proc.on('close', (code) => {
        if (code === 0) {
          resolve(
            NextResponse.json({
              success: true,
              message: 'Service started successfully',
              output,
            })
          );
        } else {
          reject(new Error(`Failed to start service: ${errorOutput}`));
        }
      });

      proc.on('error', (error) => {
        reject(error);
      });
    });
  } catch (error) {
    return NextResponse.json({ error: String(error) }, { status: 500 });
  }
}
