FROM centos/python-36-centos7

WORKDIR /di_store
ENV LC_ALL en_US.UTF-8
User root

RUN yum install -y epel-release
RUN yum install -y gflags
RUN python3 -m pip install --upgrade pip

ADD setup.py setup.py
ADD di_store di_store
ADD go go
RUN python3 setup.py sdist
RUN python3 -m pip install --no-cache-dir ./dist/*

RUN rm -rf /di_store/*