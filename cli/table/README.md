# Package cli/table

This package contains methods to print tables to the command line

# Table::Print

Prints a table to the command Line

## Example

```go
import (
    "github.com/MathiasMantai/gotools/cli/table"
)

func main() {
    var table table.Table

    tableError := table.Print([][]string{{"Rowid", "Name", "Price"}, {"1", "Product1", "200"}, {"2", "Product2", "500"}}, 2)
    
    if tableError != nil {
        fmt.Println(tableError)
        return
    }
}
```

Output:
```
----------------------------------
|  Rowid  |      Name  |  Price  |
----------------------------------
|      1  |  Product1  |    200  |
----------------------------------
|      2  |  Product2  |    500  |
----------------------------------
```


# Table::PrintWithHeader

Prints a table to the command Line

## Example

```go
import (
    "github.com/MathiasMantai/gotools/cli/table"
)

func main() {
    var table table.Table

    tableError := table.PrintWithHeader([][]string{{"Rowid", "Name", "Price"}, {"1", "Product1", "200"}, {"2", "Product2", "500"}}, "red", 2)
    
    if tableError != nil {
        fmt.Println(tableError)
        return
    }
}
```

Output (Note that header will be bold and colored in the specified color):
```
----------------------------------
|  Rowid  |      Name  |  Price  |
----------------------------------
|      1  |  Product1  |    200  |
----------------------------------
|      2  |  Product2  |    500  |
----------------------------------
```