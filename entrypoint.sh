#!/bin/sh
# aweful-dns binds one resolver per team on 127.0.x.y:53; dnsmasq on the container address forwards by team name.
set -e
mkdir -p /etc/dnsmasq.d
aweful-dns -config-dir /confs -dnsmasq-file /etc/dnsmasq.d/aweful-dns.conf "$@" &
sleep 2
exec dnsmasq --keep-in-foreground --bind-interfaces --listen-address="$(hostname -i | awk '{print $1}')" \
  --no-resolv --no-hosts --cache-size=0 --log-queries --conf-dir=/etc/dnsmasq.d
