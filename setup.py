import setuptools
import os
import shutil
from importlib.machinery import SourceFileLoader

version_module = SourceFileLoader(
    'version_module', 'di_store/version.py').load_module()
version = version_module.version

dist_path = 'dist'
if os.path.exists(dist_path):
    shutil.rmtree(dist_path)

setuptools.setup(
    name='DI-store',
    version=version,
    description='Decision AI Store',
    author='OpenDILab Contributors',
    author_email='opendilab.contact@gmail.com',
    license='Apache License, Version 2.0',
    keywords='Decision AI Store',
    packages=setuptools.find_packages(),
    package_data={'di_store': ['bin/*']},
    install_requires=['grpcio', 'pyarrow', 'etcd3-py', 'jaeger-client',
                      'grpcio-opentracing', 'grpcio-tools', 'pyyaml', 'numpy'],
    python_requires='>=3.6',
    zip_safe=False,
    entry_points={
        'console_scripts': ['di_store=di_store.cmd.executor:main'],
    },
)
