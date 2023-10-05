import { exec } from 'child_process';

export type CommandResult = {
    data?: any;
    error?: string;
};

export function executeCommand(command: string): Promise<CommandResult> {
    return new Promise((resolve) => {
        exec(command, (error, stdout, stderr) => {
            if (error || stderr) {
                const errorMessage = error ? error.message : stderr;
                resolve({ error: `Command execution failed: ${errorMessage}` });
            } else {
                resolve({ data: stdout });
            }
        });
    });
}
