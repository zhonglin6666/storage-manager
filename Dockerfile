FROM centos:7.4.1708

RUN cp -a /usr/share/zoneinfo/Asia/Shanghai /etc/localtime \
    && yum install -y net-tools telnet \
    && yum -y install nfs-utils \
    && yum -y install epel-release \
    && yum -y install jq \
    && yum clean all

ADD bin/storage-manager /storage-manager

ENTRYPOINT ["/storage-manager"]