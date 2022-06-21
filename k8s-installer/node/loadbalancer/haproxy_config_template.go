package loadbalancer

const haproxyTemplate = `global
        log /dev/log    local1 warning
        chroot /var/lib/haproxy
        user haproxy
        group haproxy
        daemon
        nbproc 1

defaults
        log     global
        timeout connect 5s
        timeout client  10m
        timeout server  10m
%s`

const sectionTemplate = `%s %s
%s`
const optionTemplate = "        %s"