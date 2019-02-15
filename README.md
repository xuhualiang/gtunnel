# gtunnel
`gtunnel` redirects data flow from socket to socket, SSL encap/decap can be applied. This is a Golang implementation in replace of `stunnel`. We don't have the pain of using openssl library, and we benefit from lightweight go-routine.

`gtunnel` has to be very scalable, and measurement will be placed.

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

## generate a testing cert/key pair with openssl
```
openssl req -newkey rsa:2048 -nodes
	-keyout key.pem
	-x509 -days 3650 -out cert.pem
	-subj /C=US/ST=Massachusetts/L=Cambridge/O=Organization/OU=OrganizationUnit/CN=CommonName
```
