package featuremap

import (
	"bytes"
	"container/list"
	"encoding/json"
	"fmt"
	"io"

	"github.com/giantswarm/microerror"

	"github.com/giantswarm/gg/pkg/colour"
)

// KVPair for initializing from a list of key-value pairs, or for looping
// entries in the same order.
type KVPair struct {
	Key   string
	Value interface{}
}

// FeatureMap has similar operations as the default map, but maintains the order
// of inserted keys. Similar to map, all single key operations e.g. get, set and
// delete runs at O(1). Although the JSON spec says the keys order of an object
// should not matter, sometimes the order of JSON objects and their keys matters
// when printing them for humans. Therefore we have to maintain the object keys
// in the same order as they come in.
//
// Disclaimer, same as Go's default map, FeatureMap is not safe for concurrent
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
type FeatureMap struct {
	colr colour.Palette
	dict map[string]interface{}
	list *list.List
	keys map[string]*list.Element
}

func NewWithPalette(palette colour.Palette) *FeatureMap {
	return &FeatureMap{
		colr: palette,
		dict: make(map[string]interface{}),
		list: list.New(),
		keys: make(map[string]*list.Element),
	}
}

func (fm *FeatureMap) Delete(key string) (value interface{}, ok bool) {
	value, ok = fm.dict[key]
	if ok {
		fm.list.Remove(fm.keys[key])
		delete(fm.keys, key)
		delete(fm.dict, key)
	}
	return
}

func (fm *FeatureMap) EntriesIter() func() (*KVPair, bool) {
	e := fm.list.Front()
	return func() (*KVPair, bool) {
		if e != nil {
			key := e.Value.(string)
			e = e.Next()
			return &KVPair{key, fm.dict[key]}, true
		}
		return nil, false
	}
}

func (fm *FeatureMap) EntriesReverseIter() func() (*KVPair, bool) {
	e := fm.list.Back()
	return func() (*KVPair, bool) {
		if e != nil {
			key := e.Value.(string)
			e = e.Prev()
			return &KVPair{key, fm.dict[key]}, true
		}
		return nil, false
	}
}

func (fm *FeatureMap) Get(key string) interface{} {
	return fm.dict[key]
}

func (fm *FeatureMap) GetValue(key string) (value interface{}, ok bool) {
	value, ok = fm.dict[key]
	return
}

func (fm *FeatureMap) Has(key string) bool {
	_, ok := fm.dict[key]
	return ok
}

func (fm *FeatureMap) Keys() []string {
	var keys []string
	for e := fm.list.Front(); e != nil; e = e.Next() {
		keys = append(keys, e.Value.(string))
	}

	return keys
}

func (fm *FeatureMap) Len() int {
	return len(fm.dict)
}

func (fm *FeatureMap) MarshalJSON() ([]byte, error) {
	res := append([]byte{}, '{')

	front, back := fm.list.Front(), fm.list.Back()
	for e := front; e != nil; e = e.Next() {
		k := e.Value.(string)
		v := fm.dict[k]

		res = append(res, fm.colr.Key(fmt.Sprintf("%q:", k))...)

		var b []byte
		fmt.Printf("%#v\n", v)
		b, err := json.Marshal(v)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		_, ok := v.(string)
		if ok {
			res = append(res, fm.colr.Value(string(b))...)
		} else {
			res = append(res, b...)
		}

		if e != back {
			res = append(res, ',')
		}
	}

	return append(res, '}'), nil
}

func (fm *FeatureMap) Set(key string, value interface{}) {
	if _, ok := fm.dict[key]; !ok {
		fm.keys[key] = fm.list.PushBack(key)
	}
	fm.dict[key] = value
}

func (fm *FeatureMap) UnmarshalJSON(data []byte) error {
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

	err = fm.parseobject(dec)
	if err != nil {
		return microerror.Mask(err)
	}

	_, err = dec.Token()
	if err != io.EOF {
		return microerror.Maskf(executionFailedError, "data must be JSON object")
	}

	return nil
}

func (fm *FeatureMap) Values() []string {
	var values []string
	for _, k := range fm.Keys() {
		values = append(values, fm.dict[k].(string))
	}

	return values
}

func (fm *FeatureMap) parseobject(dec *json.Decoder) error {
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
		value, err = handledelim(fm.colr, t, dec)
		if err != nil {
			return microerror.Mask(err)
		}

		fm.keys[key] = fm.list.PushBack(key)
		fm.dict[key] = value
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

func handledelim(p colour.Palette, t json.Token, dec *json.Decoder) (interface{}, error) {
	if delim, ok := t.(json.Delim); ok {
		switch delim {
		case '{':
			om2 := NewWithPalette(p)
			err := om2.parseobject(dec)
			if err != nil {
				return nil, microerror.Mask(err)
			}
			return om2, nil
		case '[':
			var value []interface{}
			value, err := parsearray(p, dec)
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

func parsearray(p colour.Palette, dec *json.Decoder) ([]interface{}, error) {
	var t json.Token
	arr := make([]interface{}, 0)
	for dec.More() {
		t, err := dec.Token()
		if err != nil {
			return nil, microerror.Mask(err)
		}

		var value interface{}
		value, err = handledelim(p, t, dec)
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
