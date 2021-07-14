gooastream
====

Overview

This library was created with the aim of achieving the same thing in Golang with reference to [Akka Stream](https://github.com/akka/akka).


## Description
You can combine units called Source, Flow, and Sink to create a complex stream pipeline.

## Usage
<details>
<summary>See Sample</summary>

```go
package gooastream

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/require"
	"sync"
	"testing"
	"time"
)

/*
	リストの要素を10倍にして標準出力するストリーム
	Source o-> Flow -> Sink
*/
func Test_Simple_Runnable(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	var wg sync.WaitGroup

	n := 100
	list := make([]interface{}, n)
	wg.Add(n)
	for i := 0; i < n; i++ {
		list[i] = i
	}
	source := NewSource(list, 100)
	flow := NewFlow(func(i interface{}) (interface{}, error) {
		return i.(int) * 10, nil
	}, 0)
	sink := NewSink(func(i interface{}) error {
		fmt.Println(i)
		wg.Done()
		return nil
	}, 0)

	runnable := source.Via(flow).To(sink)
	_, _, done, runningCancel := runnable.Run(ctx)
	go func() {
		wg.Wait()
		runningCancel()
	}()

	// runningCancel()が呼ばれるまでブロック
	done()
	require.NoError(t, ctx.Err())
}

/*
	リストの要素を10倍にして標準出力するストリーム
	Chan o-> Source -> Flow -> Sink
*/
func Test_Simple_Channel_Runnable(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	var wg sync.WaitGroup

	n := 100
	wg.Add(n)
	sourceChannel, source := NewChannelSource(100)
	go func() {
		for i := 0; i < n; i++ {
			sourceChannel.Push(ctx, i)
		}
		// いつでもキャンセルできる
		// sourceChannel.Close()
	}()

	flow := NewFlow(func(i interface{}) (interface{}, error) {
		return i.(int) * 10, nil
	}, 0)
	sink := NewSink(func(i interface{}) error {
		fmt.Println(i)
		wg.Done()
		return nil
	}, 0)

	runnable := source.Via(flow).To(sink)
	_, _, done, runningCancel := runnable.Run(ctx)
	go func() {
		wg.Wait()
		runningCancel()
	}()

	// sourceChannel.Close()が呼ばれるかrunningCancel()が呼ばれるまでブロック
	done()
	require.NoError(t, ctx.Err())
}

/*
	balancerで複数のストリームで処理し、margeで一個のストリームにまとめる
	Chan ──> Source ─┬─ Flow ─┬─> Sink
                         ├─ Flow ─┤
                         └─ Flow ─┘
*/
func Test_Balance_Marge_Runnable(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	var wg sync.WaitGroup

	n := 100
	wg.Add(n)
	sourceChannel, source := NewChannelSource(100)
	go func() {
		for i := 0; i < n; i++ {
			sourceChannel.Push(ctx, i)
		}
		// いつでもキャンセルできる
		// sourceChannel.Close()
	}()

	// balanceとmargeのポート数は揃える必要がある(panicする)
	port := 3
	balanceFlow := NewBalanceFlow(port, 0)
	childFlowA := NewFlow(func(i interface{}) (interface{}, error) {
		time.Sleep(time.Millisecond * 10)
		return i.(int) * 2, nil
	}, 0)
	childFlowB := NewFlow(func(i interface{}) (interface{}, error) {
		time.Sleep(time.Millisecond * 20)
		return i.(int) * -2, nil
	}, 0)
	childFlowC := NewFlow(func(i interface{}) (interface{}, error) {
		time.Sleep(time.Millisecond * 30)
		return 0, nil
	}, 0)
	margeFlow := NewMargeFlow(port, 100)

	balanceMargeFlow := balanceFlow.
		IVia([]Flow{childFlowA, childFlowB, childFlowC}).
		Via(margeFlow)

	sink := NewSink(func(i interface{}) error {
		fmt.Println(i)
		wg.Done()
		return nil
	}, 100)

	runnable := source.Via(balanceMargeFlow).To(sink)
	_, _, done, runningCancel := runnable.Run(ctx)
	go func() {
		wg.Wait()
		runningCancel()
	}()

	// sourceChannel.Close()が呼ばれるかrunningCancel()が呼ばれるまでブロック
	done()
	require.NoError(t, ctx.Err())
}

/*
	基本的にどこでCancelやCloseを呼んでもRunnableはキャンセルされる
	1. sourceChannelをCloseした時
	2. 各Taskでエラーを返した時
*/
func Test_Error_Handle_Task(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	n := 100
	sourceChannel, source := NewChannelSource(100)
	go func() {
		for i := 0; i < n; i++ {
			err := sourceChannel.Push(ctx, i)
			require.NoError(t, err)
		}
		// いつでもキャンセルできる
		//sourceChannel.Close()
	}()

	flow := NewFlow(func(i interface{}) (interface{}, error) {
		return i, nil
	}, 0)
	sink := NewSink(func(i interface{}) error {
		require.LessOrEqual(t, i, 50, "50以上の時にエラーを返すことで、50を含むそれ以降がキャンセルされる")
		if i.(int) >= 50 {
			return fmt.Errorf("force")
		}
		return nil
	}, 0)

	runnable := source.Via(flow).To(sink)
	_, _, done, runningCancel := runnable.Run(ctx)
	go func() {
		_ = runningCancel
	}()

	// sourceChannel.Close()が呼ばれるかrunningCancel()が呼ばれるまでブロック
	done()
	require.NoError(t, ctx.Err())
}

func Test_Error_Handle_Channel(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	var wg sync.WaitGroup

	n := 100
	wg.Add(n)
	sourceChannel, source := NewChannelSource(100)
	go func() {
		for i := 0; i < n; i++ {
			err := sourceChannel.Push(ctx, i)
			require.NoError(t, err)
		}
		// いつでもキャンセルできる
		sourceChannel.Close()
	}()

	flow := NewFlow(func(i interface{}) (interface{}, error) {
		require.Fail(t, "sourceChannelを読んでいるのでここは呼ばれるべきではない")
		return i, nil
	}, 0)
	sink := NewSink(func(i interface{}) error {
		require.Fail(t, "sourceChannelを読んでいるのでここは呼ばれるべきではない")
		wg.Done()
		return nil
	}, 0)

	runnable := source.Via(flow).To(sink)
	_, _, done, runningCancel := runnable.Run(ctx)
	go func() {
		wg.Wait()
		runningCancel()
	}()

	// sourceChannel.Close()が呼ばれるかrunningCancel()が呼ばれるまでブロック
	done()
	require.NoError(t, ctx.Err())
}

```

</details>

## Install
```bash
$ go get github.com/BambooTuna/gooastream@v1.1.0
```

## Contribution

## Author
[BambooTuna](https://github.com/BambooTuna)