package instruction

import (
	"context"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"strings"
)

// TODO: do not enable permission verification unless you can load all collections at start up
func (t *InstructionExecutor) PermissionVerify(op string, collection string, publickey string) bool {
	ctx, _ := context.WithTimeout(context.Background(), t.Config.ReadTimeout)
	//切换集合至permissions集合
	response,err:=t.storageExecutor.Select(ctx,CommandCollection,bson.M{"To":publickey},bson.M{},1,0)
	if err!=nil {
		logrus.WithError(err).Warn("权限验证失败")
		return false
	}
	if len(response.Content) == 0 {
		logrus.Debug("没有该用户的权限信息")
		return false
	}
	data, err := bson.Marshal(response.Content[0])
	if err !=nil{
		logrus.Fatal("转化为对应json失败")
		return false;
	}
	Doc:=PermissionsDoc{}
	err = bson.Unmarshal(data, &Doc)
	if err !=nil {
		logrus.Fatal("转化为对应Permissions结构体失败")
	}
	//先比较是不是前缀相同
	for _, s:=range Doc.CollectionPrefix{
		if(strings.HasPrefix(collection,s)){
			return true
		};
	}
	//TODO：在比较权限是否是curd操作
	
	//比较前缀是否和Doc.Curd的相同
	for _, t:=range Doc.Curd{
		if(strings.HasPrefix(collection,t)){
			return true
		};
	}
	return true
}

//权限验证
//func (t *InstructionExecutor) Check(op string, collection string, publickey string) bool {
//	flag := false
//outside:
//	for _, coll := range Colls {
//		if coll.Collection == collection {
//			if coll.PublicKey == publickey {
//				switch op {
//				case Insert, UpdateCollection:
//					flag = true
//				case Update:
//					if coll.Feature["allow_update"].(bool) == true {
//						flag = true
//					}
//				case Delete:
//					if coll.Feature["allow_delete"].(bool) == true {
//						flag = true
//					}
//				}
//			} else if coll.Feature["cooperate"].(bool) == true {
//				switch op {
//				case Insert:
//					allows := coll.Feature["allow_insert_members"].([]string)
//					for _, pk := range allows {
//						if pk == publickey {
//							flag = true
//						}
//					}
//				case Update:
//					allows := coll.Feature["allow_update_members"].([]string)
//					for _, pk := range allows {
//						if pk == publickey {
//							flag = true
//						}
//					}
//				case Delete:
//					allows := coll.Feature["allow_delete_members"].([]string)
//					for _, pk := range allows {
//						if pk == publickey {
//							flag = true
//						}
//					}
//				}
//			}
//			break outside
//		}
//	}
//	return flag
//}
