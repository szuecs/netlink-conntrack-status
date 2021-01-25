This tool exports data to query by your monitoring system.
Data you get is similar to `conntrack -S`, which is not yet exposed
via Prometheus node-exporter nor cadvisor.


```
# build and show version string
% make
GO111MODULE= go build -ldflags "-X main.version=v0.0.1 -X main.commit=0d1bf2d" -o netlink-conntrack-status .
% ./netlink-conntrack-status --version
./netlink-conntrack-status: v0.0.1 - commit: 0d1bf2d

# build with custom version and git commit hash
% make VERSION=0.1 COMMIT_HASH=foo
GO111MODULE= go build -ldflags "-X main.version=0.1 -X
main.commit=foo" -o netlink-conntrack-status .
% ./netlink-conntrack-status --version
./netlink-conntrack-status: 0.1 - commit: foo

# run as daemon get Prometheus exported counters, SIGINT and SIGTERM shutdown
% sudo ./netlink-conntrack-status --daemon -update-interval 5s
^C2021/01/25 13:02:42 shutting down

# run once, get JSON
% sudo ./netlink-conntrack-status
{"found":0,"invalid":42,"ignore":26332,"insert":0,"insert_failed":0,"drop":0,"early_drop":0,"error":0,"search_restart":187}
```
