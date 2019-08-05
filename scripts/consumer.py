from kafka import KafkaConsumer

if __name__ == '__main__':
    consumer = KafkaConsumer('anlink', bootstrap_servers=['47.100.222.11:30040'])
    for message in consumer:
        # message value and key are raw bytes -- decode if necessary!
        # e.g., for unicode: `message.value.decode('utf-8')`
        print ("%s:%d:%d: key=%s value=%s" % (message.topic, message.partition,
                                              message.offset, message.key,
                                              message.value))
