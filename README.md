# Pull-Push

POC to work sanely with `io.Reader` / `io.Writer`.

From

```go
n, err := io.Copy(
  outerWriter1(
    outerWriter2(
      outerWriter3(
        sink,
      ),
    ),
  ),
  outerReader3(
    outerReader2(
      outerReader1(
        src,
      ),
    ),
  ),
)
```

to

```go
  pipe := PP{}
  pipe.
    Pull(src).
    Pull(outerReader1()).
    Pull(outerReader2()).
    Pull(outerReader3()).
    Push(outerWriter1()).
    Push(outerWriter2()).
    Push(outerWriter3()).
    Push(sink)

  res := pipe.Copy(make([]byte, 1024))
  log.Printf("read %v / wrote %v / err %v\n",res.Read(), res.Wrote(), res.Error())
```

# Problem

None of the std api is made to work with it, so its current usage is kind of limited.
