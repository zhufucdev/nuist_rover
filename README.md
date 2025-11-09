# NUIST Rover

Get through on the router.

## Build for OpenWrt

The following command creates a minimized build for OpenWrt
running on MIPS processors.

```shell
GOOS=linux GOARCH=mipsle GOMIPS=softfloat go build -ldflags "-w -s" .
```

## Configuration

File defaults to `/etc/nuistrover/config.toml`, and can be
specified by passing the `--configuration` or `-c` flag.

Format is as follows.

```toml
serverurl = "<your server>"
verbose = "<one of 'log', 'info', 'exception', 'error'>"
retry = 3
testinterval = "5m"
retryinterval = "30s"

[onlinecheck]
enabled = true      # enable online check
method = "portal"   # or "ping"
host = "1.2.4.8"    # for ping method
count = 4           # number of pings
threshold = 0.25    # success rate threshold (0.25 = 25%)

[accounts.wan]
username = "<your account>"
password = "<your password>"
isp = "<one of 'internal', 'telecom', 'mobile', 'unicom'>"

# Specific more accounts for multi-dial
[accounts.wanmac0]
username = "..."
password = "..."
isp = "telecom"

[accounts.wanmac1]
username = "..."
password = "..."
isp = "mobile"
```
