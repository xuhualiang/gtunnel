# gtunnel
`gtunnel` is a proxy that translates existing socket into SSL/TLS. It is a `golang` implementation that replaces `stunnel`. We benefit from efficient `goroutine`. we cross plantforms with a few hundred lines of go code.

Unlike `stunnel`, `gtunnel` is simple enough, efficent and scalable. Detailed measurement will be placed.

## socket to socket redirection
```
gtunnel -pair 127.0.0.1:1001-127.0.0.1:1002
```

## socket to SSL/TLS socket redirection
```
gtunnel -pair 127.0.0.1:1001-127.0.0.1:1002:s
```
where `:s` stands SSL/TLS socket. Use this to switch your client into SSL/TLS with minimum change.

## SSL/TLS socket to regular socket redirection
```
gtunnel -pair 127.0.0.1:443:s-127.0.0.1:80 -cert test/cert.pem -key test/key.pem
```
where `:s` stands for SSL/TLS socket. Use this to convert your web server into https.

## generate a testing cert/key pair with openssl
```
openssl req -newkey rsa:2048 -nodes \
    -keyout key.pem                 \
    -x509 -days 365 -out cert.pem   \
    -subj /C=US/ST=Massachusetts/L=Dover/O=Organization/OU=OrganizationUnit/CN=CommonName
```

## performance compare with stunnel (on linux)

### test setup
  - with stunnel front + back end
    - echo server: ./bin/hack-ech
    - stunnel test/st-front.cfg
    - stunnel test/st-back.cfg
    - test case: for i in `seq 4`; do ./bin/hack-test ; done

  - with gtunnel front + back end
    - echo server: ./bin/hack-ech
    - ./bin/gtunnel -pair :10002-127.0.0.1:10001:s
    - ./bin/gtunnel -pair :10001:s-127.0.0.1:10000:p -cert test/cert.pem -key test/key.pem
    - test case: for i in `seq 4`; do ./bin/hack-test ; done

### with the following results (throughput + latency), we/gtunnel beats stunnel in every respect

### what we get from stunnel
```
-bash-4.2$ for i in `seq 4`; do ./bin/hack-test ; done
512.00 MB 372.49 MB/s
0.48 seconds, 4096 messages, 0.12 milli sec/message
512.00 MB 372.61 MB/s
0.43 seconds, 4096 messages, 0.10 milli sec/message
512.00 MB 341.11 MB/s
0.44 seconds, 4096 messages, 0.11 milli sec/message
512.00 MB 330.72 MB/s
0.48 seconds, 4096 messages, 0.12 milli sec/message
```

### what we get from gtunnel
```
-bash-4.2$ for i in `seq 4`; do ./bin/hack-test ; done
512.00 MB 626.60 MB/s
0.42 seconds, 4096 messages, 0.10 milli sec/message
512.00 MB 637.26 MB/s
0.45 seconds, 4096 messages, 0.11 milli sec/message
512.00 MB 627.39 MB/s
0.45 seconds, 4096 messages, 0.11 milli sec/message
512.00 MB 638.92 MB/s
0.38 seconds, 4096 messages, 0.09 milli sec/message
```

## want a mutual and customized certificate verification? Reach out to aihua.yang@aliyun.com
