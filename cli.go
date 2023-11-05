package main

import (
    "log"
    "fmt"
    "path"
    "os"
    "flag"
    "encoding/json"
    "github.com/nqd/flat"
    "github.com/ghodss/yaml"
)

type filesArgs []string

type file struct {
    name string
    path string
    obj map[string]interface{}
}

type vfile struct {
    name string
    root_path string
    path string
    obj map[string]interface{}
}

type opts struct {
    Dirs []string `json:"dirs"`
}

type dir struct {
    name string
    path string
}

func (d dir) get_name() string {
    return d.name
}

func (d dir) list_files(root_path string, exclude_dirs bool) []string {
    file_paths := []string{}
    files, err := os.ReadDir(root_path + "/" + d.path)
    if err != nil {
        log.Fatal(err)
        os.Exit(1)
    }
    for _, e := range files {
        if exclude_dirs && e.IsDir() {
            continue
        }
        name := path.Clean(e.Name())
        file_paths = append(file_paths, name)
    }
    return file_paths
}
type group struct {
    path string
    dirs []dir
}

type file_map struct {
    name string
    paths map[string][]string
}

func (fm file_map) add_file(name string, path string) {
    v, ok := fm.paths[name]
    if !ok {
        fm.paths[name] = []string{}
    }
    fm.paths[name] = append(v, path)
}

func (fm file_map) add_dir(d dir) {
    files := d.list_files(fm.name, true)
    for _, f_name := range files {
        file_path := d.get_name() + "/" + f_name
        fm.add_file(f_name, file_path)
    }
}


func (g group) get_dirs() []dir {
    return g.dirs
}

func (g group) get_file_map(include_root_files bool) file_map {
    fm := file_map{g.path,map[string][]string{}}
    for _, d := range g.get_dirs() {
        fm.add_dir(d)
    }
    if include_root_files {
        fm.add_dir(dir{".","."})
    }
    return fm
}

func (fm file_map) load_path(name string) []vfile {
    paths, ok := fm.paths[name]
    if ! ok {
        paths = []string{}
    }
    fs, _ := new_files(fm.name, paths)
    return fs
}

func (fm file_map) get_keymap_group(name string) keymap_group {
    fs := fm.load_path(name)
    return keymap_group{name, new_keymap(fs)}
}

func (fm file_map) get_keys() []string {
    keys := []string{}
    for key := range fm.paths {
        keys = append(keys, key)
    }
    return keys
}

func (fm file_map) get_keymap_groups() []keymap_group {
    kmgs := []keymap_group{}
    for _, key := range fm.get_keys() {
        kmg := fm.get_keymap_group(key)
        kmgs = append(kmgs, kmg)
    }
    return kmgs
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
    if json.Valid(b) {
        err = json.Unmarshal(b, &f)
    } else {
        err = yaml.Unmarshal(b, &f)
    }
    return f, err
}

func new_file(root_dir string, path string) (* vfile, error) {
    b, err := os.ReadFile(root_dir + "/" + path)
    if err != nil {
        log.Fatal(err)
        os.Exit(1)
    }
    f := vfile{path, root_dir, path, map[string]interface{}{}}
    if json.Valid(b) {
        err = json.Unmarshal(b, &f.obj)
    } else {
        err = yaml.Unmarshal(b, &f.obj)
    }
    return &f, err
}

func new_opts(path string) (opts, error) {
    b, err := os.ReadFile(path)
    if err != nil {
        return opts{}, err
    }
    var c opts
    if json.Valid(b) {
        err = json.Unmarshal(b, &c)
    } else {
        err = yaml.Unmarshal(b, &c)
    }
    return c, err
}

func new_group(config_path string, root_dir string) * group {
    config, err := new_opts(config_path+"/.carver.yaml")
    if err != nil {
        log.Fatal(err)
        os.Exit(1)
    }
    var config_paths []dir
    for _, dstr := range config.Dirs {
        dir_path := dstr
        dir_obj := dir{dir_path,path.Clean(dir_path)}
        config_paths = append(config_paths, dir_obj)
    }
    return &group{root_dir,config_paths}
}

func new_files(root_dir string, file_paths []string) ([]vfile, error) {
    var fs []vfile
    for _, file_path := range file_paths {
        file_path_absolute := path.Clean(file_path)
        f, err := new_file(root_dir, file_path_absolute)
        if err != nil {
            return []vfile{}, err
        }
        fs = append(fs, *f)
    }
    return fs, nil
}

func new_keymap(files []vfile) monad {
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
        file_ext := path.Ext(file_path_absolute)
        objI, _ := flat.Unflatten(obj, nil)
        var objStr []byte
        if file_ext == ".json" {
            objStr, _ = json.MarshalIndent(objI, "", "  ")
        } else {
            objStr, _ = yaml.Marshal(objI)
        }
        err := os.MkdirAll(path.Dir(file_path_absolute), 0750)
        if err != nil {
            log.Fatal(err)
        }
        err = os.WriteFile(file_path_absolute, objStr, 0666)
        if err != nil {
            log.Fatal(err)
        }
    }
}

func printUsage() {
    fmt.Println(`usage: carver [options] command

  command:
    normalize          normalize CONFIG_DIR and store the result in NORMALIZED_DIR
    merge              merge NORMALIZED_DIR and store the result in CONFIG_DIR
    help               print this message

  options:
    -c CONFIG_DIR      configuration directory
    -n NORMALIZED_DIR  normalized directory
    `)
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
        printUsage()
        os.Exit(1)
    }
    sub_args := os.Args[2:]
    c = path.Clean(c)
    n = path.Clean(n)
    // a group is a root directory and the env directories
    // e.g. group{"project/", ["envA/","envB/"]}

    // a file map is a kv map from file_names -> []file_paths
    // e.g. service1.json -> [envA/service1.json, envB/service1.json]

    // the keymap type stores the mapping from object keys -> val types -> vals
    // -> files containing the vals
    // this allows us to index by key to get key info with fast performance.
    // a keymap stores the information of a file across all environments. you
    // can reconstruct a file in all environments with only the keymap file.
    // the only thing that's missing is the unpathed filename (e.g.
    // "service1.json"). you could just truncate one of the paths (e.g.
    // filename("envA/service1.json")) to get the name, but that's not as
    // elegant.
    // e.g. "service_name" -> "string" -> "foo" -> "envA/service1.json"

    // the monad type stores a keymap with a bind function

    // the keymap_group type stores a monad with an id. the id provides the
    // unpathed filename, which we can use to create the common file when
    // merging json.
    switch os.Args[1] {
    case "normalize":
        normalizeCmd.Parse(sub_args)
        fm := new_group(c, c).get_file_map(false)
        kmgs := fm.get_keymap_groups()
        for _, kmg := range kmgs {
            filenames := kmg.km.
                bind(
                    normalize,
                    []interface{}{
                        kmg.id,
                        len(kmgs),
                    }...).
                km.to_files()
            writeFiles(n, filenames)
        }
    case "merge":
        mergeCmd.Parse(sub_args)
        fm := new_group(c, n).get_file_map(true)
        kmgs := fm.get_keymap_groups()
        for _, kmg := range kmgs {
            filenames := kmg.km.
                bind(
                    resolve,
                    []interface{}{
                        kmg.id,
                        kmg.km.get_names(),
                    }...).
                km.to_files()
            writeFiles(c, filenames)
            // fmt.Println(filenames)
        }
    default:
        printUsage()
    }
}

