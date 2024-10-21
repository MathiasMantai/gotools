# Package cli

This package contains functionality for interacting with the command line


# Functions

## PrintColor
Prints a string with a specified color to the command line

### Example

```go
import (
    "github.com/MathiasMantai/gotools/cli"
)

func main() {
    cli.PrintColor("This text should be green", "green", true)
}
```

## PrintBold
Prints a string as bold to the command line

### Example

```go
import (
    "github.com/MathiasMantai/gotools/cli"
)

func main() {
    cli.PrintBold("This text should be bold", true)
}
```


## PrintBoldAndColor
Prints a string with a specified color and bold to the command line

### Example

```go
import (
    "github.com/MathiasMantai/gotools/cli"
)

func main() {
    cli.PrintBoldAndColor("This text should be bold and blue", "blue", true)
}
```


## PrintWithTime
Prints the current time (based on the system time) in frontn of a string

### Example

```go

import (
    "github.com/MathiasMantai/gotools/cli"
)

func main() {
    cli.PrintWithTime("This is the current time", true)
}
```



## PrintWithTimeAndColor
Prints the current time (based on the system time) in front of a colored string

### Example

```go

import (
    "github.com/MathiasMantai/gotools/cli"
)

func main() {
    cli.PrintWithTimeAndColor("This should be colored blue", "blue", true)
}
```
