go build

cd script
open -a Terminal.app ./8000.sh &
open -a Terminal.app ./8001.sh &
open -a Terminal.app ./8002.sh &

sleep 8s
open -a Terminal.app ./evaluate.sh &