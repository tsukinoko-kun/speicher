// non-sucking database
//
// Example:
//
//	type Foo struct {
//		A string `json:"a"`
//		B int    `json:"b"`
//	}
//
//	list, err := speicher.LoadList[Foo]("./data/foo.json")
//
//	if err != nil {
//		panic(err)
//	}
//
//	list.Lock()
//	list.Append(Foo{A: "baz", B: 42})
//	list.Unlock()
//
//	list.RLock()
//	if foo, ok := list.Find(func(f Foo) bool { return f.B > 40 }); ok {
//		fmt.Println(foo.A)
//	}
//	list.RUnlock()
//
//	list.Save()
package speicher
