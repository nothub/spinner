# 🎡 Spinner

## Usage

Wrap a func:

```go
spinner.Spin("Fetching...", func () { work() })
```

Or do whatever:

```go
sp := spinner.Start("Fetching...")
defer sp.Stop()
...
```

## Options

There are some:

```go
sp := spinner.Start("Downloading",
    spinner.WithFrames([]string{"🌑", "🌘", "🌗", "🌖", "🌕", "🌔", "🌓", "🌒"}),
    spinner.WithStartDelay(1*time.Second),   // default 3s; 0 animates instantly
    spinner.WithDelay(250*time.Millisecond), // frame interval, default 250ms
    spinner.WithMaxWidth(80),                // line cap in terminal cells, default 80
    spinner.WithLabelFunc(func() string {    // regenerates the label every frame
        return fmt.Sprintf("Downloading %d/%d", done.Load(), total)
    }))
defer sp.Stop()
```

The label func runs on the spinner's goroutine, so whatever it reads has to be safe for concurrent use.
