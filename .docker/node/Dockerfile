FROM golang:1.25

ENV GO111MODULE on

ARG USERNAME=pluslab
ARG GROUPNAME=pluslab
ARG UID=1000
ARG GID=1000
ARG PASSWORD=password
RUN apt-get update && apt-get install -y iproute2 iputils-ping dnsutils iperf3 vim nano tcpdump sudo lsb-release && \
	groupadd -g $GID $GROUPNAME && \
	useradd -m -s /bin/bash -u $UID -g $GID -G sudo $USERNAME && \
	echo $USERNAME:$PASSWORD | chpasswd && \
	echo "$USERNAME   ALL=(ALL) NOPASSWD:ALL" >> /etc/sudoers
