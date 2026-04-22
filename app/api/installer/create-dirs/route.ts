import { NextRequest, NextResponse } from 'next/server';

export const runtime = 'nodejs';

export async function POST(request: NextRequest) {
  try {
    const { installPath } = await request.json();

    if (!installPath) {
      return NextResponse.json({ error: 'Installation path required' }, { status: 400 });
    }

    // On user's machine, this will be handled by the installer script
    // This endpoint just acknowledges the request
    return NextResponse.json({
      success: true,
      message: 'Directory creation initiated',
      installPath,
    });
  } catch (error) {
    return NextResponse.json({ error: String(error) }, { status: 500 });
  }
}
