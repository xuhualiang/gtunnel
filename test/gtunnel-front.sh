opt=root-ca-file:server-name

./bin/gtunnel -pair :10002:p-127.0.0.1:10001:s -verify $opt
