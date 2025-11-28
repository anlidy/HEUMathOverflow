#!/bin/env python3
import os
import sys
import time
import shutil
import argparse
import subprocess
import signal
from pathlib import Path
import functools


class Colors:
    RED = '\033[91m'
    GREEN = '\033[92m'
    YELLOW = '\033[93m'
    BLUE = '\033[94m'
    MAGENTA = '\033[95m'
    CYAN = '\033[96m'
    WHITE = '\033[97m'
    BOLD = '\033[1m'
    UNDERLINE = '\033[4m'
    END = '\033[0m'


def cprint(*values, color=Colors.WHITE, **kwargs):
    print(color + ' '.join(str(v) for v in values) + Colors.END, **kwargs)


def locate_project_root(cd: bool = True) -> Path:
    '''定位项目根目录（包含 backend 和 frontend 目录的目录）'''
    script_dir = Path(__file__).resolve()
    root = [
        p for p in script_dir.parents
        if (p.name.endswith('backend'))
    ]
    neibours = os.listdir(script_dir.parent)
    flag = (('backend' in neibours) or (
        'frontend' in neibours)) and ('.git' in neibours)
    root = root[0].parent if root else '.' if flag else None
    if root is None:
        raise FileNotFoundError(
            '未能定位项目根目录，请确保脚本debug.py存在于backend目录及其子目录中')
    root = Path(root)
    if cd:
        try:
            os.chdir(root)
        except Exception as e:
            cprint('进入项目根目录失败，错误信息：'+str(e), color=Colors.RED)
            cprint("请确保脚本debug.py存在于backend目录及其子目录中", color=Colors.CYAN)
    return root


def cd_run(target_dir: Path):
    '''带参装饰器：切换到 `target_dir` 目录，运行被装饰的函数，然后切回原目录'''
    def decorator(func):
        @functools.wraps(func)
        def wrapper(*args, **kwargs):
            cwd = os.getcwd()
            try:
                os.chdir(target_dir)
                return func(*args, **kwargs)
            finally:
                try:
                    os.chdir(cwd)
                except Exception:
                    cprint('无法切回原目录', color=Colors.RED)
        return wrapper
    return decorator


def crun(cmd, **kwargs):
    cprint(f'RUN: {" ".join(cmd)}', color=Colors.BLUE)
    return subprocess.run(cmd, check=True, **kwargs)


def _run_with_signal_forward(cmd, stdin=None, stdout=None, stderr=None):
    '''运行 cmd 并将 SIGINT/SIGTERM 转发给子进程，以便 Ctrl+C 能作用于子进程。
    对子进程使用单独的进程组，并将信号转发到该进程组。
    这样 Python 父进程可以保留并在需要时执行清理工作。'''
    cprint(f'RUN (forward signals): {" ".join(cmd)}', color=Colors.CYAN)
    # start child in its own process group
    proc = subprocess.Popen(cmd, preexec_fn=os.setsid,
                            stdin=stdin, stdout=stdout, stderr=stderr)

    def _forward(signum, frame):
        try:
            os.killpg(proc.pid, signum)
        except Exception:
            pass

    # 设置新的信号处理函数，并保存旧的处理函数以便恢复
    old_int = signal.getsignal(signal.SIGINT)
    old_term = signal.getsignal(signal.SIGTERM)
    signal.signal(signal.SIGINT, _forward)
    signal.signal(signal.SIGTERM, _forward)

    try:
        rc = proc.wait()
    finally:
        # 无论如何都要恢复旧的信号处理函数
        signal.signal(signal.SIGINT, old_int)
        signal.signal(signal.SIGTERM, old_term)

    return rc


def keep_dir(func):
    '''装饰器：保持当前工作目录不变，便于被其他代码调用'''
    @functools.wraps(func)
    def wrapper(*args, **kwargs):
        cwd = os.getcwd()
        try:
            return func(*args, **kwargs)
        finally:
            try:
                os.chdir(cwd)
            except Exception:
                pass  # 或记录日志
    return wrapper
