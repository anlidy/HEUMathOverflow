#!/usr/bin/env python3
"""run.py - Python wrapper to run docker compose for this project.

Behavior:
- cd to repo/backend/docker
- if .env missing and .env.example exists, copy it
- first arg is compose subcommand (default: up), remaining args forwarded
- runs docker compose for both docker-compose.yml and docker-compose.service.yml
"""
import os
import sys
import shutil
import subprocess
from pathlib import Path


help_str = """
(Unix like, interpreter path needed if on Windows)
Usage: run [docker-compose up args...]

Examples:
  ./run.py up -d         # detached
  ./run.py up --build -d # build then detached
  ./run.py down       # run `down` instead of `up`
"""

def main(argv: list[str] | None = None) -> int:
    argv = list(argv or sys.argv[1:])

    # Locate script directory and change to backend/docker
    script_dir = Path(__file__).resolve().parent
    docker_dir = (script_dir / '..' / 'docker').resolve()
    try:
        os.chdir(docker_dir)
    except Exception as e:
        print(f"Failed to change to docker dir {docker_dir}: {e}", file=sys.stderr)
        return 2

    # Ensure .env exists
    env_path = Path('.env')
    example = Path('.env.example')
    if not env_path.exists() and example.exists():
        try:
            shutil.copy(example, env_path)
            print('Copied .env.example -> .env')
        except Exception as e:
            print(f'Failed to copy .env.example: {e}', file=sys.stderr)

    # Help
    if len(argv) >= 1 and argv[0] in ('-h', '--help'):
        print(help_str)
        return 0

    # Require explicit compose subcommand (no default)
    if len(argv) == 0:
        print('Missing compose subcommand (e.g. up or down).')
        print(help_str)
        return 2

    if argv[0].startswith('-'):
        print('First argument must be a compose subcommand (e.g. up or down); options must follow the subcommand.')
        print(help_str)
        return 2

    cmd = argv[0]
    extra = argv[1:]

    # Compose commands: match existing bash behavior
    # if down -> run service then base; else -> base then service
    def run_cmd(file: str) -> int:
        full = ['docker', 'compose', '-p', 'backend', '-f', file, cmd] + extra
        print('Running:', ' '.join(full))
        try:
            res = subprocess.run(full, check=False)
            return res.returncode
        except FileNotFoundError:
            print('docker compose not found. Is Docker installed and in PATH?', file=sys.stderr)
            return 127

    rc = 0
    if cmd == 'down':
        rc = run_cmd('docker-compose.service.yml')
        if rc != 0:
            return rc
        rc = run_cmd('docker-compose.yml')
        if rc != 0:
            return rc
        rc = subprocess.run(['../bin/del_img']).returncode
    else:
        rc = run_cmd('docker-compose.yml')
        if rc != 0:
            return rc
        rc = run_cmd('docker-compose.service.yml')

    return rc


if __name__ == '__main__':
    raise SystemExit(main())
