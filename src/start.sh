go build
./src "8002" &
./src "8001" &
./src "8000" &
java -Dserver.port=9000 -DcheckSumPath=/Users/dujiacheng.jason/Documents/AliCompetition/checkSumSmall.data -jar /Users/dujiacheng.jason/Documents/AliCompetition/scoring-1.0-SNAPSHOT.jar