# Generating api.go

File was generated using [dbus-codegen-go](https://github.com/amenzhinsky/dbus-codegen-go)
from the interface of [org.freedesktop.resolve1](https://www.freedesktop.org/software/systemd/man/org.freedesktop.resolve1.html).
To get the XML representation of the interface, use the following:

```sh
gdbus introspect \
    --xml --system \
    --dest org.freedesktop.resolve1 \
    --object-path /org/freedesktop/resolve1
```

To regenerate the file from this XML, use the following:

```sh
dbus-codegen-go \
    -package=resolve1 \
    -client-only \
    -camelize \
    -prefix=org.freedesktop.resolve1 \
    -only=org.freedesktop.resolve1.Manager
```

Unfortunately some post-processing changes are required after generation or it
won't compile. Post-generation, make the following changes:

1. Remove `"fmt"` and `"errors"` from the generated file ([issue](https://github.com/amenzhinsky/dbus-codegen-go/issues/11))
1. Change `flags` to `in_flags` in the inputs of 4 methods with errors ([issue](https://github.com/amenzhinsky/dbus-codegen-go/issues/10))
