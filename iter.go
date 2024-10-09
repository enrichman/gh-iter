package ghiter

import (
	"context"
	"errors"
	"fmt"
	"iter"
	"net/url"
	"reflect"
	"strconv"
	"strings"

	"github.com/google/go-github/v66/github"
)

type Iterator[T, O any] struct {
	opt O

	ctx  context.Context
	args []string

	fn  func(ctx context.Context, opt O) ([]T, *github.Response, error)
	fn1 func(ctx context.Context, arg1 string, opt O) ([]T, *github.Response, error)
	fn2 func(ctx context.Context, arg1, arg2 string, opt O) ([]T, *github.Response, error)

	raw *github.Response
	err error
}

func NewFromFn[T, O any](
	fn func(ctx context.Context, opt O) ([]T, *github.Response, error),
) *Iterator[T, O] {
	return &Iterator[T, O]{
		fn: fn,
	}
}

func NewFromFn1[T, O any](
	fn1 func(ctx context.Context, arg1 string, opt O) ([]T, *github.Response, error),
	arg1 string,
) *Iterator[T, O] {
	return &Iterator[T, O]{
		fn1:  fn1,
		args: []string{arg1},
	}
}

func NewFromFn2[T, O any](
	fn2 func(ctx context.Context, arg1, arg2 string, opt O) ([]T, *github.Response, error),
	arg1 string,
	arg2 string,
) *Iterator[T, O] {
	return &Iterator[T, O]{
		fn2:  fn2,
		args: []string{arg1, arg2},
	}
}

func (it *Iterator[T, O]) Ctx(ctx context.Context) *Iterator[T, O] {
	it.ctx = ctx
	return it
}

func (it *Iterator[T, O]) Args(args ...string) *Iterator[T, O] {
	it.args = args
	return it
}

func (it *Iterator[T, O]) Opts(opt O) *Iterator[T, O] {
	it.opt = opt
	return it
}

func (it *Iterator[T, O]) Raw() *github.Response {
	return it.raw
}

func (it *Iterator[T, O]) Err() error {
	return it.err
}

func (it *Iterator[T, O]) All() iter.Seq[T] {
	initialize(it)

	return func(yield func(T) bool) {
		if err := validate(it); err != nil {
			it.err = err
			return
		}

		for {
			parts, resp, err := it.do()
			it.raw = resp

			if err != nil {
				it.err = err
				return
			}

			// push the result until yield, or an error occurs
			for _, p := range parts {
				if !yield(p) || contextErr(it) {
					return
				}
			}

			// no more results, break
			if resp.NextPage == 0 {
				return
			}

			// get the next page from the link header
			links := ParseLinkHeader(resp.Header.Get("link"))
			if next, found := links.FindByRel("next"); found {
				nextURL, err := url.Parse(next.URL)
				if err != nil {
					it.err = err
					return
				}

				vals := make(map[string]string)
				for k, v := range nextURL.Query() {
					vals[k] = v[0]
				}

				updateOptions(it.opt, vals)
			}
		}
	}
}

func initialize[T, O any](it *Iterator[T, O]) {
	// initialize context if nil
	if it.ctx == nil {
		it.ctx = context.Background()
	}

	// initialize options if nil
	if reflect.ValueOf(it.opt).IsNil() {
		optionPointerType := reflect.TypeOf(it.opt)
		optionValue := reflect.New(optionPointerType.Elem())
		if opt, ok := optionValue.Interface().(O); ok {
			it.opt = opt
		}
	}
}

func validate[T, O any](it *Iterator[T, O]) error {
	if it.fn == nil && it.fn1 == nil && it.fn2 == nil {
		return errors.New("no func provided")
	}

	numOfArgs := len(it.args)
	if it.fn1 != nil {
		if numOfArgs != 1 {
			args := strings.Join(it.args, ",")
			return fmt.Errorf("wrong number of arguments: expected 1, got %d [%s]", numOfArgs, args)
		}

		if it.args[0] == "" {
			return errors.New("empty argument[0]")
		}
	}

	if it.fn2 != nil {
		if numOfArgs != 2 {
			args := strings.Join(it.args, ",")
			return fmt.Errorf("wrong number of arguments: expected 2, got %d [%s]", numOfArgs, args)
		}

		if it.args[0] == "" {
			return errors.New("empty argument[0]")
		}

		if it.args[1] == "" {
			return errors.New("empty argument[1]")
		}
	}

	return nil
}

func contextErr[T, O any](it *Iterator[T, O]) bool {
	if err := it.ctx.Err(); err != nil {
		it.err = err
		return true
	}
	return false
}

func (it *Iterator[T, O]) do() ([]T, *github.Response, error) {
	if it.fn != nil {
		return it.fn(it.ctx, it.opt)
	} else if it.fn1 != nil {
		return it.fn1(it.ctx, it.args[0], it.opt)
	} else if it.fn2 != nil {
		return it.fn2(it.ctx, it.args[0], it.args[1], it.opt)
	}

	return nil, nil, errors.New("no func provided")
}

// updateOptions will update the github options based on the provided map and the `url` tag.
// If the field in the struct has a `url` tag it tries to set the value of the field from the one
// found in the map, if any.
func updateOptions(v any, m map[string]string) {
	valueOf := reflect.ValueOf(v)
	typeOf := reflect.TypeOf(v)

	if valueOf.Kind() == reflect.Pointer {
		valueOf = valueOf.Elem()
		typeOf = typeOf.Elem()
	}

	for i := 0; i < valueOf.NumField(); i++ {
		structField := typeOf.Field(i)
		fieldValue := valueOf.Field(i)

		// if field is of type struct then iterate over the pointer
		if structField.Type.Kind() == reflect.Struct {
			if fieldValue.CanAddr() {
				updateOptions(fieldValue.Addr().Interface(), m)
			}
		}

		// otherwise check if it has a 'url' tag
		urlTag := structField.Tag.Get("url")
		if urlTag != "" {
			urlParam := strings.Split(urlTag, ",")[0]

			if fieldValue.IsValid() && fieldValue.CanSet() {
				if v, found := m[urlParam]; found {
					switch fieldValue.Kind() {
					case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
						if i, err := strconv.Atoi(v); err == nil {
							fieldValue.SetInt(int64(i))
						}
					case reflect.Ptr:
						fieldValue.Set(reflect.ValueOf(&v))
					default:
						fieldValue.Set(reflect.ValueOf(v))
					}
				}
			}
		}
	}
}
