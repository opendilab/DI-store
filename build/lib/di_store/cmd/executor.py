import sys
import subprocess
import os
import signal
from os import path
from di_store.driver.etcd_server_driver import main as etcd_server_driver_main
from di_store.common.util import create_executable_tmp_file

platform = sys.platform
if platform not in ('darwin', 'linux'):
    raise Exception(f'unsupported platform: {platform}')

dir_path = path.dirname(path.realpath(__file__))
executable = {
    'node_tracker': path.join(dir_path, '../bin', f'node_tracker_{platform}'),
    'storage_server': path.join(dir_path, '../bin', f'storage_server_{platform}'),
}


def main():
    args = sys.argv
    root_cmd = path.basename(sys.argv[0])
    try:
        if len(args) < 2:
            raise Exception('not enough arguments')
        exec_name = args[1]

        sys.argv[0] = exec_name
        if exec_name == 'help':
            show_help(root_cmd)
        elif exec_name == 'etcd_server':
            etcd_server_driver_main(sys.argv[2:])
        elif exec_name not in executable:
            raise Exception(f'command `{exec_name}` is not supported')
        else:
            if exec_name == 'storage_server':
                plasma_store_server_exec = path.join(
                    dir_path, '../bin', f'plasma-store-server-{platform}')
                plasma_store_server_exec_tmp = create_executable_tmp_file(
                    plasma_store_server_exec)
                os.environ['PLASMA_STORE_SERVER_EXEC'] = plasma_store_server_exec_tmp.name

            exec = executable[exec_name]
            exec_tmp = create_executable_tmp_file(exec)
            cmd = [exec_tmp.name] + sys.argv[2:]
            p = subprocess.Popen(cmd)

            def handle_signal(sig, frame):
                print('receive signal:', signal.Signals(sig).name)
                p.send_signal(signal.SIGINT)
                sys.exit()

            for sig in (signal.SIGINT, signal.SIGTERM):
                signal.signal(sig, handle_signal)

            p.wait()

    except Exception as e:
        print('Exception:', e)
        show_help(root_cmd)


def show_help(root_cmd):
    print('for more information:')
    print(
        f'    {root_cmd} [etcd_server | node_tracker | storage_server] --help')


if __name__ == '__main__':
    main()
