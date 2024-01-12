package core

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"reflect"
	"strconv"
	"strings"

	"golang.org/x/net/idna"
)

type NodeValType int

const (
	ValTypeString NodeValType = iota
	ValTypeInt
	ValTypeFloat
	ValTypeBool
	ValTypeArray
	ValTypeMap
)

func (n NodeValType) String() string {
	return [...]string{"string", "int", "float", "bool", "array", "map"}[n]
}

type NodeValue struct {
	Type  NodeValType
	value interface{}
}

// String returns the string representation of the value.
// If the value is not a string and valid type, it will be converted to a string.
// Otherwise, it will return an empty string.
func (n NodeValue) String() string {
	if n.Type == ValTypeString {
		return n.value.(string)
	}

	if n.Type == ValTypeInt {
		return strconv.Itoa(n.value.(int))
	}

	if n.Type == ValTypeFloat {
		return strconv.FormatFloat(n.value.(float64), 'f', -1, 64)
	}

	if n.Type == ValTypeBool {
		return strconv.FormatBool(n.value.(bool))
	}

	return ""
}

// Int returns the int representation of the value.
// If the value is not an int and valid type, it will be converted to an int.
// Otherwise, it will return 0.
func (n NodeValue) Int() int {
	if n.Type == ValTypeInt {
		return n.value.(int)
	}

	if n.Type == ValTypeString {
		i, err := strconv.Atoi(n.value.(string))
		if err != nil {
			return 0
		}
		return i
	}

	if n.Type == ValTypeFloat {
		return int(n.value.(float64))
	}

	if n.Type == ValTypeBool {
		if n.value.(bool) {
			return 1
		}
		return 0
	}

	return 0
}

// Float returns the float representation of the value.
// If the value is not a float and valid type, it will be converted to a float.
// Otherwise, it will return 0.
func (n NodeValue) Float() float64 {
	if n.Type == ValTypeFloat {
		return n.value.(float64)
	}

	if n.Type == ValTypeInt {
		return float64(n.value.(int))
	}

	if n.Type == ValTypeString {
		f, err := strconv.ParseFloat(n.value.(string), 64)
		if err != nil {
			return 0
		}
		return f
	}

	if n.Type == ValTypeBool {
		if n.value.(bool) {
			return 1
		}
		return 0
	}

	return 0
}

// Bool returns the bool representation of the value.
// If the value is not a bool and valid type, it will be converted to a bool.
// Otherwise, it will return false.
func (n NodeValue) Bool() bool {
	if n.Type == ValTypeBool {
		return n.value.(bool)
	}

	if n.Type == ValTypeInt {
		return n.value.(int) != 0
	}

	if n.Type == ValTypeFloat {
		return n.value.(float64) != 0
	}

	if n.Type == ValTypeString {
		b, err := strconv.ParseBool(n.value.(string))
		if err != nil {
			return false
		}
		return b
	}

	return false
}

type Node struct {
	Key      string
	Value    *NodeValue
	Parent   *Node
	Children *[]Node
}

type ResorceRecord struct {
	Name  string
	Class string
	Type  string
	TTL   int
	Data  string
}

func (r *ResorceRecord) String() string {
	return fmt.Sprintf("%s\t%s\t%d\t%s\t%s", r.Name, r.Class, r.TTL, r.Type, r.Data)
}

func CreateFromJSON(bytes []byte) (*Node, error) {
	var mapData map[string]interface{}
	if err := json.Unmarshal(bytes, &mapData); err != nil {
		return nil, err
	}

	tree, err := mapToNode("root", nil, mapData)
	if err != nil {
		return nil, err
	}

	tree.Key = ""
	return tree, nil
}

func mapToNode(key string, parent *Node, m map[string]interface{}) (*Node, error) {
	tree := &Node{
		Key: key,
		Value: &NodeValue{
			Type:  ValTypeMap,
			value: nil,
		},
		Children: &[]Node{},
		Parent:   parent,
	}

	for k, v := range m {
		switch reflect.TypeOf(v).Kind() {
		case reflect.Map:
			if child, err := mapToNode(k, tree, v.(map[string]interface{})); err == nil {
				tree.AddChild(child)
			} else {
				return nil, err
			}

		case reflect.Slice:
			if child, err := sliceToNode(k, tree, v.([]interface{})); err == nil {
				tree.AddChild(child)
			} else {
				return nil, err
			}

		default:
			if child, err := primitiveToNode(k, tree, v); err == nil {
				tree.AddChild(child)
			} else {
				return nil, err
			}
		}
	}

	return tree, nil
}

func sliceToNode(key string, parent *Node, s []interface{}) (*Node, error) {
	tree := &Node{
		Key: key,
		Value: &NodeValue{
			Type:  ValTypeArray,
			value: nil,
		},
		Parent:   parent,
		Children: &[]Node{},
	}

	for i, v := range s {
		switch reflect.TypeOf(v).Kind() {
		case reflect.Map:
			if child, err := mapToNode(strconv.Itoa(i), tree, v.(map[string]interface{})); err == nil {
				tree.AddChild(child)
			} else {
				return nil, err
			}

		case reflect.Slice:
			if child, err := sliceToNode(strconv.Itoa(i), tree, v.([]interface{})); err == nil {
				tree.AddChild(child)
			} else {
				return nil, err
			}

		default:
			if child, err := primitiveToNode(strconv.Itoa(i), tree, v); err == nil {
				tree.AddChild(child)
			} else {
				return nil, err
			}
		}
	}

	return tree, nil
}

func primitiveToNode(key string, parent *Node, v interface{}) (*Node, error) {
	node := &Node{
		Key:      key,
		Value:    nil,
		Parent:   parent,
		Children: nil,
	}

	switch reflect.TypeOf(v).Kind() {
	case reflect.Int:
		node.Value = &NodeValue{
			Type:  ValTypeInt,
			value: v.(int),
		}
		return node, nil

	case reflect.Float64:
		node.Value = &NodeValue{
			Type:  ValTypeFloat,
			value: v.(float64),
		}
		return node, nil

	case reflect.Bool:
		node.Value = &NodeValue{
			Type:  ValTypeBool,
			value: v.(bool),
		}
		return node, nil

	case reflect.String:
		node.Value = &NodeValue{
			Type:  ValTypeString,
			value: v.(string),
		}
		return node, nil
	}

	return nil, fmt.Errorf("invalid primitive type; value = %v, type = %v", v, reflect.TypeOf(v))
}

func (n *Node) AddChild(node *Node) {
	*n.Children = append(*n.Children, *node)
}

func (n *Node) GetChild(key string) *Node {
	for _, child := range *n.Children {
		if child.Key == key {
			return &child
		}
	}

	return nil
}

func (n *Node) GetChildValue(key string) *NodeValue {
	child := n.GetChild(key)
	if child == nil {
		return nil
	}

	return child.Value
}

// RRs returns the resource records of the node.
// The base argument is the base domain name of the node, and must be ended with a dot(root).
// For example, if the base is "yum.onl.", the resource record of the node "foo" will be "foo.yum.onl.".
//
// If the node is a map or an array, this returns the array of RRs, including the RRs of the children and the one as the pointer.
//
// Data field are constructed as follows:
//
//	`v=did:dinsec; t=<type>; d=<data>`
//	- type: the type of the value, "p" for premitives, "m" for map pointer, "a" for array pointer.
//	- data: the base64 encoded data of the value for premitives, the stringified number of the children for array, or the comma-separated string of the base64-encoded key for map.
func (n *Node) RRs(base string) []*ResorceRecord {
	rrs := []*ResorceRecord{}
	key := base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString([]byte(n.Key))

	// check if the node is root
	if n.Parent == nil && n.Value.Type == ValTypeMap {
		keys := []string{}
		recName := fmt.Sprintf("_did.%s", base)

		for _, child := range *n.Children {
			rrs = append(rrs, child.RRs(recName)...)
			keys = append(keys, base64.URLEncoding.
				WithPadding(base64.NoPadding).EncodeToString([]byte(child.Key)))
		}

		rrs = append(rrs, &ResorceRecord{
			Name:  recName,
			Class: "IN",
			Type:  "TXT",
			TTL:   3600,
			Data:  fmt.Sprintf("\"v=did:dnssec; t=m; d=%s\"", strings.Join(keys, ",")),
		})

		return rrs
	}

	recName := fmt.Sprintf("%s.%s", key, base)
	if n.Parent.Value.Type == ValTypeArray {
		recName = fmt.Sprintf("%s.%s", n.Key, base)
	}

	switch n.Value.Type {
	case ValTypeMap:
		keys := []string{}
		for _, child := range *n.Children {
			rrs = append(rrs, child.RRs(recName)...)
			keys = append(keys, base64.URLEncoding.
				WithPadding(base64.NoPadding).EncodeToString([]byte(child.Key)))
		}

		rrs = append(rrs, &ResorceRecord{
			Name:  recName,
			Class: "IN",
			Type:  "TXT",
			TTL:   3600,
			Data:  fmt.Sprintf("\"v=did:dnssec; t=m; d=%s\"", strings.Join(keys, ",")),
		})

	case ValTypeArray:
		for _, child := range *n.Children {
			rrs = append(rrs, child.RRs(recName)...)
		}

		rrs = append(rrs, &ResorceRecord{
			Name:  recName,
			Class: "IN",
			Type:  "TXT",
			TTL:   3600,
			Data:  fmt.Sprintf("\"v=did:dnssec; t=a; d=%d\"", len(*n.Children)),
		})

	case ValTypeString:
		rrs = append(rrs, &ResorceRecord{
			Name:  recName,
			Class: "IN",
			Type:  "TXT",
			TTL:   3600,
			Data: fmt.Sprintf(
				"\"v=did:dnssec; t=p; d=%s=%s\"",
				n.Value.Type.String(),
				base64.URLEncoding.WithPadding(base64.NoPadding).
					EncodeToString([]byte(n.Value.String())),
			),
		})

	default:
		rrs = append(rrs, &ResorceRecord{
			Name:  recName,
			Class: "IN",
			Type:  "TXT",
			TTL:   3600,
			Data: fmt.Sprintf(
				"\"v=did:dnssec; t=p; d=%s=%s\"",
				n.Value.Type.String(), n.Value.String(),
			),
		})
	}

	return rrs
}

func (n *Node) DumpRRs(f io.Writer, base string) error {
	for _, rr := range n.RRs(base) {
		if _, err := f.Write([]byte(rr.String() + "\n")); err != nil {
			return err
		}
	}
	return nil
}

func (n *Node) Print() error {
	return recursivePrintTree(n, 0)
}

func (n *Node) JSON() ([]byte, error) {
	if n.Value.Type == ValTypeMap {
		m, err := nodeToMap(n)
		if err != nil {
			return nil, err
		}

		return json.MarshalIndent(m, "", "  ")
	}

	if n.Value.Type == ValTypeArray {
		s, err := nodeToSlice(n)
		if err != nil {
			return nil, err
		}

		return json.MarshalIndent(s, "", "  ")
	}

	return nil, fmt.Errorf("invalid node type; typ = %s", n.Value.Type.String())
}

func Resolve(did string) (*Node, error) {
	if err := validateDidSyntax(did); err != nil {
		return nil, err
	}

	fqdn := strings.Split(did, ":")[2]
	node, err := resolve("_did."+fqdn+".", nil)
	if err != nil {
		return nil, err
	}

	fmt.Println("resolving done")
	return node, nil
}

func resolve(fqdn string, parent *Node) (*Node, error) {
	fmt.Printf("resolving %s...\n", fqdn)
	txt, err := net.LookupTXT(fqdn)
	if err != nil {
		return nil, err
	}

	var rType rValType
	var values []string

	for _, v := range txt {
		if typ, vals, err := parseRecordValue(v); err != nil || typ == rValTypeInvalid {
			continue
		} else {
			rType = typ
			values = vals
			break
		}
	}

	if len(values) == 0 {
		return nil, fmt.Errorf("no valid record found")
	}

	var node *Node
	if parent == nil {
		node = &Node{
			Key: "",
			Value: &NodeValue{
				Type:  ValTypeMap,
				value: nil,
			},
			Parent:   parent,
			Children: &[]Node{},
		}
	} else {
		node = &Node{
			Key:      strings.Split(fqdn, ".")[0],
			Parent:   parent,
			Children: &[]Node{},
		}
	}

	switch rType {
	case rValTypeMapPointer:
		for _, v := range values {
			node.Value = &NodeValue{
				Type:  ValTypeMap,
				value: nil,
			}

			next := fmt.Sprintf("%s.%s", v, fqdn)
			if child, err := resolve(next, node); err != nil {
				return nil, err
			} else {
				node.AddChild(child)
			}
		}
	case rValTypeArrayPointer:
		if len(values) != 1 {
			return nil, fmt.Errorf("got multiple values for array type; values = %v", values)
		}

		count, err := strconv.Atoi(values[0])
		if err != nil {
			return nil, err
		}

		node.Value = &NodeValue{
			Type: ValTypeArray,
		}

		for i := 0; i < count; i++ {
			next := fmt.Sprintf("%d.%s", i, fqdn)
			if child, err := resolve(next, node); err != nil {
				return nil, err
			} else {
				node.AddChild(child)
			}
		}

	case rValTypePremitive:
		if len(values) != 1 {
			return nil, fmt.Errorf("got multiple values for premitive type; values = %v", values)
		}

		typAndVal := strings.Split(values[0], "=")
		if len(typAndVal) != 2 {
			return nil, fmt.Errorf("invalid premitive value; got = %s", values[0])
		}

		switch typAndVal[0] {
		case "string":
			str, err := base64.URLEncoding.WithPadding(base64.NoPadding).DecodeString(typAndVal[1])
			if err != nil {
				return nil, err
			}

			node.Value = &NodeValue{
				Type:  ValTypeString,
				value: string(str),
			}

		case "int":
			i, err := strconv.Atoi(typAndVal[1])
			if err != nil {
				return nil, err
			}

			node.Value = &NodeValue{
				Type:  ValTypeInt,
				value: i,
			}

		case "float":
			f, err := strconv.ParseFloat(typAndVal[1], 64)
			if err != nil {
				return nil, err
			}

			node.Value = &NodeValue{
				Type:  ValTypeFloat,
				value: f,
			}

		case "bool":
			b, err := strconv.ParseBool(typAndVal[1])
			if err != nil {
				return nil, err
			}

			node.Value = &NodeValue{
				Type:  ValTypeBool,
				value: b,
			}

		default:
			return nil, fmt.Errorf("invalid premitive type; got = %s", typAndVal[0])
		}
	}

	return node, nil
}

func validateDidSyntax(did string) error {
	ary := strings.Split(did, ":")
	if len(ary) != 3 {
		return fmt.Errorf("invalid did syntax; did = %s", did)
	}

	if ary[0] != "did" {
		return fmt.Errorf("invalid URI scheme; expected = did, actual = %s", ary[0])
	}

	if ary[1] != "dnssec" {
		return fmt.Errorf("invalid DID method; expected = dnssec, actual = %s", ary[1])
	}

	// check if the ary[2] is a valid FQDN
	if _, err := idna.Lookup.ToASCII(ary[2]); err != nil {
		return fmt.Errorf("invalid domain name; got = %s", ary[2])
	}

	return nil
}

type rValType int

const (
	rValTypeInvalid    rValType = -1
	rValTypeMapPointer rValType = iota
	rValTypeArrayPointer
	rValTypePremitive
)

func parseRecordValue(value string) (rValType, []string, error) {
	// remove quotes and spaces
	value = strings.Trim(value, "\"")
	value = strings.Trim(value, "'")
	value = strings.Trim(value, " ")
	value = strings.ReplaceAll(value, " ", "")
	value = strings.ReplaceAll(value, "\t", "")

	ary := strings.Split(value, ";")
	mapping := map[string]string{}

	for _, v := range ary {
		kv := strings.Split(v, "=")
		if len(kv) != 2 && len(kv) != 3 {
			return rValTypeInvalid, nil, fmt.Errorf("invalid record value; got = %s", value)
		}

		if len(kv) == 2 {
			mapping[kv[0]] = kv[1]
		} else {
			mapping[kv[0]] = fmt.Sprintf("%s=%s", kv[1], kv[2])
		}
	}

	if mapping["v"] != "did:dnssec" {
		return rValTypeInvalid, nil, fmt.Errorf("invalid record version; got = %s", mapping["v"])
	}

	if mapping["d"] == "" {
		return rValTypeInvalid, nil, fmt.Errorf("invalid record data; got = %s", mapping["d"])
	}

	switch mapping["t"] {
	case "m":
		return rValTypeMapPointer, strings.Split(mapping["d"], ","), nil
	case "a":
		return rValTypeArrayPointer, strings.Split(mapping["d"], ","), nil
	case "p":
		return rValTypePremitive, []string{mapping["d"]}, nil
	default:
		return rValTypeInvalid, nil, fmt.Errorf("invalid value type; got = %s, expected = p || a || m", mapping["t"])
	}
}

func nodeToMap(node *Node) (map[string]interface{}, error) {
	m := map[string]interface{}{}

	for _, child := range *node.Children {
		bytes, err := base64.URLEncoding.WithPadding(base64.NoPadding).DecodeString(child.Key)
		if err != nil {
			return nil, err
		}
		key := string(bytes)

		switch child.Value.Type {
		case ValTypeMap:
			if childMap, err := nodeToMap(&child); err == nil {
				m[key] = childMap
			} else {
				return nil, err
			}

		case ValTypeArray:
			if childSlice, err := nodeToSlice(&child); err == nil {
				m[key] = childSlice
			} else {
				return nil, err
			}

		default:
			m[key] = child.Value.value
		}
	}

	return m, nil
}

func nodeToSlice(node *Node) ([]interface{}, error) {
	s := []interface{}{}

	for _, child := range *node.Children {
		switch child.Value.Type {
		case ValTypeMap:
			if childMap, err := nodeToMap(&child); err == nil {
				s = append(s, childMap)
			} else {
				return nil, err
			}

		case ValTypeArray:
			if childSlice, err := nodeToSlice(&child); err == nil {
				s = append(s, childSlice)
			} else {
				return nil, err
			}

		default:
			s = append(s, child.Value.value)
		}
	}

	return s, nil
}

func recursivePrintTree(tree *Node, depth int) error {
	indent := getIndent(depth)

	key := tree.Key
	if tree.Parent != nil && tree.Parent.Value.Type != ValTypeArray {
		keyb, err := base64.URLEncoding.WithPadding(base64.NoPadding).DecodeString(tree.Key)
		if err != nil {
			return err
		}
		key = string(keyb)
	}

	if tree.Value.Type != ValTypeMap && tree.Value.Type != ValTypeArray {
		fmt.Printf("%s%s: %s (%s)\n", indent, key, tree.Value, tree.Value.Type.String())
	} else {
		fmt.Printf("%s%s: (%s)\n", indent, key, tree.Value.Type.String())
		for _, child := range *tree.Children {
			recursivePrintTree(&child, depth+1)
		}
	}

	return nil
}

func getIndent(depth int) string {
	indent := ""
	for i := 0; i < depth; i++ {
		indent += "  "
	}

	return indent
}
