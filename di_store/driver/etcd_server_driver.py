import os
from os import path
import shutil
import argparse
import socket
import platform
from signal import SIGINT, SIGTERM
import asyncio
from di_store.common.config import Config
from di_store.common.util import AsyncSubprocess, create_executable_tmp_file


async def run(command):
    process = await AsyncSubprocess.create(command)
    for signal in [SIGINT, SIGTERM]:
        loop = asyncio.get_event_loop()
        loop.add_signal_handler(signal, process.terminate)
    await process.wait()


def delete_default_dir():
    current_dir = os.getcwd()
    default_dir = path.join(current_dir, 'default.etcd')
    shutil.rmtree(default_dir, True)


def main(args):
    parser = argparse.ArgumentParser(description='start a etcd server')
    parser.add_argument('conf_path',  type=str, nargs=1,
                        help='config file path for etcd server')
    parser.add_argument('--log-level', choices=['debug', 'info', 'warn', 'error', 'panic', 'fatal'],
                        help='log level for etcd server')

    args = parser.parse_args(args)
    conf_path = args.conf_path[0]
    log_level = args.log_level  # todo read log level from config file
    hostname = socket.gethostname()

    config = Config(conf_path)
    server_info = config.etcd_server(hostname)
    listen_client_urls = ','.join(server_info['listen_client_urls'])
    listen_peer_urls = ','.join(server_info['listen_peer_urls'])
    unsafe_no_fsync = True

    bin_path = os.path.abspath(os.path.join(
        os.path.dirname(__file__), '..', 'bin', f'etcd-{platform.system().lower()}'))
    bin_tmp = create_executable_tmp_file(bin_path)
    cmd = [bin_tmp.name]

    if unsafe_no_fsync:
        cmd.append('--unsafe-no-fsync')

    if log_level:
        cmd.extend(['--log-level', log_level])

    cmd.extend(['--advertise-client-urls', listen_client_urls,
                '--listen-peer-urls', listen_peer_urls,
                '--listen-client-urls', listen_client_urls])

    delete_default_dir()

    loop = asyncio.get_event_loop()
    loop.run_until_complete(run(cmd))


if __name__ == "__main__":
    main(None)
