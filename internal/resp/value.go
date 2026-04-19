package resp

type Value struct {
	Typ    byte
	Str    string
	Num    int64
	Array  []Value
	IsNull bool
}
