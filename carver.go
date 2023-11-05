package main

import (
    "encoding/json"
    "github.com/nqd/flat"
)


type keymap_node struct {
    Count int `json:"count"`
    Paths map[string]interface{} `json:"paths"`
}

type keymap map[string]map[string]map[string]keymap_node

type monad struct {
    err error
    km keymap
    names []string
}

type keymap_group struct {
    id string
    km monad
}

func type_to_string(v interface{}) string {
    switch v.(type) {
    case bool:
        return "bool"
    default:
        return "string"
    }
}

func (m monad) get_names() []string {
    return m.names
}

func (km keymap) get_node(p string, v interface{}) keymap_node {
    value_type := type_to_string(v)
    value_bytes, _ := json.Marshal(v)
    value_str := string(value_bytes)
    kmn, ok := km[p][value_type][value_str]
    if !ok {
        return keymap_node{Count: 0, Paths: map[string]interface{}{}}
    }
    return kmn
}

func (km keymap) set_node(p string, v interface{}, kmn keymap_node) keymap {
    value_type := type_to_string(v)
    value_bytes, _ := json.Marshal(v)
    value_str := string(value_bytes)
    _, ok := km[p]
    if !ok {
        km[p] = map[string]map[string]keymap_node{}
    }
    _, ok = km[p][value_type]
    if !ok {
        km[p][value_type] = map[string]keymap_node{}
    }
    km[p][value_type][value_str] = kmn
    return km
}

func (m monad) bind(f func(keymap, ...interface{}) (keymap, error), args ...interface{}) monad {
    if m.err != nil {
        return monad{m.err, m.km, m.names}
    }
    result, err := f(m.km, args...)
    return monad{err, result, m.names}
}

func (km keymap) to_files() map[string]map[string]interface{} {
    filenames := map[string]map[string]interface{}{}
    for p := range km {
        for t := range km[p] {
            for vStr := range km[p][t] {
                kmn := km[p][t][vStr]
                var v interface{}
                json.Unmarshal([]byte(vStr), &v)
                for n := range kmn.Paths {
                    _, exists := filenames[n]
                    if !exists {
                        filenames[n] = map[string]interface{}{}
                    }
                    switch vv := v.(type) {
                    default:
                        filenames[n][p] = vv
                    }
                }
            }
        }
    }
    return filenames
}

func km_merge(km_updated keymap, fs ...interface{}) (keymap, error) {
    f := fs[0].(vfile)
    km_flat, _ := flat.Flatten(f.obj, nil)
    for path, value := range km_flat {
        kmn := km_updated.get_node(path, value)
        kmn.Count++
        kmn.Paths[f.path] = map[string]interface{}{}
        km_updated = km_updated.set_node(path, value, kmn)
    }
    return km_updated, nil
}

func normalize(km keymap, args ...interface{}) (keymap, error) {
    common_name := args[0].(string)
    num_files := args[1].(int)
    for p := range km {
        for t := range km[p] {
            for vStr := range km[p][t] {
                kmn := km[p][t][vStr]
                var v interface{}
                json.Unmarshal([]byte(vStr), &v)
                if kmn.Count == num_files {
                    kmn.Paths = map[string]interface{}{
                        common_name: map[string]interface{}{},
                    }
                }
                km[p][t][vStr] = kmn
            }
        }
    }
    return km, nil
}

func resolve(km keymap, args ...interface{}) (keymap, error) {
    common_name := args[0].(string)
    names := args[1].([]string)
    num_files := len(names)
    paths_new := map[string]interface{}{}
    for _, name := range names {
        if name == common_name {
            continue
        }
        paths_new[name] = map[string]interface{}{}
    }
    for p := range km {
        for t := range km[p] {
            for vStr := range km[p][t] {
                kmn := km[p][t][vStr]
                _, err := kmn.Paths[common_name]
                if err {
                    kmn.Paths = paths_new
                    kmn.Count = num_files
                } else {
                }
                km[p][t][vStr] = kmn
            }
        }
    }
    return km, nil
}
