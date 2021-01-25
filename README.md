This tool exports data to query by your monitoring system.
Data you get is similar to `conntrack -S`, which is not yet exposed
via Prometheus node-exporter nor cadvisor.


```
# run as daemon get Prometheus exported counters, SIGINT and SIGTERM shutdown
% sudo ./netlink-conntrack-status --daemon -update-interval 5s
^Zzsh: exit 148
zsh: suspended
% bg
[2]  - continued

# show metrics
% curl localhost:9090/metrics
# HELP netlink_conntrack_drop The total of conntrack -S drop.
# TYPE netlink_conntrack_drop counter
netlink_conntrack_drop 0
# HELP netlink_conntrack_earlyDrop The total of conntrack -S earlyDrop.
# TYPE netlink_conntrack_earlyDrop counter
netlink_conntrack_earlyDrop 0
# HELP netlink_conntrack_error The total of conntrack -S error.
# TYPE netlink_conntrack_error counter
netlink_conntrack_error 0
# HELP netlink_conntrack_found The total of conntrack -S found.
# TYPE netlink_conntrack_found counter
netlink_conntrack_found 0
# HELP netlink_conntrack_ignore The total of conntrack -S ignore.
# TYPE netlink_conntrack_ignore counter
netlink_conntrack_ignore 26332
# HELP netlink_conntrack_insert The total of conntrack -S insert.
# TYPE netlink_conntrack_insert counter
netlink_conntrack_insert 0
# HELP netlink_conntrack_insertFailed The total of conntrack -S insertFailed.
# TYPE netlink_conntrack_insertFailed counter
netlink_conntrack_insertFailed 0
# HELP netlink_conntrack_invalid The total of conntrack -S invalid.
# TYPE netlink_conntrack_invalid counter
netlink_conntrack_invalid 42
# HELP netlink_conntrack_searchRestart The total of conntrack -S searchRestart.
# TYPE netlink_conntrack_searchRestart counter
netlink_conntrack_searchRestart 187

# stop exporter with C-c
% fg
[2]  - running
^C2021/01/25 13:14:29 shutting down
```

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

# run once, get JSON
% sudo ./netlink-conntrack-status
{"found":0,"invalid":42,"ignore":26332,"insert":0,"insert_failed":0,"drop":0,"early_drop":0,"error":0,"search_restart":187}
```
