import { NextResponse} from 'next/server'
import { join } from "node:path";
import {readFileSync} from "node:fs";

export async function GET() {
    try {
        const packageJsonPath = join(process.cwd(), 'package.json');
        const packageJson = JSON.parse(readFileSync(packageJsonPath, 'utf8'));
        const version = packageJson.version;

        const info = {
            version: version
        }
        return NextResponse.json({info}, {status: 200})
    } catch (error) {
        console.error(error, __filename)
        return NextResponse.json(
            {error: 'Internal server error'},
            {status: 500}
        )
    }
}
