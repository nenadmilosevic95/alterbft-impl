#!/bin/bash

LOG_DIR="./logs"
AGENT_DIR="./agent"
RENDEZVOUS_DIR="./rendezvous"
RENDEZVOUS_LOG="./rendezvous.log"

N=$1
shift
if [ -z "$N" ]
then
	echo "Usage: $0 <number of server agents> [<agents parameters>]"
	exit 1
fi

kill_all() {
	echo "+Killing processes"
	killall $@ rendezvous
	killall $@ agent
}

ctrlc_trap() {
	echo -e "\n+Trapped CTRL+C"
	kill_all -9 -v
	exit 3
}

trap 'ctrlc_trap' SIGINT

# 0. Kill possible running processes
kill_all -9

# 1. Compile and start a rendezvous server in background

pushd $RENDEZVOUS_DIR > /dev/null
go build || exit 2
popd > /dev/null

echo "+Starting rendezvous"
echo "" > $RENDEZVOUS_LOG
$RENDEZVOUS_DIR/rendezvous -o $RENDEZVOUS_LOG &
while ! grep "Rendezvous address:" $RENDEZVOUS_LOG
do
	sleep 1
done

# 2. Create log directory if needed, and clean log files

LOG_FLAG_FOUND=
for PAR in $@
do
	if [ "$PAR" == "-dir" ]; then
		LOG_FLAG_FOUND=1
	elif [ -n "$LOG_FLAG_FOUND" ]; then
		LOG_FLAG_FOUND=
		LOG_DIR=$PAR
	fi
done

echo "+Log directory set to: $LOG_DIR"
mkdir -pv $LOG_DIR
rm -fv $LOG_DIR/deliveries.*

# 3. Compile and start $N-1 server agents in background

pushd $AGENT_DIR > /dev/null
go build || exit 2
popd > /dev/null


ID=0
while [ "$ID" -lt "$((N-1))" ]
do
	echo -e "+Starting server agent p$ID in background"
	$AGENT_DIR/agent -i $ID -n $N -dir $LOG_DIR $@ > $LOG_DIR/a.$ID 2>&1 &

	ID=$((ID+1))
done

# 4. Start a sever server in foreground

echo "+Starting server agent p$ID"
$AGENT_DIR/agent -i $ID -n $N -dir $LOG_DIR $@ > $LOG_DIR/a.$ID 2>&1

# 5. Kill rendezvous and wait for agents to finish

killall -9 rendezvous
rm -f $RENDEZVOUS_LOG
wait

# 6. Summary of produced files

echo "+Generated files:"
for ID in `seq 0 $((N-1))`
do
	wc -l $LOG_DIR/deliveries.$ID
done
