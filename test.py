"""
Test script for managing multiple processes
"""

#pylint: disable=missing-function-docstring, missing-class-docstring

import logging
import os
import subprocess as sp
import signal
import sys
import threading
from typing import Callable

from dotenv import load_dotenv

logging.basicConfig(
    level=logging.DEBUG,
    force=True,
    handlers=[
        logging.FileHandler('run.log'),
        logging.StreamHandler(sys.stdout)
    ]
)

load_dotenv()
NAMESERVER_ADDRESS = os.getenv("APP_NAMESERVER_URL", "http://localhost:7000")
MATCHMAKER_ADDRESS = os.getenv("APP_MATCHMAKER_BIND", "localhost:").split(':')[0]
GAMESERVER_ADDRESS = os.getenv("APP_GAMESERVER_BIND", "localhost:").split(':')[0]
BASE_MATCHMAKER_PORT = int(os.getenv("APP_MATCHMAKER_BIND", ":2000").split(':')[-1])
BASE_GAMESERVER_PORT = int(os.getenv("APP_GAMESERVER_BIND", ":2000").split(':')[-1])

HELP_MESSAGE = (
    'Commands:\n'
    '- help:\n'
    '\t Display this message\n'
    '- quit:\n'
    '\t Shutdown all servers and exit\n'
    '- start <"matchmaker" | "gameserver">+:\n'
    '\t Spawn new matchmaker(s) or gameserver(s)\n'
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
            logging.info('[%s] %s ', self.pid, line.rstrip())
        if self.proc.poll() is None:
            self.proc.wait()
        logging.debug('[%s] Terminated', self.pid)
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
            cmd=[NPM_CMD, 'run', 'start'],
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
                logging.info(HELP_MESSAGE)
            elif cmd.lower() == 'list':
                if len(args) != 1:
                    logging.error('Improper usage: expected 1 argument found %d', len(args))
                    continue
                server_type = args[0]
                if server_type.lower() == 'matchmaker':
                    logging.info('Matchmakers:')
                    for matchmaker in matchmakers:
                        logging.info('\t- %s', matchmaker)
                elif server_type.lower() == 'gameserver':
                    logging.info('Game Servers:')
                    for game_server in game_servers:
                        logging.info('\t- %s', game_server)
                else:
                    logging.info(
                        'Unexpected server type "%s": expected "matchmaker" or "gameserver"',
                        server_type
                    )
            elif cmd.lower() == 'stop':
                if len(args) != 1:
                    logging.error('Improper usage: expected 1 argument found %d', len(args))
                    continue
                pid = args[0]
                if pid in matchmakers:
                    logging.info('Stopping matchmaker "%s"', pid)
                    matchmakers[pid].stop()
                    logging.info('Matchmaker "%s" stopped', pid)
                elif pid in game_servers:
                    logging.info('Stopping game server "%s"', pid)
                    game_servers[pid].stop()
                    logging.info('Game server "%s" stopped', pid)
                else:
                    logging.error(
                        'Unexpected server type "%s": expected "matchmaker" or "gameserver"',
                        server_type
                    )
            elif cmd.lower() == 'kill':
                if len(args) != 1:
                    logging.error('Improper usage: expected 1 argument found %d', len(args))
                    continue
                pid = args[0]
                if pid in matchmakers:
                    logging.info('Killing matchmaker "%s"', pid)
                    matchmakers[pid].kill()
                    logging.info('Matchmaker "%s" killed', pid)
                elif pid in game_servers:
                    logging.info('Killing game server "%s"', pid)
                    game_servers[pid].kill()
                    logging.info('Game server "%s" killed', pid)
                else:
                    logging.error(
                        'Unexpected server type "%s": expected "matchmaker" or "gameserver"',
                        server_type
                    )
            elif cmd.lower() == 'start':
                if len(args) < 1:
                    logging.error(
                        'Improper usage: expected 1 or more arguments found %d',
                        len(args)
                    )
                for server_type in args:
                    if server_type.lower() == 'matchmaker':
                        pid = f'MM{matchmaker_id:02}'
                        port = BASE_MATCHMAKER_PORT + matchmaker_id
                        matchmaker = CheckeredProcess(
                            pid=pid,
                            cmd=[
                                'go',
                                'run',
                                'cmd/matchmaker/main.go',
                                f'--addr={MATCHMAKER_ADDRESS}:{port}'
                            ],
                            cwd='backend',
                            callback = lambda: matchmakers.pop(pid, None)
                        )
                        matchmaker_id += 1
                        matchmakers[pid] = matchmaker
                        matchmaker.start()
                        logging.info('Created matchmaker with pid: "%s"', pid)
                    elif server_type.lower() == 'gameserver':
                        pid = f'GS{game_server_id:02}'
                        port = BASE_GAMESERVER_PORT + game_server_id
                        game_server = CheckeredProcess(
                            pid=pid,
                            cmd=[
                                'go',
                                'run',
                                'cmd/game-server/main.go',
                                f'--addr={GAMESERVER_ADDRESS}:{port}'
                            ],
                            cwd='backend',
                            callback = lambda: game_servers.pop(pid, None)
                        )
                        game_server_id += 1
                        game_servers[pid] = game_server
                        game_server.start()
                        logging.info('Created gameserver with pid: "%s"', pid)
                    else:
                        logging.error('Unexpected server type "%s"', server_type)
            else:
                logging.error('Unrecognized command "%s"', cmd)
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
