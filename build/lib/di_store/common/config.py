import yaml

from .exception import ConfigError

unset = object()


class Config(object):
    def __init__(self, conf_path):
        with open(conf_path) as file:
            self.conf = yaml.load(file, Loader=yaml.FullLoader)
        self.conf_path = conf_path

    # todo
    def validate_config(self):
        pass

    def etcd_server(self, hostname=None, default=unset):
        if 'etcd_servers' not in self.conf or not self.conf['etcd_servers']:
            raise ConfigError(
                f'`etcd_servers` is not found in config file {self.conf_path}')
        if hostname is None:
            return self.conf['etcd_servers']
        server_info = [server for server in self.conf['etcd_servers']
                       if server['hostname'] == hostname]
        if server_info:
            server_info = server_info[0]
            if 'listen_client_urls' not in server_info or not server_info['listen_client_urls']:
                raise ConfigError(
                    f'`listen_client_urls` of etcd server {hostname} is not found in config file {self.conf_path}')
            return server_info
        elif default is unset:
            raise ConfigError(
                f'etcd server with hostname `{hostname}` is not in config file {self.conf_path}')
        else:
            return default

    def node_tracker(self, hostname=None, default=unset):
        if 'node_trackers' not in self.conf or not self.conf['node_trackers']:
            raise ConfigError(
                f'`node_trackers` is not found in config file {self.conf_path}')
        if hostname is None:
            return self.conf['node_trackers']
        node_tracker = [
            itr for itr in self.conf['node_trackers'] if itr['hostname'] == hostname]
        if node_tracker:
            node_tracker = node_tracker[0]
            if 'rpc_host' not in node_tracker or 'rpc_port' not in node_tracker:
                raise ConfigError(
                    f'`rpc_host` or `rpc_port` of node tracker {hostname} is not found in config file {self.conf_path}')
            return node_tracker
        elif default is unset:
            raise ConfigError(
                f'node tracker with hostname {hostname} is not in config file {self.conf_path}')
        else:
            return default

    def storage_server(self, hostname):
        if 'storage_servers' not in self.conf or not self.conf['storage_servers']:
            raise ConfigError(
                f'`storage_servers` is not found in config file {self.conf_path}')
        storage_server = [
            itr for itr in self.conf['storage_servers'] if itr['hostname'] == hostname
        ]
        if not storage_server:
            default_storage_server = [
                itr for itr in self.conf['storage_servers'] if itr['hostname'] == '*'
            ]
            if not default_storage_server:
                raise ConfigError(
                    f'storage server with hostname {hostname} is not found in config file {self.conf_path}')
            storage_server = default_storage_server[0].copy()
            storage_server['hostname'] = hostname
        else:
            storage_server = storage_server[0]

        return storage_server

    def check_existence(self, section_name):
        if section_name not in self.conf or not self.conf[section_name]:
            raise ConfigError(
                f'{section_name} is not found in config file {self.conf_path}')

    def find_by_hostname(self, section, hostname, try_default=False):
        target = next(
            (itr for itr in self.conf[section] if itr['hostname'] == hostname), None)
        if not target and try_default:
            target = next(
                (itr.copy() for itr in self.conf[section] if itr['hostname'] == '*'), None)

        if not target:
            raise ConfigError(
                f'item with hostname {hostname} is not found in config file {self.conf_path}')

    # def storage_client(self, hostname):
    #     self.check_existence('storage_clients')
    #     info = self.find_by_hostname(
    #         'storage_clients', hostname, try_default=True)
    #     info['storage_server'] = info['storage_server'].replace(
    #         '{hostname}', hostname)
    #     return info
