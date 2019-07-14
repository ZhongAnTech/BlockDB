import json
from kafka import KafkaConsumer, KafkaProducer

d = {
    "thread": "http-nio-8080-exec-5",
    "level": "INFO",
    "loggerName": "auditing",
    "message": "TTT",
    "endOfBatch": False,
    "loggerFqcn": "org.apache.logging.slf4j.Log4jLogger",
    "instant": {
        "epochSecond": 1561375556,
        "nanoOfSecond": 447000000
    },
    "contextMap": {
        "id": "122",
        "user": "XXX"
    },
    "threadId": 0,
    "threadPriority": 5
}
e = {
    "Identity": "hahaha",
    "Type": "mongodb",
    "Ip": "172.28.152.101",
    "PrimaryKey": "nothing",
    "TimeStamp": 1561375556,
    "Data": d,
    "Before": "og",
    "After": "nothing",

}

if __name__ == '__main__':
    producer = KafkaProducer(bootstrap_servers='172.28.152.102:30092')

    for i in range(1):
        ss = json.dumps(e)
        # ss += '\0'
        # producer.send('tech-tech-anlink-web-gateway-201907101551', bytes(ss, 'utf-8'))
        producer.send('anlink', bytes(ss, 'utf-8'))
