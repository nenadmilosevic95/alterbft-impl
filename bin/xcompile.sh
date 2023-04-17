#/bin/bash

AGENT_DIR="./agent"
CLIENT_DIR="./client"
RENDEZVOUS_DIR="./rendezvous"

for DIR in $AGENT_DIR $CLIENT_DIR $RENDEZVOUS_DIR
do
	DIRNAME=$(basename $DIR)
	echo "+Compiling $DIRNAME..."
	pushd $DIR > /dev/null
	go build || exit 2
	popd > /dev/null
	ls -l $DIR/$DIRNAME
done
