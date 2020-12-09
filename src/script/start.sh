cd script

./8000.sh &
./8001.sh &
./8002.sh &
sleep 8s
open -a Terminal.app ./evaluate.sh &