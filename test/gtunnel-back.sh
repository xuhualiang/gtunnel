
crt=cert-file
key=key-file

./bin/gtunnel -pair :10001:s-127.0.0.1:10000:p -cert $crt -key $key
