# Testing with GoMock: A Tutorial

This is a quick tutorial on how to test code using the [GoMock](https://github.com/golang/mock) mocking library and the built-in testing package `testing`.

## Installation

First, we need to install the the `gomock` package `github.com/golang/mock/gomock` as well as the `mockgen` code generation tool `github.com/golang/mock/mockgen`. Technically, we *could* do without the code generation tool, but then we'd have to write our mocks by hand, which is tedious and error-prone.

Both packages can be installed using `go get`:

```bash
go get github.com/golang/mock/gomock
go get github.com/golang/mock/mockgen
```

We can verify that the `mockgen` binary was installed successfully by running

```bash
$GOPATH/bin/mockgen
```

This should output usage information and a flag list. Now that we're all set up, we're ready to test some code! 

## Basic Usage

Usage of *GoMock* follows four basic steps:

1. Use `mockgen` to generate a mock for the interface you wish to mock.
1. In your test, create an instance of `gomock.Controller` and pass it to your mock object's constructor to obtain a mock object.
1. Call `EXPECT()` on your mocks to set up their expectations and return values
1. Call `Finish()` on the mock controller to assert the mock's expectations

Let's look at a small example to demonstrate the above workflow. To keep things simple, we'll be looking at just two files -- an interface `Doer` in the file `doer/doer.go` that we wish to mock and a struct `User`  in `user/user.go` that uses the `Doer` interface.

The interface that we wish to mock is just a couple of lines -- it has a single method `DoSomething` that does something with an  `int` and a `string` and returns an `error`:

> `doer/doer.go`
> ```golang
> package doer
>
> type Doer interface {
>     DoSomething(int, string) error
> }
> ```

Here's the code that we want to test while mocking out the `Doer` interface:

> `user/user.go`
> ```golang
> package user
>
> import "github.com/sgreben/testing-with-gomock/doer"
>
> type User struct {
>     Doer doer.Doer
> }
>
> func (u *User) Use() error {
>     return u.Doer.DoSomething(123, "Hello GoMock")
> }
> ```

We'll put the mock for `Doer` in a package `mocks` in the root directory and the test for `User` in the file `user/user_test.go`. We start by creating a directory `mocks` that will contain our mock implementations and then running `mockgen` on the `doer` package:

```bash
mkdir -p mocks
```
```bash
mockgen -destination=mocks/mock_doer.go -package=mocks github.com/sgreben/testing-with-gomock/doer Doer 
```

We have to create the directory `mocks` ourselves because GoMock won't do it for us and will quit with an error instead. Here's what the arguments given to `mockgen` mean:

- `-destination=mocks/mock_doer.go`: put the generated mocks in the file `mocks/mock_doer.go`.
- `-package=mocks`: put the generated mocks in the package `mocks`
- `github.com/sgreben/testing-with-gomock/doer`: generate mocks for this package
- `Doer`: generate mocks for this interface. We can specify multiple interfaces here as a comma-separated list (e.g. `Doer1,Doer2`).

If `$GOPATH/bin` is not in our `$PATH`, we'll have to call `mockgen` via  `$GOPATH/bin/mockgen`. In the following we'll assume that we have run

```bash
export PATH=$GOPATH/bin:$PATH
```

either in the init script of the shell or as part of an interactive session.

Next, we define a *mock controller* inside our test, passing a `*testing.T` to its constructor, and use it to construct a mock of the `Doer` interface. We also `defer` its `Finish` method -- more on this later.

```golang
mockCtrl := gomock.NewController(t)
defer mockCtrl.Finish()

mockDoer := mocks.NewMockDoer(mockCtrl)
```

We can now call `EXPECT()` on the `mockDoer` to set up its expectations in our test. The call to `EXPECT()` returns an object (called a mock _recorder_) providing methods of the same names as the real object. 

Calling one of the methods on the mock recorder specifies an expected call with the given arguments. You can then chain other properties onto the call, such as:

- the return value (via `.Return(...)`)
- the number of times this call is expected to occur (via `.Times(number)`, or via `.MaxTimes(number)` and `.MinTimes(number)`)

In our case, we want to assert that `mockerDoer`'s `Do` method will be called _once_, with `123` and `"Hello GoMock"` as arguments, and will return `nil`:

```golang
mockDoer.EXPECT().DoSomething(123, "Hello GoMock").Return(nil).Times(1)
```

That's it - we've now specified our first mock call! Here's the complete example:

> `user/user_test.go`
> ```golang
> package user_test
>
> import (
>   "github.com/sgreben/testing-with-gomock/mocks"
>   "github.com/sgreben/testing-with-gomock/user"
> )
>
> func TestUse(t *testing.T) {
>     mockCtrl := gomock.NewController(t)
>     defer mockCtrl.Finish()
>
>     mockDoer := mocks.NewMockDoer(mockCtrl)
>     testUser := &user.User{Doer:mockDoer}
>
>     // Expect Do to be called once with 123 and "Hello GoMock" as parameters, and return nil from the mocked call.
>     mockDoer.EXPECT().DoSomething(123, "Hello GoMock").Return(nil).Times(1)
>
>     testUser.Use()
> }

The last thing to happen is the deferred `Finish()` -- it asserts the expectations for `mockDoer`. It's idiomatic to `defer` this call to `Finish` at the point of declaration of the mock controller -- this way we don't forget to assert the mock expectations later. 

Finally, we're ready to run our tests:

```bash
$ go test -v github.com/sgreben/testing-with-gomock/user
=== RUN   TestUse
--- PASS: TestUse (0.00s)
PASS
ok      github.com/sgreben/testing-with-gomock/user     0.007s
```

If you need to construct more than one mock, you can reuse the mock controller -- its `Finish` method will then assert the expectations of all mocks associated with the controller.

We might also want to assert that the value returned by the `Use` method is indeed the one returned to it by `DoSomething`. We can write another test, creating a dummy error and then specifying it as a return value for `mockDoer.DoSomething`:

> `user/user_test.go`
> ```golang
> func TestUseReturnsErrorFromDo(t *testing.T) {
>     mockCtrl := gomock.NewController(t)
>     defer mockCtrl.Finish()
>
>     dummyError := errors.New("dummy error")
>     mockDoer := mocks.NewMockDoer(mockCtrl)
>     testUser := &user.User{Doer:mockDoer}
>
>     // Expect Do to be called once with 123 and "Hello GoMock" as parameters, and return dummyError from the mocked call.
>     mockDoer.EXPECT().DoSomething(123, "Hello GoMock").Return(dummyError).Times(1)
>
>     err := testUser.Use()
>
>     if err != dummyError {
>         t.Fail()
>     }
> }

## Using *GoMock* with `go:generate`

For a large number of packages and interfaces to mock, running `mockgen` for each package and interface individually is cumbersome. To alleviate this problem, the `mockgen` command may be placed in a special [`go:generate` comment](https://blog.carlmjohnson.net/post/2016-11-27-how-to-use-go-generate/).

In our example, we can add a `go:generate` comment just below the `package` statement of our `doer.go`:

> `doer/doer.go`
> ```golang
> package doer
>
> //go:generate mockgen -destination=../mocks/mock_doer.go -package=mocks github.com/sgreben/testing-with-gomock/doer Doer
>
> type Doer interface {
>     DoSomething(int, string) error
> }
> ```

Note that at the point where `mockgen` is called, the current working directory is `doer` -- hence we need to specify `../mocks/` as the directory to write our mocks to, not just `mocks/`.

We can now comfortably generate all mocks specified by such a comment by running

```bash
go generate ./...
```

from the project's root directory. Note that there is no space between `//` and `go:generate` in the comment. This is required for `go generate` to pick up the comment as an instruction to process.

A reasonable policy on where to put the `go:generate` comment and which interfaces to include is the following:

- One `go:generate` comment per file containing interfaces to be mocked
- Include all interfaces to generate mocks for in the call to `mockgen`
- Put the mocks in a package `mocks` and write the mocks for a file `X.go` into `mocks/mock_X.go`.

This way, the `mockgen` call is close to the actual interfaces, while avoiding the overhead of separate calls and destination files for each interface.

## Using argument matchers

Sometimes, you don't care about the specific arguments a mock is called with. With *GoMock*, a parameter can be expected to have a fixed value (by specifying the value in the expected call) or it can be expected to match a predicate, called a *Matcher*. Matchers are used to represent ranges of expected arguments to a mocked method. The following matchers are pre-defined in *GoMock*:

- `gomock.Any()`: matches any value (of any type)
- `gomock.Eq(x)`: uses reflection to match values that are `DeepEqual` to `x`
- `gomock.Nil()`: matches `nil`
- `gomock.Not(m)`: (where `m` is a Matcher) matches values not matched by the matcher `m`
- `gomock.Not(x)`: (where `x` is *not* a Matcher) matches values not `DeepEqual` to `x`

For example, if we don't care about the value of the first argument to `Do`, we could write:

```golang
mockDoer.EXPECT().DoSomething(gomock.Any(), "Hello GoMock")
```

*GoMock* automatically converts arguments that are *not* of type `Matcher` to `Eq` matchers, so the above call is equivalent to:

```golang
mockDoer.EXPECT().DoSomething(gomock.Any(), gomock.Eq("Hello GoMock"))
```
 
You can define your own matchers by implementing the `gomock.Matcher` interface:

> `gomock/matchers.go` (excerpt)
> ```golang
> type Matcher interface {
>     Matches(x interface{}) bool
>     String() string
> }
> ```

The `Matches` method is where the actual matching happens, while `String` is used to generate human-readable output for failing tests. For example, a matcher checking an argument's type could be implemented as follows:

> `match/oftype.go`
> ```golang
> package match
> 
> import (
>     "reflect"
>     "github.com/golang/mock/gomock"
> )
> 
> type ofType struct{ t string }
> 
> func OfType(t string) gomock.Matcher {
>     return &ofType{t}
> }
> 
> func (o *ofType) Matches(x interface{}) bool {
>     return reflect.TypeOf(x).String() == o.t
> }
> 
> func (o *ofType) String() string {
>     return "is of type " + o.t
> }
> ```

We can then use our custom matcher like this:

```golang
// Expect Do to be called once with 123 and any string as parameters, and return nil from the mocked call.
mockDoer.EXPECT().
    DoSomething(123, match.OfType("string")).
    Return(nil).
    Times(1)
```

We've split the above call across multiple lines for readability. For more complex mock calls this is a handy way of making the mock specification more readable. Note that in Go we have to put the dot at the _end_ of each line in a sequence of chained calls. Otherwise, the parser will consider the line ended and we'll get a syntax error.


## Asserting call order

The order of calls to an object is often important. *GoMock* provides a way to assert that one call must happen after another call, the `.After` method. For example,

```golang
callFirst := mockDoer.EXPECT().DoSomething(1, "first this")
callA := mockDoer.EXPECT().DoSomething(2, "then this").After(callFirst)
callB := mockDoer.EXPECT().DoSomething(2, "or this").After(callFirst)
```

specifies that `callFirst` must occur before either `callA` or `callB`.

*GoMock* also provides a convenience function `gomock.InOrder` to specify that the calls must be performed in the exact order given. This is less flexible than using `.After` directly, but can make your tests more readable for longer sequences of calls:

```golang
gomock.InOrder(
    mockDoer.EXPECT().DoSomething(1, "first this"),
    mockDoer.EXPECT().DoSomething(2, "then this"),
    mockDoer.EXPECT().DoSomething(3, "then this"),
    mockDoer.EXPECT().DoSomething(4, "finally this"),
)
```

[Under](https://github.com/golang/mock/blob/master/gomock/call.go#L256-) the hood, `InOrder` uses `.After` to chain the calls in sequence.

## Specifying mock actions

Mock objects differ from real implementations in that they don't implement any of their behavior -- all they do is provide canned responses at the appropriate moment and record their calls. However, sometimes you need your mocks to do more than that. Here, *GoMock*'s `Do` actions come in handy. Any call may be decorated with an action by calling `.Do` on the call with a function to be executed whenever the call is matched:

```golang
mockDoer.EXPECT().
    DoSomething(gomock.Any(), gomock.Any()).
    Return(nil).
    Do(func(x int, y string) {
        fmt.Println("Called with x =",x,"and y =", y)
    })
```

Complex assertions about the call arguments can be written inside `Do` actions. For example, if the first (`int`) argument of `DoSomething` should be less than or equal to the length of the second (`string`) argument, we can write:

```golang
mockDoer.EXPECT().
    DoSomething(gomock.Any(), gomock.Any()).
    Return(nil).
    Do(func(x int, y string) {
        if x > len(y) {
            t.Fail()
        }
    })
```

The same functionality could _not_ be implemented using custom matchers, since we are _relating_ the concrete values, whereas matchers only have access to one argument at a time.

## Summary

In this post, we've seen how to generate mocks using `mockgen` and how to batch mock generation using `go:generate` comments and the `go generate` tool. We've covered the expectation API, including argument matchers, call frequency, call order and `Do`-actions.

If you have any questions or if you feel that there's something missing or unclear, please don't hesitate to let me know in the comments!