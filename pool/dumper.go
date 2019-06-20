package pool

type Dumper struct {
	Name string
}

func (d *Dumper) Write(p []byte) (n int, err error) {
	//fmt.Println(d.Name)
	//fmt.Println(hex.Dump(p))
	return len(p), nil
}
