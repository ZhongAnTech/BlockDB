import datetime
import json

import requests

# host = "http://47.100.122.212:30047"
host = "http://127.0.0.1:28020"

if __name__ == '__main__':
    e = {
        "identity": "user_id_234823",
        "type": "test",
        "ip": "222.333.22.33",
        "primary_key": "unique_id_2852",
        "timestamp": int(datetime.datetime.now().timestamp()) * 1000,
        "data": {"yourdata": 444,"another": {"yes": "no"}},
    }

    s = requests.Session()
    # for i in range(1):
    #     resp = s.post(host + '/doc', data=json.dumps(e))
    #     print(resp.text)

    # resp = s.get('http://47.100.122.212:30047/docs/0x9bd1cc3bfd9c502548ced94585720b4064dceee018c23400b9569d1cddbaa867')
    # print(resp.text)
    # resp = s.get('http://127.0.0.1:28020/query?height=5169')
    resp = s.get(host + '/query?int-data.data.yourdata=444')
    print(json.dumps(resp.json(), indent=4))
