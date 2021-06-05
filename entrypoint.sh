#!/usr/bin/env bash

set -e

REGION=$REGION
SER_USER=$SER_USER
SER_SECRET=$SER_SECRET
SIP_SERVICE=$SIP_SERVICE
KAMGO_SERVICE=$KAMGO_SERVICE
NUMBER_SERVICE=$NUMBER_SERVICE
HEARTBEAT_SERVICE=$HEARTBEAT_SERVICE
RATING_ROUTING_SERVICE=$RATING_ROUTING_SERVICE
CDR_SERVICE=$CDR_SERVICE
RECORDING_SERVICE=$RECORDING_SERVICE

pushd /etc
		echo "[ENTRYPOINT] - Updating config.toml"
		sed -i "s/<REGION>/$REGION/g w /dev/stdout" config.toml
		sed -i "s/<PRIV_USER>/$SER_USER/g w /dev/stdout" config.toml
		sed -i "s/<PRIV_SECRET>/$SER_SECRET/g w /dev/stdout" config.toml
		sed -i "s#<HEARTBEAT_SERVICE>#$HEARTBEAT_SERVICE#g w /dev/stdout" config.toml
		sed -i "s#<RATE_ROUTE_SERVICE>#$RATING_ROUTING_SERVICE#g w /dev/stdout" config.toml
		sed -i "s#<NUMBER_SERVICE>#$NUMBER_SERVICE#g w /dev/stdout" config.toml
		sed -i "s#<SIP_SERVICE>#$SIP_SERVICE#g w /dev/stdout" config.toml
		sed -i "s#<KAMGO_SERVICE>#$KAMGO_SERVICE#g w /dev/stdout" config.toml
		sed -i "s#<RECORDING_SERVICE>#$RECORDING_SERVICE#g w /dev/stdout" config.toml
popd

/app
