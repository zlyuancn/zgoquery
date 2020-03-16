/*
-------------------------------------------------
   Author :       Zhang Fan
   date：         2020/2/28
   Description :  一个并发查询函数
-------------------------------------------------
*/

package zgoquery

import (
    "context"
    "errors"
    "sync"

    "github.com/zlyuancn/zerrors"
)

var ErrNoQueryFns = errors.New("没有传入查询函数")

type GoQueryFn func(ctx context.Context) (a interface{}, err error)

func GoQuery(ctx context.Context, fns ...GoQueryFn) (interface{}, error) {
    if len(fns) == 0 {
        return nil, ErrNoQueryFns
    }

    ctx, cancel := context.WithCancel(ctx)
    done := make(chan struct{})

    out := make(chan interface{}, len(fns))
    errc := make(chan error, len(fns)+1)

    var wg sync.WaitGroup
    wg.Add(len(fns))

    for _, fn := range fns {
        go func(fn GoQueryFn) {
            a, err := fn(ctx)
            if err != nil {
                errc <- err
            } else {
                out <- a
            }
            wg.Done()
        }(fn)
    }

    go func() {
        wg.Wait()
        close(done)
    }()

    select {
    case a := <-out:
        cancel()
        return a, nil
    case <-ctx.Done():
        errc <- ctx.Err()
    case <-done:
        cancel()
    }

    select {
    case a := <-out:
        return a, nil
    default:
    }

    errs := zerrors.NewErrors()
    for {
        select {
        case e := <-errc:
            errs.AddErrs(e)
        default:
            return nil, errs.Err()
        }
    }
}
