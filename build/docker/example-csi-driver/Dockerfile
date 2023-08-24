FROM BASE_IMAGE

WORKDIR /opt/eazycloud


RUN ln -sf /usr/share/zoneinfo/Asia/Shanghai /etc/localtime && \
      echo "Asia/Shanghai" > /etc/timezone && \
      mkdir -p /var/log/eazycloud && \
      mkdir -p /etc/eazycloud

COPY IMAGE_COMMAND /opt/eazycloud/bin/
COPY IMAGE_CONFIG /etc/eazycloud/

ENTRYPOINT ["/opt/eazycloud/bin/IMAGE_COMMAND"]
CMD ["-c", "/etc/eazycloud/IMAGE_CONFIG"]