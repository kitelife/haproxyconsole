#!/bin/sh

/usr/local/haproxy/sbin/haproxy -f /usr/local/haproxy/conf/haproxy.conf -st `cat /usr/local/haproxy/haproxy.pid`
