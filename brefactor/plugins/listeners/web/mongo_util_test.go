package web

//func TestMgo(t *testing.T) {
//	mgo:= InitMgo("mongodb://localhost:27017","test",[]string{"coll", "coll1"})
//	hex1, err:=mgo.Insert("coll", bson.D{{"a",1},{"b","abc"}})
//	if err != nil {
//		t.Error("fail to insert: ", err)
//	}
//
//	hex2, err := mgo.Insert("coll",bson.D{{"a",2},{"b","efg"}})
//	if err != nil {
//		t.Error("fail to insert: ", err)
//	}
//
//	_, err =mgo.Update("coll",bson.D{{"a",1},{"b","abc"}},bson.D{{"a",3},{"b","klm"}},"set")
//	if err != nil {
//		t.Error("fail to update: ", err)
//	}
//
//	response, err := mgo.Select("coll", bson.D{{"a",bson.D{{"$ne",nil}}}}, bson.D{{"a",-1}},0,0)
//	if err != nil {
//		t.Error("fail to select: ", err)
//	}
//	fmt.Println(response)
//
//	_, err =mgo.Delete("coll", hex1)
//	if err != nil {
//		t.Error("fail to delete: ", err)
//	}
//
//	_, err =mgo.Delete("coll", hex2)
//	if err != nil {
//		t.Error("fail to delete: ", err)
//	}
//
//	hex3, err:=mgo.Insert("coll1", bson.D{{"a",1},{"b","abc"}})
//	if err != nil {
//		t.Error("fail to insert: ", err)
//	}
//
//	_, err =mgo.Delete("coll1", hex3)
//	if err != nil {
//		t.Error("fail to delete: ", err)
//	}
//}
