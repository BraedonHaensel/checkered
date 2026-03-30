"""
Test script for managing multiple processes
"""

#pylint: disable=missing-function-docstring, missing-class-docstring

import os
import subprocess as sp
import signal
import threading
from typing import Callable
import time

from dotenv import load_dotenv

load_dotenv()
NAMESERVER_ADDRESS = os.getenv("APP_NAMESERVER_URL", "http://localhost:7000")
BASE_MATCHMAKER_PORT = int(os.getenv("APP_MATCHMAKER_BIND", ":2000").lstrip(':'))
BASE_GAMESERVER_PORT = int(os.getenv("APP_GAMESERVER_BIND", ":2000").lstrip(':'))

HELP_MESSAGE = (
    'Commands:\n'
    '- help:\n'
    '\t Display this message\n'
    '- quit:\n'
    '\t Shutdown all servers and exit\n'
    '- start <"matchmaker" | "gameserver">:\n'
    '\t Spawn a new matchmaker or gameserver\n'
    '- stop <"matchmaker" | "gameserver"> <id>:\n'
    '\t Stop the matchmaker or gameserver with the specified id\n'
    '- kill <"matchmaker" | "gameserver"> <id>:\n'
    '\t Kill the matchmaker or gameserver with the specified id\n'
    '- list <"matchmaker" | "gameserver">:\n'
    '\t List all active matchmakers or gameservers\n'
)

NPM_CMD = 'npm' if os.name != 'nt' else 'C:\\Program Files\\nodejs\\npm.cmd'

class CheckeredProcess:
    def __init__(
        self,
        pid: str,
        cmd: list[str],
        cwd: str | None = None,
        env: dict[str, str] | None = None,
        callback: Callable[[], None] | None = None
    ) -> None:
        self.pid = pid
        self.cmd = cmd
        self.cwd = cwd
        self.env = env
        self.proc: sp.Popen | None = None
        self.read_thread: threading.Thread | None = None
        self.callback = callback

    def _read_thread(self) -> None:
        for line in self.proc.stdout:
            print(f'[{self.pid}] {line.rstrip()}')
        if self.proc.poll() is None:
            self.proc.wait()
        print(f'[{self.pid}] Terminated')
        if self.callback is not None:
            self.callback()

    def start(self) -> 'CheckeredProcess':
        if self.proc is not None:
            raise ValueError('Process already started')
        popen_kwargs: dict[str, object] = {}
        if os.name == 'nt':
            # Required so Windows console control events target only the child process group.
            popen_kwargs['creationflags'] = sp.CREATE_NEW_PROCESS_GROUP
        self.proc = sp.Popen(
            args=self.cmd,
            text=True,
            stdin=sp.DEVNULL,
            stdout=sp.PIPE,
            stderr=sp.STDOUT,
            cwd=self.cwd,
            env=self.env,
            bufsize=1,
            **popen_kwargs,
        )
        self.read_thread = threading.Thread(
            target=self._read_thread,
            daemon=True,
        )
        self.read_thread.start()
        return self

    def kill(self) -> None:
        if self.proc is None:
            return
        if os.name != 'nt':
            self.proc.kill()
        else:
            os.system(f'TASKKILL /F /T /PID {self.proc.pid}')
        if self.read_thread is not None:
            self.read_thread.join()
        self.proc = None
        self.read_thread = None

    def stop(self) -> None:
        if self.proc is None:
            return
        if self.proc.poll() is None:
            if os.name != 'nt':
                self.proc.send_signal(signal.SIGINT)
            else:
                # Avoid CTRL_C_EVENT because it can interrupt this Python process too.
                self.proc.send_signal(signal.CTRL_BREAK_EVENT)

            try:
                self.proc.wait(timeout=5)
            except sp.TimeoutExpired:
                self.proc.terminate()
                self.proc.wait(timeout=5)

        if self.read_thread is not None:
            self.read_thread.join()
        self.proc = None
        self.read_thread = None

def main() -> None:
    name_server: CheckeredProcess = None
    frontend_server: CheckeredProcess = None
    matchmakers: dict[str, CheckeredProcess] = {}
    game_servers: dict[str, CheckeredProcess] = {}

    matchmaker_id = 0
    game_server_id = 0

    # Start name server
    try:
        name_server = CheckeredProcess(
            pid='Name Server',
            cmd=['go', 'run', '.'],
            cwd='name-server',
            env=os.environ
        )
        name_server.start()

        # Start frontend server
        frontend_server = CheckeredProcess(
            pid='Front-end Server',
            cmd=[NPM_CMD, 'run', 'dev'],
            cwd='frontend',
            env={**os.environ}
        )
        frontend_server.start()

        while True:
            cmd = input()
            if not cmd:
                continue
            cmd, *args = cmd.split()
            if cmd.lower() == 'quit':
                break
            elif cmd.lower() == 'help':
                print(HELP_MESSAGE)
            elif cmd.lower() == 'list':
                if len(args) != 1:
                    print(f'Improper usage: expected 1 argument found {len(args)}')
                    continue
                server_type = args[0]
                if server_type.lower() == 'matchmaker':
                    print('Matchmakers:')
                    for matchmaker in matchmakers:
                        print(f'\t- {matchmaker}')
                elif server_type.lower() == 'gameserver':
                    print('Game Servers:')
                    for game_server in game_servers:
                        print(f'\t- {game_server}')
                else:
                    print(
                        f'Unexpected server type "{server_type}": '
                        'expected "matchmaker" or "gameserver"'
                    )
            elif cmd.lower() == 'stop':
                if len(args) != 2:
                    print(f'Improper usage: expected 2 arguments found {len(args)}')
                    continue
                server_type, pid = args
                if server_type.lower() == 'matchmaker':
                    if pid not in matchmakers:
                        print(f'Matchmaker "{pid}" not found')
                        continue
                    print(f'Stopping matchmaker "{pid}"')
                    matchmakers[pid].stop()
                    print(f'Matchmaker "{pid}" stopped')
                elif server_type.lower() == 'gameserver':
                    if pid not in game_servers:
                        print(f'Game Server "{pid}" not found')
                        continue
                    print(f'Stopping game server "{pid}"')
                    game_servers[pid].stop()
                    print(f'Game server "{pid}" stopped')
                else:
                    print(
                        f'Unexpected server type "{server_type}": '
                        'expected "matchmaker" or "gameserver"'
                    )
            elif cmd.lower() == 'kill':
                if len(args) != 2:
                    print(f'Improper usage: expected 2 arguments found {len(args)}')
                server_type, pid = args
                if server_type.lower() == 'matchmaker':
                    if pid not in matchmakers:
                        print(f'Matchmaker "{pid}" not found')
                        continue
                    print(f'Killing matchmaker "{pid}"')
                    matchmakers[pid].kill()
                    print(f'Matchmaker "{pid}" killed')
                elif server_type.lower() == 'gameserver':
                    if pid not in game_servers:
                        print(f'Game Server "{pid}" not found')
                        continue
                    print(f'Killing game server "{pid}"')
                    game_servers[pid].kill()
                    print(f'Game server "{pid}" killed')
                else:
                    print(
                        f'Unexpected server type "{server_type}": '
                        'expected "matchmaker" or "gameserver"'
                    )
            elif cmd.lower() == 'start':
                if len(args) != 1:
                    print(f'Improper usage: expected 1 argument found {len(args)}')
                server_type = args[0]
                if server_type.lower() == 'matchmaker':
                    pid = f'MM{matchmaker_id:02}'
                    port = BASE_MATCHMAKER_PORT + matchmaker_id
                    matchmaker = CheckeredProcess(
                        pid=pid,
                        cmd=['go', 'run', 'cmd/matchmaker/main.go', f'--addr=localhost:{port}'],
                        cwd='backend',
                        callback = lambda: matchmakers.pop(pid, None)
                    )
                    matchmaker_id += 1
                    matchmakers[pid] = matchmaker
                    matchmaker.start()
                    print(f'Created matchmaker with pid: "{pid}"')
                elif server_type.lower() == 'gameserver':
                    pid = f'GS{game_server_id:02}'
                    port = BASE_GAMESERVER_PORT + game_server_id
                    game_server = CheckeredProcess(
                        pid=pid,
                        cmd=['go', 'run', 'cmd/game-server/main.go', f'--addr=localhost:{port}'],
                        cwd='backend',
                        callback = lambda: game_servers.pop(pid, None)
                    )
                    game_server_id += 1
                    game_servers[pid] = game_server
                    game_server.start()
                    print(f'Created gameserver with pid: "{pid}"')
            else:
                print(f'Unrecognized command "{cmd}"')
    finally:
        if name_server is not None:
            name_server.stop()
        if frontend_server is not None:
            frontend_server.stop()
        for matchmaker in list(matchmakers.values()):
            matchmaker.stop()
        for game_server in list(game_servers.values()):
            game_server.stop()

if __name__ == '__main__':
    main()
