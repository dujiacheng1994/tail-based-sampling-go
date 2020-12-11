go build

cd script
open -a Terminal.app ./evaluate.sh &
sleep 10s
open -a Terminal.app ./8000.sh &
open -a Terminal.app ./8001.sh &
open -a Terminal.app ./8002.sh &

