package instruction

// TODO: do not enable permission verification unless you can load all collections at start up
func (t *InstructionExecutor) PermissionVerify(op string, collection string, publickey string) bool {
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
