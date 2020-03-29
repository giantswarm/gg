package formatter

import (
	"bytes"
	"container/list"
	"encoding/json"
	"fmt"
	"io"

	"github.com/giantswarm/microerror"
)

// KVPair for initializing from a list of key-value pairs, or for looping
// entries in the same order.
type KVPair struct {
	Key   string
	Value interface{}
}

// OrderedMap has similar operations as the default map, but maintains the order
// of inserted keys. Similar to map, all single key operations e.g. get, set and
// delete runs at O(1). Although the JSON spec says the keys order of an object
// should not matter, sometimes the order of JSON objects and their keys matters
// when printing them for humans. Therefore we have to maintain the object keys
// in the same order as they come in.
//
// Disclaimer, same as Go's default map, OrderedMap is not safe for concurrent
// use. If you need atomic access, may use a sync.Mutex to synchronize.
//
// More references may be found below.
//
//     Go maps in action     https://blog.golang.org/go-maps-in-action
//     JSON and Go           https://blog.golang.org/json-and-go
//     Go-Ordered-JSON       https://github.com/virtuald/go-ordered-json
//     Python OrderedDict    https://github.com/python/cpython/blob/2.7/Lib/collections.py#L38
//     port OrderedDict      https://github.com/cevaris/ordered_map
//     original proposal     https://gitlab.com/c0b/go-ordered-json/-/blob/49bbdab258c2e707b671515c36308ea48134970d/ordered.go
//
type OrderedMap struct {
	m    map[string]interface{}
	l    *list.List
	keys map[string]*list.Element
}

func NewOrderedMap() *OrderedMap {
	return &OrderedMap{
		m:    make(map[string]interface{}),
		l:    list.New(),
		keys: make(map[string]*list.Element),
	}
}

func NewOrderedMapFromKVPairs(pairs []*KVPair) *OrderedMap {
	om := NewOrderedMap()
	for _, pair := range pairs {
		om.Set(pair.Key, pair.Value)
	}
	return om
}

func (om *OrderedMap) Delete(key string) (value interface{}, ok bool) {
	value, ok = om.m[key]
	if ok {
		om.l.Remove(om.keys[key])
		delete(om.keys, key)
		delete(om.m, key)
	}
	return
}

func (om *OrderedMap) EntriesIter() func() (*KVPair, bool) {
	e := om.l.Front()
	return func() (*KVPair, bool) {
		if e != nil {
			key := e.Value.(string)
			e = e.Next()
			return &KVPair{key, om.m[key]}, true
		}
		return nil, false
	}
}

func (om *OrderedMap) EntriesReverseIter() func() (*KVPair, bool) {
	e := om.l.Back()
	return func() (*KVPair, bool) {
		if e != nil {
			key := e.Value.(string)
			e = e.Prev()
			return &KVPair{key, om.m[key]}, true
		}
		return nil, false
	}
}

func (om *OrderedMap) Get(key string) string {
	return om.m[key].(string)
}

func (om *OrderedMap) GetValue(key string) (value interface{}, ok bool) {
	value, ok = om.m[key]
	return
}

func (om *OrderedMap) Has(key string) bool {
	_, ok := om.m[key]
	return ok
}

func (om *OrderedMap) Keys() []string {
	var keys []string
	for e := om.l.Front(); e != nil; e = e.Next() {
		keys = append(keys, e.Value.(string))
	}

	return keys
}

func (om *OrderedMap) Len() int {
	return len(om.m)
}

func (om *OrderedMap) MarshalJSON() ([]byte, error) {
	res := append([]byte{}, '{')

	front, back := om.l.Front(), om.l.Back()
	for e := front; e != nil; e = e.Next() {
		k := e.Value.(string)
		res = append(res, fmt.Sprintf("%q:", k)...)

		var b []byte
		b, err := json.Marshal(om.m[k])
		if err != nil {
			return nil, microerror.Mask(err)
		}

		res = append(res, b...)
		if e != back {
			res = append(res, ',')
		}
	}

	return append(res, '}'), nil
}

func (om *OrderedMap) Set(key string, value interface{}) {
	if _, ok := om.m[key]; !ok {
		om.keys[key] = om.l.PushBack(key)
	}
	om.m[key] = value
}

func (om *OrderedMap) UnmarshalJSON(data []byte) error {
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.UseNumber()

	// must open with a delim token '{'
	t, err := dec.Token()
	if err != nil {
		return microerror.Mask(err)
	}
	if delim, ok := t.(json.Delim); !ok || delim != '{' {
		return microerror.Maskf(executionFailedError, "data must be JSON object")
	}

	err = om.parseobject(dec)
	if err != nil {
		return microerror.Mask(err)
	}

	_, err = dec.Token()
	if err != io.EOF {
		return microerror.Maskf(executionFailedError, "data must be JSON object")
	}

	return nil
}

func (om *OrderedMap) Values() []string {
	var values []string
	for _, k := range om.Keys() {
		values = append(values, om.m[k].(string))
	}

	return values
}

func (om *OrderedMap) parseobject(dec *json.Decoder) error {
	for dec.More() {
		t, err := dec.Token()
		if err != nil {
			return microerror.Mask(err)
		}

		key, ok := t.(string)
		if !ok {
			return microerror.Maskf(executionFailedError, "data must be JSON object")
		}

		t, err = dec.Token()
		if err == io.EOF {
			break
		} else if err != nil {
			return microerror.Mask(err)
		}

		var value interface{}
		value, err = handledelim(t, dec)
		if err != nil {
			return microerror.Mask(err)
		}

		// om.keys = append(om.keys, key)
		om.keys[key] = om.l.PushBack(key)
		om.m[key] = value
	}

	t, err := dec.Token()
	if err != nil {
		return microerror.Mask(err)
	}
	if delim, ok := t.(json.Delim); !ok || delim != '}' {
		return microerror.Maskf(executionFailedError, "data must be JSON object")
	}

	return nil
}

func handledelim(t json.Token, dec *json.Decoder) (interface{}, error) {
	if delim, ok := t.(json.Delim); ok {
		switch delim {
		case '{':
			om2 := NewOrderedMap()
			err := om2.parseobject(dec)
			if err != nil {
				return nil, microerror.Mask(err)
			}
			return om2, nil
		case '[':
			var value []interface{}
			value, err := parsearray(dec)
			if err != nil {
				return nil, microerror.Mask(err)
			}
			return value, nil
		default:
			return nil, microerror.Maskf(executionFailedError, "unexpected delimiter %#q", delim)
		}
	}

	return t, nil
}

func parsearray(dec *json.Decoder) ([]interface{}, error) {
	var t json.Token
	arr := make([]interface{}, 0)
	for dec.More() {
		t, err := dec.Token()
		if err != nil {
			return nil, microerror.Mask(err)
		}

		var value interface{}
		value, err = handledelim(t, dec)
		if err != nil {
			return nil, microerror.Mask(err)
		}
		arr = append(arr, value)
	}
	t, err := dec.Token()
	if err != nil {
		return nil, microerror.Mask(err)
	}
	if delim, ok := t.(json.Delim); !ok || delim != ']' {
		return nil, microerror.Maskf(executionFailedError, "data must be JSON array")
	}

	return arr, nil
}
