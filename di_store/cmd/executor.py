import sys
import subprocess
from os import path
from di_store.driver.etcd_server_driver import main as etcd_server_driver_main

dir_path = path.dirname(path.realpath(__file__))
executable = {
    'node_tracker': path.join(dir_path, '../bin', 'node_tracker'),
    'storage_server': path.join(dir_path, '../bin', 'storage_server'),
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
            exec = executable[exec_name]
            cmd = [exec] + sys.argv[2:]
            subprocess.run(cmd)
    except Exception as e:
        print('Exception:',e)
        show_help(root_cmd)


def show_help(root_cmd):
    print('for more information:')
    print(
        f'    {root_cmd} [etcd_server | node_tracker | storage_server] --help')


if __name__ == '__main__':
    main()
