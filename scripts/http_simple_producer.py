import datetime
import json

import requests

if __name__ == '__main__':
    e = {
        "identity": "user_id_234823",
        "type": "test",
        "ip": "222.333.22.33",
        "primary_key": "unique_id_2852",
        "timestamp": int(datetime.datetime.now().timestamp()) * 1000,
        "data": "test",
    }

    s = requests.Session()
    # for i in range(1000):
    #     resp = s.post('http://127.0.0.1:28020/doc', data=json.dumps(e))
    #     print(resp.text)

    resp = s.get('http://127.0.0.1:28020/docs/0xe9a3265a3f532c52fa1f96fa107bdffe9eabc1290337879aa217b11105fd8ee6')
    print(resp.text)
    # resp = s.get('http://127.0.0.1:28020/query?height=395799')
    # resp = s.get('http://127.0.0.1:28020/query?int-height=395799')
    # print(json.dumps(resp.json(), indent=4))
