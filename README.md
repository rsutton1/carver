# Carver

Carver is an easy-to-use command-line tool that seamlessly organizes JSON files.

## Example

Say you have the following (contrived) configurations for a hypothetical app.

dev.json:
```
{
    "foo": "bar",
    "env": "dev",
    "feature_flags": {
        "featureA": true,
        "featureB": true,
        "new_feature": true
    },
    "domain": "dev.internal.example.com",
    "tls": false
}
```

staging.json:
```
{
    "foo": "bar",
    "env": "staging",
    "feature_flags": {
        "featureA": true,
        "featureB": true
    },
    "domain": "staging.internal.example.com",
    "tls": true
}
```

prod.json:
```
{
    "foo": "bar",
    "env": "prod",
    "feature_flags": {
        "featureA": true
    },
    "domain": "example.com",
    "tls": true
}
```

We can use Carver to consolidate these files into a common file with
environment-specific overrides.

```
$ carver normalize
Generated ./common.json
Updated ./dev.json
Updated ./staging.json
Updated ./prod.json
```

After running Carver, the files contain the following content:

common.json
```
{
    "foo": "bar",
    "feature_flags": {
        "featureA": true
    }
}
```

dev.json
```
{
    "env": "dev",
    "feature_flags": {
        "featureB": true,
        "new_feature": true
    },
    "domain": "dev.internal.example.com",
    "tls": false
}
```

staging.json:
```
{
    "env": "staging",
    "feature_flags": {
        "featureB": true
    },
    "domain": "staging.internal.example.com",
    "tls": true
}
```

prod.json:
```
{
    "env": "prod",
    "domain": "example.com",
    "tls": true
}
```

This demonstrates that Carver consolidates common key-value pairs by moving them
into the common file, `common.json`. For instance, Carver consolidated the "foo"
key because it has same value (of "bar") in all files. However, notice that
"tls" key was not consolidated, since the value differs in the "dev"
file. Whenever it runs, Carver ensures that the common file contains any
keys which have the same value in all files.

Carver is idempotent so it can be run repeatedly. If it finds the files are
already consolidated, it will not make any changes.

Run carver with `merge` to restore the files:

```
$ carver merge
Removed ./common.json
Updated ./dev.json
Updated ./staging.json
Updated ./prod.json
```
