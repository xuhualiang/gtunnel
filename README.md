# gtunnel
`gtunnel` is a proxy that translates existing client/server into SSL. It is a `golang` implementation that replaces `stunnel`. We benefit from efficient `goroutine`. we cross plantforms with a few hundred lines of go code.

As of Feb 19, 2019, `gtunnel` can work as a `load balancer` with simple `round robin` rotation method.

Unlike `stunnel`, `gtunnel` is simple enough, efficent and scalable. Detailed measurement will be placed. Contact `hualiang.xu@gmail.com` for more.


## socket to socket redirection
```
[sock-to-sock]
connect = pp/localhost:1001/localhost:1002
```
where `pp` stands for plain to plain socket

## regular socket to SSL socket redirection
```
[sock-to-ssl]
connect = ps/localhost:1001/localhost:1002
cert = test/cert.pem
key  = test/key.pem
```
where `ps` stands for plain socket to SSL socket. Typical usage of this is to switch your old app client into SSL with minimum change.

## SSL socket to regular socket redirection
```
[ssl-to-socket]
connect = sp/localhost:1001/localhost:1002
cert = test/cert.pem
key  = test/key.pem
```
where `sp` stands for plain socket to SSL socket. Typical usage of this is to switch your old app server into SSL with minimum change.

## load balance
```
[my-load-balancer]
connect = sp/localhost:1001/localhost:20001,localhost:20002
cert = test/cert.pem
key  = test/key.pem
```

## generate a testing cert/key pair with openssl
```
openssl req -newkey rsa:2048 -nodes
	-keyout key.pem
	-x509 -days 3650 -out cert.pem
	-subj /C=US/ST=Massachusetts/L=Cambridge/O=Organization/OU=OrganizationUnit/CN=CommonName
```
