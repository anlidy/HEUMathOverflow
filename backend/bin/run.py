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
from debug import cprint, Colors, locate_project_root


help_str = """
(Unix like, interpreter path needed if on Windows)
Usage: run [docker-compose up args...]

Examples:
  ./run.py up -d         # detached
  ./run.py up --build -d # build then detached
  ./run.py down       # run `down` instead of `up`
"""


def run_cmd(*args, file='docker-compose.service.yml', project='backend', **kwargs) -> subprocess.CompletedProcess:
    """
    Compose commands: match existing bash behavior
    """
    full = ['docker', 'compose'] + (['-p', project] if project else []) + \
        (['-f', file] if file else []) + list(args)
    print('Running:', ' '.join(full))
    return subprocess.run(full, **kwargs)


def main(argv: list[str] | None = None) -> int:
    argv = list(argv or sys.argv[1:])

    try:
        root = locate_project_root()
        os.chdir(root / 'backend' / 'docker')
    except Exception as e:
        cprint('进入项目根目录失败，错误信息：'+str(e), color=Colors.RED)
        cprint("请确保脚本run.py存在于backend目录及其子目录中", color=Colors.CYAN)
        return 1

    # Ensure .env exists
    env_path = Path('.env')
    example = Path('.env.example')
    if not env_path.exists() and example.exists():
        try:
            shutil.copy(example, env_path)
            print('复制文件 .env.example -> .env')
        except Exception as e:
            print(f'复制文件失败: {e}', file=sys.stderr)

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
    stop_cmd = ['down', 'stop', 'kill']
    if cmd in stop_cmd:
        run_cmd(cmd, *extra).returncode or run_cmd(cmd,
                                                   *extra, file='').returncode
    else:
        return run_cmd(cmd, *extra, file='').returncode \
            or run_cmd(cmd, *extra).returncode
    # rc = 0
    # if cmd == 'down':
    #     rc = run_cmd('docker-compose.service.yml')
    #     if rc != 0:
    #         return rc
    #     rc = run_cmd('docker-compose.yml')
    #     if rc != 0:
    #         return rc
    #     rc = subprocess.run(['../bin/del_img']).returncode
    # else:
    #     rc = run_cmd('docker-compose.yml')
    #     if rc != 0:
    #         return rc
    #     rc = run_cmd('docker-compose.service.yml')

    # return rc


if __name__ == '__main__':
    raise SystemExit(main())
