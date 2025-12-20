kafka-topics --help

kafka-topics --list --bootstrap-server localhost:9092

kafka-topics --create \
  --topic execution-tasks \
  --bootstrap-server localhost:9092
