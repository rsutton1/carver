package main

import (
    "log"
    "fmt"
    "path"
    "sort"
    "os"
    "flag"
    "encoding/json"
    "golang.org/x/exp/maps"
    "github.com/nqd/flat"
)

type filesArgs []string

type keymap_node struct {
    Count int `json:"count"`
    Paths map[string]interface{} `json:"paths"`
}

// type keymap_value struct {
//     values map[string]keymap_node
// }

// type keymap_type struct {
//     types map[string]keymap_value
// }

type keymap map[string]map[string]map[string]keymap_node

type monad struct {
    err error
    km keymap
    names []string
}

type file struct {
    name string
    path string
    obj map[string]interface{}
}

func (i * filesArgs) String() string {
    return "hello"
}

func (i * filesArgs) Set(value string) error {
    *i = append(*i, value)
    return nil
}

var bf filesArgs
var cf filesArgs

func readJsonFile(path string) (interface{}, error) {
    b, err := os.ReadFile(path)
    if err != nil {
        return nil, err
    }
    var f interface{}
    json.Unmarshal(b, &f)
    return f, err
}

func merge(o1 interface{}, o2 interface{}) interface{} {
    o1m := o1.(map[string]interface{})
    o2m := o2.(map[string]interface{})
    next := o2m
    ops, opsExist := o1m["__"]
    if opsExist {
        delete(o1m, "__")
    }
    for k, v := range o1m {
        switch vv := v.(type) {
        case map[string]interface{}:
            o2v, ok := o2m[k]
            if !ok {
                o2v = make(map[string]interface{})
            }
            next[k] = merge(v, o2v)
        default:
            next[k] = vv
        }
    }
    if opsExist {
        for _, op := range ops.([]interface{}) {
            opRemove, ok := op.(map[string]interface{})["remove"]
            if ok {
                for _, keyRemove := range opRemove.([]interface{}) {
                    delete(next, keyRemove.(string))
                }
            }
            _, ok = op.(map[string]interface{})["replace"]
            if ok {
                next = o1m
            }
        }
    }
    return next
}

func intersect(s1 []string, s2 []string) []string {
    var smap map[string]bool
    smap = make(map[string]bool)
    var common map[string]bool
    common = make(map[string]bool)
    for _, k := range s1 {
        smap[k] = true
    }
    for _, k := range s2 {
        _ , ok := smap[k]
        if ok {
            common[k] = true
        }
    }
    commonList := maps.Keys(common)
    sort.Strings(commonList)
    return commonList
}

func isMap(v interface{}) bool {
    switch v.(type) {
    case map[string]interface{}:
        return true
    }
    return false
}

func common(o1 interface{}, o2 interface{}) interface{} {
    o1m := o1.(map[string]interface{})
    o2m := o2.(map[string]interface{})
    merged := make(map[string]interface{})
    keys1 := maps.Keys(o1m)
    keys2 := maps.Keys(o2m)
    common_keys := intersect(keys1, keys2)
    for _, k := range common_keys {
        v1, _ := o1m[k]
        v2, _ := o2m[k]
        v1IsMap := isMap(v1)
        v2IsMap := isMap(v2)
        if v1IsMap && v2IsMap {
            merged[k] = common(v1, v2)
        } else {
            merged[k] = o2m[k]
        }
    }
    return merged
}

// func main() {
//     flag.Var(&bf, "bf", "path to base file")
//     flag.Var(&cf, "cf", "path to child file")
//     flag.Parse()
//     b := readJsonFile(bf[0])
//     c := readJsonFile(cf[0])
//     mf := merge(c, b)
//     m, _ := json.Marshal(mf)
//     fmt.Println(string(m))
// }

func loadFiles(dir string, ignore string) []file {
    files, err := os.ReadDir(dir)
    if err != nil {
        log.Fatal(err)
        os.Exit(1)
    }
    var file_objs []file
    for _, e := range files {
        file_path := path.Clean(e.Name())
        file_path_absolute := path.Clean(dir + "/" + e.Name())
        fmt.Println(file_path)
        fmt.Println(ignore)
        if file_path == ignore {
            continue
        }
        c, err := readJsonFile(file_path_absolute)
        if err != nil {
            log.Fatal(err)
            os.Exit(1)
        }
        file_obj := file{
            file_path,
            file_path,
            c.(map[string]interface{}),
        }
        file_objs = append(file_objs, file_obj)
    }
    return file_objs
}

func keymapFiles(files []file) monad {
    m1 := monad{
        nil,
        make(keymap),
        []string{},
    }
    for _, f := range files {
        args := []interface{}{f}
        m1 = m1.bind(km_merge, args...)
        m1.names = append(m1.names, f.name)
    }
    return m1
}

func writeFiles(output_dir string, filenames map[string]map[string]interface{}) {
    for name, obj := range filenames {
        file_path_absolute := path.Clean(output_dir + "/" + name)
        objI, _ := flat.Unflatten(obj, nil)
        objStr, _ := json.MarshalIndent(objI, "", "  ")
        err := os.MkdirAll(output_dir, 0750)
        if err != nil {
            log.Fatal(err)
        }
        err = os.WriteFile(file_path_absolute, objStr, 0666)
        if err != nil {
            log.Fatal(err)
        }
    }
}

func main() {
    var c string
    var n string
    normalizeCmd := flag.NewFlagSet("normalize", flag.ExitOnError)
    normalizeCmd.StringVar(&c, "c", "./", "config directory")
    normalizeCmd.StringVar(&n, "n", "./.carver/", "normalized directory")
    mergeCmd := flag.NewFlagSet("merge", flag.ExitOnError)
    mergeCmd.StringVar(&c, "c", "./", "config directory")
    mergeCmd.StringVar(&n, "n", "./.carver/", "normalized directory")
    if len(os.Args) < 2 {
        fmt.Println("expected subcommand 'normalize' or 'merge'")
        os.Exit(1)
    }
    sub_args := os.Args[2:]
    common_path := path.Clean("common.json")
    c = path.Clean(c)
    n = path.Clean(n)
    switch os.Args[1] {
    case "normalize":
        normalizeCmd.Parse(sub_args)
        dir := c
        files := loadFiles(dir, n)
        fmt.Println(files)
        keymap_monad := keymapFiles(files)
        num_files := len(keymap_monad.names)
        filenames := keymap_monad.
            bind(
                normalize,
                []interface{}{
                    common_path,
                    num_files,
                }...).
            km.to_files()
        writeFiles(n, filenames)
        // m, _ := json.MarshalIndent(filenames, "", "  ")
    case "merge":
        mergeCmd.Parse(sub_args)
        dir := n
        common_path := path.Clean("common.json")
        files := loadFiles(dir, "")
        keymap_monad := keymapFiles(files)
        normalizeCmd.Parse(sub_args)
        names := keymap_monad.names
        filenames := keymap_monad.
            bind(
                resolve,
                []interface{}{
                    common_path,
                    names,
                }...).
            km.to_files()
        writeFiles(c, filenames)
        // m, _ := json.MarshalIndent(filenames, "", "  ")
    default:
        fmt.Println("invalid subcommand")
    }
}

func type_to_string(v interface{}) string {
    switch v.(type) {
    case bool:
        return "bool"
    default:
        return "string"
    }
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

// func (km keymap) copy() keymap {
//     km_copy := make(keymap)
//     for k := range km {
//         km_copy[k] = make(map[string]map[string]keymap_node)
//         for i := range km[k] {
//             km_copy[k][i] = make(map[string]keymap_node)
//             for p, q := range km[k][i] {
//                 new_paths := make(map[string]interface{})
//                 for x, z := range q.Paths {
//                     new_paths[x] = z
//                 }
//                 km_copy[k][i][p] = keymap_node{q.Count, new_paths}
//             }
//         }
//     }
//     // fmt.Printf("\nkm      %v\n", km)
//     // fmt.Printf("km_copy %v\n", km_copy)
//     return km_copy
// }

func (m monad) bind(f func(keymap, ...interface{}) (keymap, error), args ...interface{}) monad {
    if m.err != nil {
        return monad{m.err, m.km, m.names}
    }
    result, err := f(m.km, args...)
    str, _ := json.MarshalIndent(result,"","  ")
    fmt.Println(f)
    fmt.Println(args)
    fmt.Println(string(str))
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
    f := fs[0].(file)
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
    num_files := len(names) - 1
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
                fmt.Println(kmn.Paths)
                fmt.Println(p)
                fmt.Println(err)
                if err {
                    fmt.Println("IN")
                    kmn.Paths = paths_new
                    kmn.Count = num_files
                } else {
                    fmt.Println("OUT")
                }
                km[p][t][vStr] = kmn
            }
        }
    }
    return km, nil
}
