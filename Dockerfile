FROM fedora:43
SHELL [ "/bin/bash", "-c" ]
COPY build/miniws /bin/miniws
RUN mkdir -p "/data/{logs,config}"
RUN ls "/data"
EXPOSE 8040/tcp
ENTRYPOINT [ "/bin/miniws", "--port", "8040", "--logs-folder", "/data/logs", \
        "--config-folder", "/data/config" ]