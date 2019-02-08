# gtunnel
socket tunnel, a replacement of stunnel in Golang

## what is stunnel
https://www.stunnel.org, stunnel is to translate regular socket and tls/ssl socket. It is in C/libopenssl which complicates the program and hard to handle the error cases. It is based on `select` model which makes it less efficient.

