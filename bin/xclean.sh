#/bin/bash

AGENT_DIR="./agent"
CLIENT_DIR="./client"
RENDEZVOUS_DIR="./rendezvous"

for DIR in $AGENT_DIR $CLIENT_DIR $RENDEZVOUS_DIR
do
	DIRNAME=$(basename $DIR)
	rm -fv $DIR/$DIRNAME
done

LOG_DIR="./logs"
RENDEZVOUS_LOG="./rendezvous.log"

rm -fv $RENDEZVOUS_LOG
rm -fv $LOG_DIR/a.*
rm -fv $LOG_DIR/c.*
rm -fv $LOG_DIR/deliveries.*
