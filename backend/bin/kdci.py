#!/bin/env python3
import os
import subprocess
import time
from pathlib import Path
import argparse
from run import run_cmd as composer
from debug import Colors, cprint, locate_project_root, _run, keep_dir, cd_run


def kubeconfig_gen():
    dir = './backend/docker/cluster/k8s.config'+time_stamp
    pwd = os .getcwd()
    cprint('生成新的 Kubernetes 配置文件至 '+dir, color=Colors.BLUE)
    os.makedirs(dir, exist_ok=True)
    os.chdir(dir)
    subprocess.run(['kompose', '-f', '../../docker-compose.service.yml', 'convert',
                    '--controller', 'deployment', '--out', './'], check=True)
    subprocess.run(['kompose', '-f', '../../docker-compose.yml', 'convert',
                    '--controller', 'statefulset', '--out', './'], check=True)
    os.chdir(pwd)
    return dir


def load(images):
    cprint('装载镜像至 KIND 集群...', color=Colors.BLUE)
    for img in images:
        _run(['kind', 'load', 'docker-image', img])


@cd_run(target_dir=Path('./backend') / 'docker')
def build(specific: None | list[str] = None, project: str = 'kind') -> list[str]:
    process = composer('build', *(specific or []), project=project,
                       check=True, capture_output=True, text=True)
    images: str = process.stderr
    images = [line.strip().split()[0].strip()
              for line in images.splitlines() if line.strip().endswith('Built')]
    return images


def image_adapt(images: list[str]) -> list[str]:
    for i, img in enumerate(images):
        new = '-'.join(img.split('-')[1:])
        subprocess.run(['docker', 'tag', img, new], check=True)
        images[i] = new
    return images


def kind(name='kind'):
    clusters = subprocess.run(
        ['kind', 'get', 'clusters'], check=True, capture_output=True, text=True)
    clusters = clusters.stdout.strip().splitlines()
    if name in clusters:
        return 0
    proc = subprocess.run(['kind', 'create', 'cluster', '--name', name, '--config',
                           str(locate_project_root() / 'backend' / 'docker'/'cluster'/'kind.config.yaml')], check=True)
    return proc.returncode


def arg_handle(args=None):
    parser = argparse.ArgumentParser(description='KIND CI 部署脚本')
    parser.add_argument('--config-gen', action='store_true',
                        help='使用 kompose 生成新的 Kubernetes 配置文件')
    parser.add_argument('service', nargs='*', default=None, choices=['forum', 'user', 'audit'],
                        help='指定要构建和加载到 KIND 的服务，默认全部')
    arg = parser.parse_args(args)
    if arg.service:
        arg.service = [i+'-service' for i in arg.service]
    return arg

def apply(dir):
    cprint('正在应用 Kubernetes 配置文件：'+dir, color=Colors.BLUE)
    subprocess.run(['kubectl', 'apply', '-f', dir], check=True)


@keep_dir
def main(args=None):
    root = locate_project_root()
    arg = arg_handle(args)
    kind('kind')
    cprint(f'以根目录 @ {root} 启动 KIND CI', color=Colors.CYAN)
    images = image_adapt(build(project='kind', specific=arg.service))
    load(images)
    dir = kubeconfig_gen() if arg.config_gen else './backend/docker/cluster/k8s.config'
    apply(dir)
    cprint('KIND CI 布置完成.', color=Colors.GREEN)
    


if __name__ == '__main__':
    globals().update({'rootdir': locate_project_root(cd=False)})
    globals().update({'time_stamp': time.strftime(
        '%Y%m%d%H%M%S', time.localtime())})
    main(args=None)
