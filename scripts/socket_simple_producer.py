import datetime
import json
import socket

e = {
    "identity": "user_id_234823",
    "type": "test",
    "ip": "222.333.22.33",
    "primary_key": "unique_id_2852",
    "timestamp": int(datetime.datetime.now().timestamp()) * 1000,
    "data": {},
    "before": None,
    "after": None,

}


def ts(to_send: str):
    s.send(to_send.encode())
    data = s.recv(1024).decode()
    print(data)


if __name__ == '__main__':
    s = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
    host = "127.0.0.1"
    port = 28019
    s.connect((host, port))
    ts(json.dumps(e) + '\0')
    s.close()