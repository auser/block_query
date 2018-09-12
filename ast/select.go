package ast

// Select selects a collection item
func (ast AST) Select(f func(interface{}) interface{}) AST {
	return AST{
		Iterate: func() Iterator {
			next := ast.Iterate()

			return func() (item interface{}, ok bool) {
				var it interface{}
				it, ok = next()
				if ok {
					item = f(it)
				}
				return
			}
		},
	}
}

// SelectKeys allows us to pass a list of keys to select
// from the AST
func (ast AST) SelectKeys(keys ...string) AST {
	return AST{
		Iterate: func() Iterator {
			next := ast.Iterate()

			keySet := make(map[string]bool)
			for _, v := range keys {
				keySet[v] = true
			}

			return func() (item interface{}, ok bool) {
				var it interface{}
				it, ok = next()

				if ok {
					key := it.(KeyValue).Key.(string)
					if keySet[key] {
						item = it
					}
				}

				return
			}
		},
	}
}
