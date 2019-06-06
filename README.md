binding [![Build Status](https://drone.gitea.com/api/badges/tango/binding/status.svg)](https://drone.gitea.com/tango/binding)
 [![](http://gocover.io/_badge/gitea.com/tango/binding)](http://gocover.io/gitea.com/tango/binding)
=======

Middlware binding provides request data binding and validation for [Tango](https://gitea.com/lunny/tango).

## Installation

	go get gitea.com/tango/binding

## Example

```Go
import (
    "gitea.com/lunny/tango"
    "gitea.com/tango/binding"
)

type Action struct {
    binding.Binder
}

type MyStruct struct {
    Id int64
    Name string
}

func (a *Action) Get() string {
    var mystruct MyStruct
    errs := a.Bind(&mystruct)
    return fmt.Sprintf("%v, %v", mystruct, errs)
}

func main() {
    t := tango.Classic()
    t.Use(binding.Bind())
    t.Get("/", new(Action))
    t.Run()
}
```

Visit `/?id=1&name=2` on your browser and you will find output
```
{1 sss}, []
```

## Getting Help

- [API Reference](https://godoc.org/gitea.com/tango/binding)

## Credits

This package is forked from [macaron-contrib/binding](https://github.com/macaron-contrib/binding) with modifications.

## License

This project is under Apache v2 License. See the [LICENSE](LICENSE) file for the full license text.