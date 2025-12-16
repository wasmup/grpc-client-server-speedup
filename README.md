# grpc-client-server-speedup
improve gRPC response time by 2x

```sh
cd server/
go run .

cd ../client/
go run .

```

## with keepalive
Ping run complete: count=5457, avg_latency=424.742µs 

Ping run complete: count=5460, avg_latency=419.143µs 




```sh
cd uds-server/
go run .

cd ../uds-client/
go run .

```

# Output

```
Running 10000 UDS health checks against /tmp/health.sock...
--- UDS Latency Results ---
Samples: 10000
Min Latency: 4.253µs
Average (Mean): 10.068µs
P95 Latency: 19.569µs
P99 Latency: 28.659µs

```


