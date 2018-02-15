FROM ubuntu:trusty
RUN mkdir /var/log/olscheduler
RUN ln -sf /dev/stdout /var/log/olscheduler/access.log
RUN ln -sf /dev/stderr /var/log/olscheduler/error.log

COPY bin/olscheduler  /usr/bin/

CMD ["/usr/bin/olscheduler", "start", "-c", "/etc/olscheduler/conf/olscheduler.json"]

LABEL org.label-schema.vendor="olscheduler" \
      org.label-schema.url="https://github.com/gtotoy/olscheduler" \
      org.label-schema.name="olscheduler" \
      org.label-schema.description="An Extensible Scheduler for the OpenLambda FaaS Platform" 
