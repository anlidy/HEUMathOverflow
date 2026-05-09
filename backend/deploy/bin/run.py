#!/usr/bin/env python3
import os
import sys
import shutil
import subprocess
from pathlib import Path
from scriptlib import cprint, Colors, locate_project_root, crun, cd_run

@cd_run(target_dir=Path('./backend') / 'deploy' / 'docker')
def compose(args: list[str], **kwargs):
    return crun(['docker', 'compose'] + ['-f', 'docker-compose.service.yml'] + args, **kwargs)

def main(argv: list[str] | None = None) -> int:
    argv = list(argv or sys.argv[1:])

    try:
        root = locate_project_root()
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
    print(os.getcwd())
    try:
        process = compose(argv)
        return process.returncode
    except Exception as e:
        cprint('运行 docker compose 失败，错误信息：'+str(e), color=Colors.RED)
        return 1


if __name__ == '__main__':
    raise SystemExit(main())
