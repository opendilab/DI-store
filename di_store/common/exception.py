class Error(Exception):
    def __str__(self):
        cls_name = type(self).__name__
        msg = super(Error, self).__str__()
        return '{}({})'.format(cls_name, msg)


class ConfigError(Error):
    pass
