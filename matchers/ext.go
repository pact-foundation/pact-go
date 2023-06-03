package matchers

import "reflect"

type CustomMatchStructV2Args struct {
	StrategyFunc FieldStrategyFunc
}

type customMatchStructV2 struct {
	matchStruct matchStructV2
}

func NewCustomMatchStructV2(args CustomMatchStructV2Args) customMatchStructV2 {
	return customMatchStructV2{matchStruct: matchStructV2{args.StrategyFunc}}
}

func (m *customMatchStructV2) MatchV2(src interface{}) Matcher {
	return m.matchStruct.match(reflect.TypeOf(src), getDefaults())
}
