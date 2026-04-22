import { NextRequest, NextResponse } from 'next/server';

export const runtime = 'nodejs';

export async function POST(request: NextRequest) {
  try {
    const { installPath } = await request.json();

    if (!installPath) {
      return NextResponse.json({ error: 'Installation path required' }, { status: 400 });
    }

    // Files are copied by the MSI installer on user's machine
    return NextResponse.json({
      success: true,
      message: 'Agent files copied',
      installPath,
    });
  } catch (error) {
    return NextResponse.json({ error: String(error) }, { status: 500 });
  }
}
