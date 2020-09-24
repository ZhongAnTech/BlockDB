package instruction

import (
	"context"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"strings"
)

// TODO: do not enable permission verification unless you can load all collections at start up
func (t *InstructionExecutor) PermissionVerify(op string, collection string, publickey string) (bool, error) {
	ctx, _ := context.WithTimeout(context.Background(), t.Config.ReadTimeout)
	//切换集合至permissions集合
	response, err := t.storageExecutor.Select(ctx, "permissions", bson.M{"To": publickey}, bson.M{}, 1, 0)
	if err != nil {
		logrus.WithError(err)
		return false, err
	}
	if len(response.Content) == 0 {
		logrus.Debug("没有该用户的权限信息")
		return false, err
	}
	data, err := bson.Marshal(response.Content[0])
	if err != nil {
		logrus.WithError(err).Warn("json转化失败")
		return false, err
	}
	Doc := PermissionsDoc{}
	err = bson.Unmarshal(data, &Doc)
	if err != nil {
		logrus.WithError(err).Warn("结构体转化失败失败")
	}
	//先比较是不是前缀相同
	for _, s := range Doc.CollectionPrefix {
		if strings.HasPrefix(collection, s.Collection+"_") {
			return true, nil
		}
	}
	if op == "create" || op == "update" || op == "read" || op == "delete" {
		for _, t := range Doc.Curd {
			if strings.HasPrefix(collection, t.Collection+"_") {
				return true, nil
			}
		}
	}
	if op == "awared" {
		for _, t := range Doc.Curd {
			if strings.HasPrefix(collection, t.Collection+"_") && t.Isawared == true {
				return true, nil
			}
		}
	}
	return false, nil
}

//权限授予
func (t *InstructionExecutor) PermissionGrant(permissions string, collection string, from string, to string, isawared bool, time int64) (bool, error) {
	ctx, _ := context.WithTimeout(context.Background(), t.Config.ReadTimeout)
	//判断from是否有这个授予集合的权限。
	res1, err := t.PermissionVerify(permissions, collection, from)
	if res1 == false {
		return false, nil
	}
	//判断有没有这个账号
	//TODO:需要用户账号表
	res, err := t.storageExecutor.Select(ctx, "account", bson.M{"account": to}, bson.M{}, 0, 0)
	if err != nil {
		logrus.Fatal("查找账号失败")
	}
	if len(res.Content[0]) == 0 {
		logrus.Fatal("没有该账号信息")
	}
	//判断权限表中有没有这个账号
	res, err = t.storageExecutor.Select(ctx, "permissions", bson.M{"to": to}, bson.M{}, 0, 0)
	if err != nil {
		logrus.Fatal("查找账号失败")
	}
	//此时创建该用户的权限表并赋予权限
	if len(res.Content[0]) == 0 {
		var collectionPrefix PermissionsDetail
		var curd PermissionsDetail
		if permissions == "curd" {
			curd.Collection = collection
			curd.From = from
			curd.Isawared = isawared
			curd.Timestamp = time
		} else {
			collectionPrefix.Collection = collection
			collectionPrefix.Isawared = true
			collectionPrefix.From = from
			collectionPrefix.Timestamp = time
		}
		M := bson.M{
			"collection_prefix": collectionPrefix,
			"curd":              curd,
			"to":                to,
		}
		res, err := t.storageExecutor.Insert(ctx, "permissions", M)
		if err != nil {
			logrus.Fatal("权限赋予失败")
		}
		if res != " " {
			return true, nil
		}
		return false, nil
	} else {
		//更新权限表的集合
		//取出目标的权限集
		res, err := t.storageExecutor.Select(ctx, "permissions", bson.M{"to": to}, bson.M{}, 0, 0)
		if err != nil {
			logrus.WithError(err)
			return false, err
		}
		if len(res.Content) == 0 {
			logrus.WithError(err).Warn("没有该用户的权限信息")
			return false, err
		}
		data, err := bson.Marshal(res.Content[0])
		if err != nil {
			logrus.WithError(err).Warn("json转化失败")
			return false, err
		}
		Doc := PermissionsDoc{}
		err = bson.Unmarshal(data, &Doc)
		if err != nil {
			logrus.WithError(err).Warn("结构体转化失败失败")
			return false, err
		}
		//判断是否已经有这个权限了
		for _, value := range Doc.CollectionPrefix {
			if strings.HasPrefix(collection, value.Collection+"_") {
				return true, nil
			}
		}
		for _, value := range Doc.Curd {
			if strings.HasPrefix(collection, value.Collection+"_") && value.Isawared == isawared {
				return true, nil
			}
		}
		var add PermissionsDetail
		if permissions == "create" || permissions == "update" || permissions == "read" || permissions == "delete" {
			add.Collection = collection
			add.Isawared = isawared
			add.Timestamp = time
			add.From = from
			newCurd := append(Doc.Curd, add)
			res, err := t.storageExecutor.Update(ctx, "permissions", bson.M{"to": to}, bson.M{"curd": newCurd}, "set")
			if err != nil {
				logrus.WithError(err)
			}
			if res == 1 {
				return true, nil
			}
			return false, nil
		} else {
			add.Collection = collection
			add.Isawared = true
			add.Timestamp = time
			add.From = from
			newCollectionPrefix := append(Doc.CollectionPrefix, add)
			res, err := t.storageExecutor.Update(ctx, "permissions", bson.M{"to": to}, bson.M{"collectionPrefix": newCollectionPrefix}, "set")
			if err != nil {
				logrus.WithError(err)
				return false, err
			}
			if res == 1 {
				return true, nil
			}
			return false, nil
		}
	}
}

//权限撤销
func (t *InstructionExecutor) PermissionCancel(permissions string, collection string, to string, from string) (bool, error) {
	ctx, _ := context.WithTimeout(context.Background(), t.Config.ReadTimeout)
	//判断from是否有这个授予集合的权限。
	res1, err := t.PermissionVerify(permissions, collection, from)
	if res1 == false {
		return false, nil
	}
	//判断有没有这个账号
	res, err := t.storageExecutor.Select(ctx, "account", bson.M{"account": to}, bson.M{}, 0, 0)
	if err != nil {
		logrus.Fatal("查找账号失败")
		return false, err
	}
	if len(res.Content[0]) == 0 {
		logrus.Fatal("没有该账号信息")
		return false, err
	}
	//判断权限表中有没有这个账号
	res, err = t.storageExecutor.Select(ctx, "permissions", bson.M{"to": to}, bson.M{}, 0, 0)
	if err != nil {
		logrus.Fatal("查找账号失败")
		return false, err
	}
	if len(res.Content) == 0 {
		return false, nil
	}
	data, err := bson.Marshal(res.Content[0])
	if err != nil {
		logrus.WithError(err).Warn("json转化失败")
		return false, err
	}
	Doc := PermissionsDoc{}
	err = bson.Unmarshal(data, &Doc)
	if err != nil {
		logrus.WithError(err).Warn("结构体转化失败失败")
		return false, err
	}
	var point = -1
	for index, value := range Doc.CollectionPrefix {
		if strings.HasPrefix(collection, value.Collection+"_") {
			point = index
			break
		}
	}
	if point != -1 {
		old := Doc.CollectionPrefix
		new := append(old[:point], old[point+1:]...)
		res2, err2 := t.storageExecutor.Update(ctx, "permissions", bson.M{"to": to}, bson.M{"collectionPrefix": new}, "set")
		if err2 != nil {
			logrus.WithError(err)
		}
		if res2 == 1 {
			return true, nil
		}
		return false, nil
	}
	for index, value := range Doc.Curd {
		if strings.HasPrefix(collection, value.Collection+"_") {
			point = index
			break
		}
	}
	if point != -1 {
		old := Doc.Curd
		new := append(old[:point], old[point+1:]...)
		res2, err2 := t.storageExecutor.Update(ctx, "permissions", bson.M{"to": to}, bson.M{"curd": new}, "set")
		if err2 != nil {
			logrus.WithError(err)
		}
		if res2 == 1 {
			return true, nil
		}
		return false, nil
	}
	return false, nil
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
