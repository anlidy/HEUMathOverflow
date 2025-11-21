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


def _run(cmd, **kwargs):
    cprint(f'RUN: {" ".join(cmd)}', color=Colors.BLUE)
    subprocess.run(cmd, check=True, **kwargs)


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


def read_service_port(arg) -> str:
    '''从服务配置获取端口'''
    if arg.port is not None:
        return arg.port
    with open('./backend/cmd/'+arg.service+'-service/'+arg.service+'.yaml') as f:
        lines = f.readlines()
        for i, line in enumerate(lines):
            if line.lstrip().startswith('server:'):
                for i2, l in enumerate(lines[i+1:]):
                    if l.lstrip().startswith('port:'):
                        arg.port = l.strip().split(':')[-1].strip()
                        return arg.port
    return None


def read_network():
    '''从 docker-compose.yml 获取网络名称'''
    with open('backend/docker/docker-compose.yml') as f:
        lines = f.readlines()
        for i, line in enumerate(lines):
            if line.lstrip().startswith('networks:'):
                network_name = [l.strip().split(':')[-1].strip()
                                for l in lines[i+1:]
                                if l.lstrip().startswith('name')][0]
                return network_name
    return None


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


def arg_handle(args=None):
    '''处理命令行参数'''
    if args is None:
        args = sys.argv[1:]
    parser = argparse.ArgumentParser(
        description='调试正在运行的服务，停止名称中包含关键词 <service> 的容器，并运行一个连接到相同网络的调试容器')
    parser.add_argument('service', default='forum', nargs='?', choices=[
                        'forum', 'user', 'audit'],
                        help='要重建的服务名称，默认为 forum')
    network_name = read_network()
    parser.add_argument('-n', '--network', default=network_name,
                        help='连接到的docker网络，默认为docker-compose.yml中定义的网络')
    parser.add_argument('--dry-run', action='store_true',
                        help='只展示将要执行的命令，而不实际执行')
    parser.add_argument('-it', '--interactive', action='store_true',
                        help='以交互模式运行调试容器，附加到容器的shell，容器引导变为tail，可手工启动myservice文件或其他命令')
    parser.add_argument('-p', '--port', default=None,
                        help='服务监听的端口，默认从服务yaml配置中读取')
    parser.add_argument('--rm', action='store_true',
                        help='退出后删除调试容器')
    arg = (parser.parse_args(args))
    return arg


def _cmd_gen(arg):
    to_stop = subprocess.run(
        r'docker ps -a --format "{{.Names}}|{{.ID}}"',
        shell=True, capture_output=True, text=True).stdout.strip().split('\n')
    to_stop = dict(line.split('|')
                   for line in to_stop if arg.service+'-service' in line)
    stop_cmd = [['docker', 'stop', container_id]
                for container_id in to_stop.values()]
    tag = arg.service + '-service:debug-' + arg.time_stamp
    buildfile = './backend/docker/Dockerfile.debug'
    build_cmd = [
        'docker', 'buildx', 'build',
        '--progress=plain',
        '-f', buildfile,
        '-t', tag,
        '--build-arg', f'SERVICE={arg.service}',
        '--build-arg', f'PORT={arg.port}',
        './backend'
    ]
    container_name = f"{arg.service}-service-debug" + \
        arg.time_stamp
    run_cmd = [
        'docker', 'run',
        '-it' if arg.interactive else '-d',
        '--network', arg.network,
        '-p', f'{arg.port}:{arg.port}',
        '-e', f'SERVICE={arg.service}',
        '--name', container_name,
        '-w', '/app',
        tag,
        '' if not arg.interactive else '/bin/sh'
    ]
    run_cmd = [s for s in run_cmd if s != '']
    log_cmd = ['docker', 'logs', '-f',
               container_name] if not arg.interactive else []
    cmds = {'stop': stop_cmd,
            'build': [build_cmd],
            'run': [run_cmd],
            'log': [log_cmd] if log_cmd else []}
    return cmds


@keep_dir
def main(args=None):
    times = [time.time()]
    locate_project_root()
    arg = arg_handle(args)
    arg.time_stamp = time.strftime('%Y%m%d%H%M%S')
    read_service_port(arg)
    cmds = _cmd_gen(arg)
    if arg.dry_run:
        for cmd in cmds.values():
            for c in cmd:
                print(' '.join(c))
    else:
        for cmd in cmds['stop']:
            _run(cmd)
        for cmd in cmds['build']:
            _run(cmd)
        times.append(time.time())
        print(
            Colors.GREEN+'Debug container started., used ' +
            Colors.MAGENTA+f'{times[-1]-times[-2]:.2f} seconds' +
            Colors.END)
        for cmd in cmds['run']:
            _run(cmd, stdin=sys.stdin if arg.interactive else None,
                 stdout=sys.stdout, stderr=sys.stderr)
        for cmd in cmds['log']:
            _run_with_signal_forward(
                cmd, stdin=sys.stdin if arg.interactive else None, stdout=sys.stdout, stderr=sys.stderr)
        if arg.rm:
            print()
            _run(['docker', 'rm', '-f',
                 f'{arg.service}-service-debug' + arg.time_stamp])
    cprint('Debugging session ended.', color=Colors.GREEN)


if __name__ == '__main__':
    raise SystemExit(main())
