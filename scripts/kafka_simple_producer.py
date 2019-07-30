import datetime
import json
from kafka import KafkaConsumer, KafkaProducer

d = {
    "private_data": "Your own data here",
    "my_array": ["It", "supports", "array"],
    "my_inner_object": {
        "supported": True,
        "when": datetime.datetime.now().strftime('%Y-%m-%d %H:%M:%S')
    }

}
e = {
    "identity": "user_id_234823",
    "type": "test",
    "ip": "222.333.22.33",
    "primary_key": "unique_id_2852",
    "timestamp": int(datetime.datetime.now().timestamp()) * 1000,
    "data": d,
    "before": None,
    "after": None,

}

if __name__ == '__main__':
    producer = KafkaProducer(bootstrap_servers=['47.100.222.11:30050'])

    for i in range(1):
        ss = json.dumps(e)
        # ss += '\0'
        # producer.send('tech-tech-anlink-web-gateway-201907101551', bytes(ss, 'utf-8'))
        producer.send('anlink', bytes(ss, 'utf-8'))
        producer.flush()
        print(ss)

