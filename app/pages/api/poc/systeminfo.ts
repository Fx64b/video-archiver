// Next.js API route support: https://nextjs.org/docs/api-routes/introduction
import type { NextApiRequest, NextApiResponse } from 'next'
import { CommandResult, executeCommand } from '@/utils/command';

export default async function handler(
    req: NextApiRequest,
    res: NextApiResponse<CommandResult>
) {
    try {
        const commandResult = await executeCommand('lscpu');
        console.log(commandResult);

        res.status(200).json(commandResult);
    } catch (error) {
        console.error(`Unexpected error: ${error}`);
        res.status(500).json({ error: 'Internal Server Error' });
    }
}