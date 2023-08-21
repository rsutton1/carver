package main

import (
    // "fmt"
    "testing"
    "encoding/json"
    "reflect"
)

type testfn func(interface{}, interface{}) interface{}

func CaseMaps(t * testing.T, fn testfn, m1Str string, m2Str string, expectedStr string) {
    var m1 map[string]interface{}
    var m2 map[string]interface{}
    var expected map[string]interface{}
    json.Unmarshal([]byte(m1Str), &m1)
    json.Unmarshal([]byte(m2Str), &m2)
    json.Unmarshal([]byte(expectedStr), &expected)

    actual := fn(m2, m1).(map[string]interface{})

    eq := reflect.DeepEqual(expected, actual)
    if ! eq {
        t.Fatalf(`expected %q, got %q`, expected, actual)
    }
}

func TestMergeEmpty(t * testing.T) {
    initStr := `{
    }`
    mergeStr := `{
    }`
    expectedStr := `{
    }`
    CaseMaps(t, merge, initStr, mergeStr, expectedStr)
}

func TestMergeEmptyInit(t * testing.T) {
    initStr := `{}`
    mergeStr := `{
        "biz": "baz"
    }`
    expectedStr := `{
        "biz": "baz"
    }`
    CaseMaps(t, merge, initStr, mergeStr, expectedStr)
}

func TestMergeEmptyMerge(t * testing.T) {
    initStr := `{
        "foo": "bar"
    }`
    mergeStr := `{
    }`
    expectedStr := `{
        "foo": "bar"
    }`
    CaseMaps(t, merge, initStr, mergeStr, expectedStr)
}

func TestMergeUnion(t * testing.T) {
    initStr := `{
        "foo": "bar"
    }`
    mergeStr := `{
        "biz": "baz"
    }`
    expectedStr := `{
        "foo": "bar",
        "biz": "baz"
    }`
    CaseMaps(t, merge, initStr, mergeStr, expectedStr)
}

func TestMergeIntersection(t * testing.T) {
    initStr := `{
        "initonly": "initonly_val",
        "shared": "shared_val"
    }`
    mergeStr := `{
        "mergeonly": "mergeonly_val",
        "shared": "shared_newval"
    }`
    expectedStr := `{
        "initonly": "initonly_val",
        "mergeonly": "mergeonly_val",
        "shared": "shared_newval"
    }`
    CaseMaps(t, merge, initStr, mergeStr, expectedStr)
}

func TestMergeReplace(t * testing.T) {
    initStr := `{
        "foo": "bar"
    }`
    mergeStr := `{
        "foo": "baz"
    }`
    expectedStr := `{
        "foo": "baz"
    }`
    CaseMaps(t, merge, initStr, mergeStr, expectedStr)
}

func TestMergeNestedUnion(t * testing.T) {
    initStr := `{
        "foo": {
            "bar": "barval"
        }
    }`
    mergeStr := `{
        "foo": {
            "biz": "bizval"
        }
    }`
    expectedStr := `{
        "foo": {
            "bar": "barval",
            "biz": "bizval"
        }
    }`
    CaseMaps(t, merge, initStr, mergeStr, expectedStr)
}

func TestMergeNestedReplace(t * testing.T) {
    initStr := `{
        "foo": {
            "bar": "val"
        }
    }`
    mergeStr := `{
        "foo": {
            "bar": "newval"
        }
    }`
    expectedStr := `{
        "foo": {
            "bar": "newval"
        }
    }`
    CaseMaps(t, merge, initStr, mergeStr, expectedStr)
}

func TestMergeNestedPreserveBase(t * testing.T) {
    initStr := `{
        "foo": "bar"
    }`
    mergeStr := `{
        "biz": {
            "baz": "val"
        }
    }`
    expectedStr := `{
        "foo": "bar",
        "biz": {
            "baz": "val"
        }
    }`
    CaseMaps(t, merge, initStr, mergeStr, expectedStr)
}

func TestRemove(t * testing.T) {
    bStr := `{
        "foo": "bar",
        "biz": "baz"
    }`
    cStr := `{
        "__": [
            {
                "remove": [
                    "foo"
                ]
            }
        ]
    }`
    mStr := `{
        "biz": "baz"
    }`

    CaseMaps(t, merge, bStr, cStr, mStr)
}

func TestRemoveMultiple(t * testing.T) {
    bStr := `{
        "foo": "fooval",
        "baz": "bazval",
        "biz": "bizval"
    }`
    cStr := `{
        "__": [
            {
                "remove": [
                    "foo",
                    "baz"
                ]
            }
        ]
    }`
    mStr := `{
        "biz": "bizval"
    }`

    CaseMaps(t, merge, bStr, cStr, mStr)
}

func TestRemoveNone(t * testing.T) {
    bStr := `{
        "foo": "bar"
    }`
    cStr := `{
        "__": [
            {
                "remove": [
                ]
            }
        ]
    }`
    mStr := `{
        "foo": "bar"
    }`

    CaseMaps(t, merge, bStr, cStr, mStr)
}

func TestRemoveNested(t * testing.T) {
    bStr := `{
        "foo": {
            "biz": "baz",
            "remove_me": "val"
        }
    }`
    cStr := `{
        "foo": {
            "__": [
                {
                    "remove": [
                        "remove_me"
                    ]
                }
            ]
        }
    }`
    mExpectedStr := `{
        "foo": {
            "biz": "baz"
        }
    }`

    CaseMaps(t, merge, bStr, cStr, mExpectedStr)
}

func TestRemoveBaseEmpty(t * testing.T) {
    bStr := `{}`
    cStr := `{
        "__": [
            {
                "remove": [
                    "foo"
                ]
            }
        ]
    }`
    mStr := `{}`
    CaseMaps(t, merge, bStr, cStr, mStr)
}

func TestRemoveBaseEmptyNested(t * testing.T) {
    bStr := `{
    }`
    cStr := `{
        "foo": {
            "__": [
                {
                    "remove": [
                        "bar"
                    ]
                }
            ]
        }
    }`
    mStr := `{
        "foo": {}
    }`
    CaseMaps(t, merge, bStr, cStr, mStr)
}

func TestRemoveBaseNestedUnion(t * testing.T) {
    bStr := `{
        "foo": {
            "biz": "bizval"
        }
    }`
    cStr := `{
        "foo": {
            "__": [
                {
                    "remove": [
                        "bar"
                    ]
                }
            ],
            "baz": "bazval"
        }
    }`
    mStr := `{
        "foo": {
            "biz": "bizval",
            "baz": "bazval"
        }
    }`
    CaseMaps(t, merge, bStr, cStr, mStr)
}

func TestReplace(t * testing.T) {
    bStr := `{
        "foo": "bar"
    }`
    cStr := `{
        "__": [
            {
                "replace": {}
            }
        ],
        "biz": "bizval"
}`
    mStr := `{
        "biz": "bizval"
    }`

    CaseMaps(t, merge, bStr, cStr, mStr)
}

func TestReplaceNone(t * testing.T) {
    bStr := `{
        "foo": "bar"
    }`
    cStr := `{
        "__": [
            {
                "replace": {}
            }
        ]
}`
    mStr := `{
    }`

    CaseMaps(t, merge, bStr, cStr, mStr)
}

func TestReplaceNested(t * testing.T) {
    bStr := `{
        "foo": {
            "bar": "barval"
        }
    }`
    cStr := `{
        "foo": {
            "__": [
                {
                    "replace": {}
                }
            ],
            "biz": "bizval"
        }
}`
    mStr := `{
        "foo": {
            "biz": "bizval"
        }
    }`

    CaseMaps(t, merge, bStr, cStr, mStr)
}

type testCase[T any] struct {
    arg1, arg2, expected T
}

type testCaseOneArg[T any, O any] struct {
    in T
    expected O
}

type testCaseTwoArgs[T1 any, T2 any, O any] struct {
    arg1 T1
    arg2 T2
    expected O
}

type testCaseMonad[T1 any, T2 any, O any] struct {
    value T1
    args []T2
    expected O
}

func runTestsParallel[T any](t * testing.T, fn func(T, T) T, testCases map[string]testCase[any]) {
    for name, test := range testCases {
        test := test
        name := name
        t.Run(
            name,
            func(t * testing.T) {
                t.Parallel()

                actual := fn(test.arg1.(T), test.arg2.(T))

                eq := reflect.DeepEqual(test.expected, actual)
                if ! eq {
                    t.Fatalf(`test %s: expected %v, got %v`, name, test.expected, actual)
                }
            },
        )
    }
}

func TestIntersect(t * testing.T) {
    testCases := map[string]testCase[interface{}]{
        "empty": {
            []string{},
            []string{},
            []string{},
        },
        "empty_arg1": {
            []string{"foo"},
            []string{},
            []string{},
        },
        "empty_arg2": {
            []string{},
            []string{"foo"},
            []string{},
        },
        "nomatch": {
            []string{"foo"},
            []string{"bar"},
            []string{},
        },
        "base": {
            []string{"foo"},
            []string{"foo"},
            []string{"foo"},
        },
        "base_reflective1": {
            []string{"bar","foo"},
            []string{"foo"},
            []string{"foo"},
        },
        "base_reflective2": {
            []string{"foo"},
            []string{"bar","foo"},
            []string{"foo"},
        },
        "eq_order": {
            []string{"foo","bar"},
            []string{"bar","foo"},
            []string{"bar","foo"},
        },
        "duplicate1": {
            []string{"foo"},
            []string{"foo","foo"},
            []string{"foo"},
        },
        "duplicate2": {
            []string{"foo","foo"},
            []string{"foo"},
            []string{"foo"},
        },
    }
    runTestsParallel[[]string](t, intersect, testCases)
}

func TestCommonParallel(t * testing.T) {
    testCases := map[string]testCase[any]{
        "empty": {
            map[string]interface{}{},
            map[string]interface{}{},
            map[string]interface{}{},

        },
        "common": {
            map[string]interface{}{
                "foo": "bar",
                "siz": "saz",
            },
            map[string]interface{}{
                "foo": "bar",
                "biz": "baz",
            },
            map[string]interface{}{
                "foo": "bar",
            },

        },
        "common_overwrite": {
            map[string]interface{}{"foo": "bar"},
            map[string]interface{}{"foo": "baz"},
            map[string]interface{}{"foo": "baz"},

        },
        "common_nested_overwrite": {
            map[string]interface{}{
                "foo": map[string]interface{}{"bar": "biz"},
            },
            map[string]interface{}{
                "foo": map[string]interface{}{"bar": "baz"},
            },
            map[string]interface{}{
                "foo": map[string]interface{}{"bar": "baz"},
            },
        },
        "common_nested_intersect": {
            map[string]interface{}{
                "foo": map[string]interface{}{
                    "bar": "biz",
                    "ziz": "zaz",
                },
            },
            map[string]interface{}{
                "foo": map[string]interface{}{
                    "bar": "baz",
                    "siz": "saz",
                },
            },
            map[string]interface{}{
                "foo": map[string]interface{}{"bar": "baz"},
            },
        },
    }
    runTestsParallel[any](t, common, testCases)
}

func runTestsOneArgParallel[T any, O any](t * testing.T, fn func(T) O, testCases map[string]testCaseOneArg[T, O]) {
    for name, test := range testCases {
        test := test
        name := name
        fn := fn
        t.Run(
            name,
            func(t * testing.T) {
                t.Parallel()

                actual := fn(test.in)

                eq := reflect.DeepEqual(test.expected, actual)
                if ! eq {
                    t.Fatalf(`test %s: expected %v, got %v`, name, test.expected, actual)
                }
            },
        )
    }
}

func runTestsTwoArgsParallel[T1 any, T2 any, O any](t * testing.T, fn func(T1, T2) O, testCases map[string]testCaseTwoArgs[T1, T2, O]) {
    for name, test := range testCases {
        test := test
        name := name
        t.Run(
            name,
            func(t * testing.T) {
                t.Parallel()

                actual := fn(test.arg1, test.arg2)

                eq := reflect.DeepEqual(test.expected, actual)
                if ! eq {
                    t.Fatalf(`test %s: expected %v, got %v`, name, test.expected, actual)
                }
            },
        )
    }
}

func TestTypeToString(t * testing.T) {
    testCases := map[string]testCaseOneArg[interface{}, string]{
        "empty": {
            map[string]string{
                "foo": "bar",
            },
            "string",
        },
        "base": {
            "foo",
            "string",
        },
        "bool": {
            true,
            "bool",
        },
    }
    runTestsOneArgParallel[any, string](t, type_to_string, testCases)
}

func TestKeyMapGetNode(t * testing.T) {
    testCases := map[string]testCaseTwoArgs[string, interface{}, keymap_node]{
        "empty": {
            "foo",
            "biz",
            keymap_node{
                Count: 1,
                Paths: map[string]interface{}{
                    "example.json": map[string]interface{}{},
                },
            },
        },
        "bool": {
            "foo",
            true,
            keymap_node{
                Count: 1,
                Paths: map[string]interface{}{
                    "example.json": map[string]interface{}{},
                },
            },
        },
        "list": {
            "foo",
            []int{1,2},
            keymap_node{
                Count: 1,
                Paths: map[string]interface{}{
                    "example.json": map[string]interface{}{},
                },
            },
        },
    }
    kmStr := []byte(`{
        "foo": {
            "string": {
                "\"biz\"": {
                    "count": 1,
                    "paths": {
                        "example.json": {}
                    }
                },
                "[1,2]": {
                    "count": 1,
                    "paths": {
                        "example.json": {}
                    }
                }
            },
            "bool": {
                "true": {
                    "count": 1,
                    "paths": {
                        "example.json": {}
                    }
                }
            }
        }
    }`)
    km := make(keymap)
    json.Unmarshal(kmStr, &km)
    runTestsTwoArgsParallel[string, interface{}, keymap_node](t, km.get_node, testCases)
}

func unmarshal(kmStr []byte) keymap {
    km := make(keymap)
    json.Unmarshal(kmStr, &km)
    return km
}

func runTestsMonad[T1 any, T2 any, O any](t * testing.T, fn func(T1, ...T2) (O, error), testCases map[string]testCaseMonad[T1, T2, O]) {
    for name, test := range testCases {
        test := test
        name := name
        fn := fn
        t.Run(
            name,
            func(t * testing.T) {
                t.Parallel()

                actual, _ := fn(test.value, test.args...)

                eq := reflect.DeepEqual(test.expected, actual)
                if ! eq {
                    expected_pretty, _ := json.MarshalIndent(test.expected, "", "  ")
                    actual_pretty, _ := json.MarshalIndent(actual, "", "  ")
                    t.Fatalf(`test %s: expected %v, got %v`, name, string(expected_pretty), string(actual_pretty))
                }
            },
        )
    }
}

func TestKeyMapBind(t * testing.T) {
    testCases := map[string]testCaseMonad[keymap, interface{}, keymap]{
        "empty": {
            unmarshal([]byte(`{}`)),
            []interface{}{
                file{
                    "exampleTest.json",
                    "exampleTest.json",
                    map[string]interface{}{},
                },
            },
            unmarshal([]byte(`{
                "": {
                    "string": {
                        "{}": {
                            "count": 1,
                            "paths": {
                                "exampleTest.json": {}
                            }
                        }
                    }
                }
            }`)),
        },
        "string": {
            unmarshal([]byte(`{}`)),
            []interface{}{
                file{
                    "exampleTest.json",
                    "exampleTest.json",
                    map[string]interface{}{
                        "foo": "biz",
                    },
                },
            },
            unmarshal([]byte(`{
                "foo": {
                    "string": {
                        "\"biz\"": {
                            "count": 1,
                            "paths": {
                                "exampleTest.json": {}
                            }
                        }
                    }
                }
            }`)),
        },
        "list": {
            unmarshal([]byte(`{}`)),
            []interface{}{
                file{
                    "exampleTest.json",
                    "exampleTest.json",
                    map[string]interface{}{
                        "foo": []int{1, 2},
                    },
                },
            },
            unmarshal([]byte(`{
                "foo": {
                    "string": {
                        "[1,2]": {
                            "count": 1,
                            "paths": {
                                "exampleTest.json": {}
                            }
                        }
                    }
                }
            }`)),
        },
        "nested_string": {
            unmarshal([]byte(`{}`)),
            []interface{}{
                file{
                    "exampleTest.json",
                    "exampleTest.json",
                    map[string]interface{}{
                        "biz": map[string]interface{}{
                            "baz": "bar",
                        },
                    },
                },
            },
            unmarshal([]byte(`{
                "biz.baz": {
                    "string": {
                        "\"bar\"": {
                            "count": 1,
                            "paths": {
                                "exampleTest.json": {}
                            }
                        }
                    }
                }
            }`)),
        },
  }
  f_bind := func(km keymap, args ...interface{})(keymap, error) {
      m1 := monad{nil, km, []string{}}.bind(km_merge, args...)
      return m1.km, m1.err
  }
  runTestsMonad[keymap, interface{}, keymap](t, f_bind, testCases)
}

func TestNormalizeBind(t * testing.T) {
    testCases := map[string]testCaseMonad[keymap, interface{}, keymap]{
        "string": {
            unmarshal([]byte(`{
                "foo": {
                    "string": {
                        "\"biz\"": {
                            "count": 1,
                            "paths": {
                                "exampleTest.json": {}
                            }
                        }
                    }
                }
            }`)),
            []interface{}{
                "common.json",
                1,
            },
            unmarshal([]byte(`{
                "foo": {
                    "string": {
                        "\"biz\"": {
                            "count": 1,
                            "paths": {
                                "common.json": {}
                            }
                        }
                    }
                }
            }`)),
        },
  }
  f_bind := func(km keymap, args ...interface{})(keymap, error) {
      m1 := monad{nil, km, []string{}}.bind(normalize, args...)
      return m1.km, m1.err
  }
  runTestsMonad[keymap, interface{}, keymap](t, f_bind, testCases)
}

func TestResolveBind(t * testing.T) {
    testCases := map[string]testCaseMonad[keymap, interface{}, keymap]{
        "string": {
            unmarshal([]byte(`{
                "foo": {
                    "string": {
                        "\"biz\"": {
                            "count": 2,
                            "paths": {
                                "common.json": {}
                            }
                        }
                    }
                }
            }`)),
            []interface{}{
                "common.json",
                []string{
                    "exampleTest.json",
                    "common.json",
                },
            },
            unmarshal([]byte(`{
                "foo": {
                    "string": {
                        "\"biz\"": {
                            "count": 1,
                            "paths": {
                                "exampleTest.json": {}
                            }
                        }
                    }
                }
            }`)),
        },
  }
  f_bind := func(km keymap, args ...interface{})(keymap, error) {
      m1 := monad{nil, km, []string{}}.bind(resolve, args...)
      m1.names = []string{
          "exampleTest.json",
          "common.json",
      }
      return m1.km, m1.err
  }
  runTestsMonad[keymap, interface{}, keymap](t, f_bind, testCases)
}
