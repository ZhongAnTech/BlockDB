package instruction

import (
	"crypto/sha256"
	"encoding/json"
	"log"
	"strconv"
	"time"
)

func Execute(op int,instruction string){
	timestamp:=strconv.FormatInt(time.Now().Unix(),10)
	switch op {
	case CreateCollection:
		com:=&BlockDBCommandCollection{}
		err:=json.Unmarshal([]byte(instruction),com)
		if err != nil{
			log.Fatal("failed to unmarshal create_collection command.")
			break
		}
		//TODO: Verification of signature
		com.Timestamp=timestamp
		//计算hash
		hash,err:=getHash(com.Feature)
		if err != nil{
			log.Println("failed to marshal insert data.")
			break
		}
		com.Hash=hash
		//缓存
		Colls=append(Colls,com)
		//操作记录
		OpRecord(op,hash,com.Collection,timestamp,com.Feature,com.PublicKey,com.Signature)
		//TODO: CreateCollection(BlockDataBase,com.Collection)
		//TODO: Insert(CollIndexDataBase,com)
		//历史版本记录
		HistoryRecord(hash,com.Collection,timestamp,com.Feature,com.PublicKey,com.Signature)
		//审计记录
		Audit(op,hash,com.Collection,timestamp,com.Feature,com.PublicKey,com.Signature)

	case UpdateCollection:
		com:=&BlockDBCommandCollection{}
		err:=json.Unmarshal([]byte(instruction),com)
		if err != nil{
			log.Println("failed to unmarshal update_collection command.")
			break
		}
		//TODO: Verification of signature
		//权限验证
		if Check(op,com.Collection,com.PublicKey){
			if UpdateCollectionFeatures(com.Collection,com.Feature){
				com.Timestamp=timestamp
				OpRecord(op,com.Hash,com.Collection,timestamp,com.Feature,com.PublicKey,com.Signature)
				//TODO: Insert(CollIndexDataBase,com.Collection,com)
				HistoryRecord(com.Hash,com.Collection,timestamp,com.Feature,com.PublicKey,com.Signature)
				Audit(op,com.Hash,com.Collection,timestamp,com.Feature,com.PublicKey,com.Signature)
			}else {
				log.Println("collection "+com.Collection+" doesn't exist.")
			}
		}else{
			log.Println("update_collection permission denied")
		}

	case Insert:
		com:=&BlockDBCommandInsert{}
		err:=json.Unmarshal([]byte(instruction),com)
		if err != nil{
			log.Println("failed to unmarshal insert command.")
			break
		}
		//TODO: Verification of signature
		if Check(op,com.Collection,com.PublicKey){
			hash,err:=getHash(com.Data)
			if err != nil{
				log.Println("failed to marshal insert data.")
				break
			}
			com.Hash=hash
			com.Timestamp=timestamp
			OpRecord(op,hash,com.Collection,timestamp,com.Data,com.PublicKey,com.Signature)
			//TODO: Insert(BlockDataBase,com.Collection,com)
			HistoryRecord(hash,com.Collection,timestamp,com.Data,com.PublicKey,com.Signature)
			Audit(op,hash,com.Collection,timestamp,com.Data,com.PublicKey,com.Signature)
		}else{
			log.Println("insert permission denied")
		}

	case Update:
		com:=&BlockDBCommandUpdate{}
		err:=json.Unmarshal([]byte(instruction),com)
		if err != nil{
			log.Println("failed to unmarshal update command.")
			break
		}
		//TODO: Verification of signature
		if Check(op,com.Collection,com.PublicKey){
			com.Timestamp=timestamp
			hash:=com.Query["_hash"]
			data:=make(map[string]interface{})
			data["query"]=com.Query
			data["set"]=com.Set
			data["unset"]=com.Unset
			OpRecord(op,hash,com.Collection,timestamp,data,com.PublicKey,com.Signature)
			//TODO: Update(BlockDataBase,com.Collection,com)
			//TODO: history
			//审计记录
			Audit(op,hash,com.Collection,timestamp,data,com.PublicKey,com.Signature)
		}else{
			log.Println("update permission denied")
		}

	case Delete:
		com:=&BlockDBCommandDelete{}
		err:=json.Unmarshal([]byte(instruction),com)
		if err != nil{
			log.Println("failed to unmarshal delete command.")
			break
		}
		//TODO: Verification of signature
		//权限验证
		if Check(op,com.Collection,com.PublicKey){
			com.Timestamp=timestamp
			hash:=com.Query["_hash"]
			data:=make(map[string]interface{})
			data["query"]=com.Query
			//TODO: op records
			OpRecord(op,hash,com.Collection,timestamp,data,com.PublicKey,com.Signature)
			//TODO: delete(BlockDataBase,com.Collection,com)
			//TODO: history
			Audit(op,hash,com.Collection,timestamp,data,com.PublicKey,com.Signature)
		}else{
			log.Println("delete permission denied")
		}


	case CreateIndex:
		com:=&BlockDBCommandIndex{}
		err:=json.Unmarshal([]byte(instruction),com)
		if err != nil{
			log.Println("failed to unmarshal create_index command.")
			break
		}
		//TODO: Verification of signature
		com.Timestamp=timestamp
		//TODO: op records
		data:=make(map[string]interface{})
		data["index"] = com.Index
		OpRecord(op,"",com.Collection,timestamp,data,com.PublicKey,com.Signature)
		Indexes=append(Indexes,com)
		//TODO: CreateIndex(BlockDataBase,com.Collection,com.Index)
		//TODO: history
		//TODO: audit

	case DropIndex:
		com:=&BlockDBCommandIndex{}
		err:=json.Unmarshal([]byte(instruction),com)
		if err != nil{
			log.Println("failed to unmarshal drop_index command.")
			break
		}
		//TODO: Verification of signature
		com.Timestamp=timestamp
		//TODO: op records
		//TODO: DropIndex(BlockDataBase,com.Collection,com.Index)
		//TODO: history
		//TODO: audit

	}
}

//计算hash
func getHash(data map[string]interface{}) (string,error){
	bytes,err:=json.Marshal(data)
	if err != nil{
		return "",err
	}
	hash := sha256.Sum256(bytes)
	return string(hash[:]),nil
}

//权限验证
func Check(op int,collection string,publickey string)bool{
	flag:=false
outside:
	for _,coll := range Colls {
		if coll.Collection == collection {
			if coll.PublicKey == publickey {
				switch op {
				case Insert,UpdateCollection:
					flag=true
				case Update:
					if coll.Feature["allow_update"].(bool) == true{
						flag=true
					}
				case Delete:
					if coll.Feature["allow_delete"].(bool) ==true{
						flag=true
					}
				}
			} else if coll.Feature["cooperate"].(bool) == true {
				switch op{
				case Insert:
					allows := coll.Feature["allow_insert_members"].([]string)
					for _, pk := range allows {
						if pk == publickey {
							flag=true
						}
					}
				case Update:
					allows := coll.Feature["allow_update_members"].([]string)
					for _, pk := range allows {
						if pk == publickey {
							flag=true
						}
					}
				case Delete:
					allows := coll.Feature["allow_delete_members"].([]string)
					for _, pk := range allows {
						if pk == publickey {
							flag=true
						}
					}
				}
			}
			break outside
		}
	}
	return flag
}

//更新Coll
func UpdateCollectionFeatures(collection string,feature map[string]interface{}) bool{
	flag := false
	for _,coll := range Colls{
		if coll.Collection == collection{
			coll.Feature = feature
			flag=true
			break
		}
	}
	return flag
}