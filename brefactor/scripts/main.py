import json

from requests import Session

s = Session()

if __name__ == '__main__':
    d1 = {"op": "create_collection", "collection": "sample_collection",
          "feature": {"allow_update": True, "allow_delete": True, "cooperate": True,
                      "allow_insert_members": ["0x123456", "0x123456", "0x123456", "0x123456"],
                      "allow_update_members": ["0x123456", "0x123456", "0x123456", "0x123456"],
                      "allow_delete_members": ["0x123456", "0x123456", "0x123456", "0x123456"]}}
    opstr = json.dumps(d1)

    d = {'op_str': opstr,
         "public_key": "0x769153474351324", "signature": "0x169153474351324", "op_hash": "0x53452345"}
    url = 'http://localhost:8080/audit'
    resp = s.post(url, json=json.dumps(d))
    print(resp.text)
