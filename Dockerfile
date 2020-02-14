FROM centos:7.4.1708

RUN yum install -y net-tools telnet

ADD bin/kube-webhook /kube-webhook

ENTRYPOINT ["/kube-webhook"]