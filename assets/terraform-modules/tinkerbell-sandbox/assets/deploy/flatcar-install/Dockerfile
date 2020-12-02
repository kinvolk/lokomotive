FROM ubuntu

RUN apt update && \
		apt install -y udev gpg wget

RUN wget https://raw.githubusercontent.com/flatcar-linux/init/flatcar-master/bin/flatcar-install -O /usr/local/bin/flatcar-install && \
    chmod +x /usr/local/bin/flatcar-install
