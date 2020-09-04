package adapter

import (
	"fmt"
	"strings"
)

// The key is a stringified AbstractOrderedPair.
// Because in Go slices aren't Comparable and thus can't be map keys. I know, I'm angry too.
type AbstractConcreteMap map[string]ConcreteOrderedPair

func NewAbstractConcreteMap() *AbstractConcreteMap {
	acm := map[string]ConcreteOrderedPair{}
	return (*AbstractConcreteMap)(&acm)
}

func (acm AbstractConcreteMap) String() string {
	var sb strings.Builder
	for key, value := range acm {
		sb.WriteString(fmt.Sprintf("%v->%v\n", key, value.String()))
	}
	return sb.String()
}

func (acm *AbstractConcreteMap) AddOPs(abstractOrderedPair AbstractOrderedPair, concreteOrderedPair ConcreteOrderedPair) {
	(*acm)[abstractOrderedPair.String()] = concreteOrderedPair
}

func (acm *AbstractConcreteMap) AddIOs(abstractInputs []AbstractSymbol, abstractOutputs []AbstractSet, concreteInputs []*ConcreteSymbol, concreteOutputs []ConcreteSet) {
	abstractOP := AbstractOrderedPair{abstractInputs: abstractInputs, abstractOutputs: abstractOutputs}
	concreteOP := ConcreteOrderedPair{concreteInputs: concreteInputs, concreteOutputs: concreteOutputs}
	acm.AddOPs(abstractOP, concreteOP)
}
