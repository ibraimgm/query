# query

A bare bones, no-magic query builder for Go.

## The problem

Imagine that you have to build a dynamic `SELECT` query, that will return data from the `employee` table and *might* be filtered by the department, employee age and employee name. You have to handle the case where all filter parameters are informed, as well as none or a combination of them. This isn't particularly *hard*, but leads to a boring chain of if statements like that:

```go
func MyFunc(db *sql.DB, department *string, name *string, age *int) (*sql.Rows, error) {
  sql := "SELECT id,name,age,dept FROM employees WHERE 1=1"
  var argc int
  var params []interface{}

  if department != nil {
    argc++
    sql += fmt.Sprintf(" AND dept = $%d", argc)
    params = append(params, department)
  }

  if name != nil {
    argc++
    sql += fmt.Sprintf(" AND name = $%d", argc)
    params = append(params, name)
  }

  if age != nil {
    argc++
    sql += fmt.Sprintf(" AND age > $%d", argc)
    params = append(params, age)
  }

  rows, err := db.Query(sql, params...)
  ...
}
```

It is not super hard, but it is boring, prone to dumb mistakes (adding the wrong parameter, for example) and it will get far worse if the filtering becomes more complex or more filters are added. There are ways to mitigate this by using an ORM, but if you're not a fan of that approach, you will eventually write similar code to the one above.

The thing about the above code is that is a simple checking plus string concatenation, but end up repeated every time just because the arguments are a little different. How about we encapsulate the boring part on a simple construct, `Builder` that does this for us?

```go
func MyFunc(db *sql.DB, department *string, name *string, age *int) (*sql.Rows, error) {
  var b query.Builder
  b.Add("SELECT id,name,age,dept FROM employees WHERE 1=1")
  b.AddIf(" AND dept = ?", department)
  b.AddIf(" AND name = ?", name)
  b.AddIf(" AND age > ?", age)

  rows, err := db.Query(b.String(), b.Params()...)
  ...
}
```

Much easier. And it does *exactly* what the previous code does, nothing more, nothing less. In example above, we could see two functions: `Add`, that *unconditionally* adds the provided string to the buffer (doing parameter substitution, if any is provided) and `AddIf`, that does the same thing but *only if the parameter is non-nil*.

It is important to note that the `query.Builder` does not do anything "magic" behind the scenes; it does not parse your query (you can put anything an he will happily accept), it does not guarantee that the query and/or parameter have valid types (that is on you!), it does not add keywords or whitespace (note the speces added on each line, to make the query valid), etc. Even the **parameter substitution** is dumb as a brick: it just searches for `?` and replaces with `$1`, `$2` and so on.

This is what I like to call the **"no magic allowed principle"**: if you (the developer) is building your queries by hand, you should have full control and libraries should not stay in your way or have unpredictable side effects.

### Extra buffers

The basic example uses `Add` and `AddIf` to build a query, but what if, for some reason, you cannot easily build you wuery in the "correct" order? For example, imagine that a given field will be visible in the `select` clause only if it is also specified as an `order by` option. Again, this is not something *hard* to handle, but it is quite annoying to have multiple checks for the same thing.

For those cases, `query.Builder` also gives you `From`, `Order`, `Where` and `WhereIf` methods. **They behave exactly like the Add/AddIf methods**, the only difference is that they write into different buffers, that are combined in the correct order when you call `String()`. The previous example could be written as:

```go
func MyFunc(db *sql.DB, department *string, name *string, age *int) (*sql.Rows, error) {
  var b query.Builder
  b.Add("SELECT id,name,age,dept")
  b.From(" FROM employees")
  b.Where(" WHERE 1=1")
  b.WhereIf(" AND dept = ?", department)
  b.WhereIf(" AND name = ?", name)
  b.WhereIf(" AND age > ?", age)

  rows, err := db.Query(b.String(), b.Params()...)
  ...
}
```

Again, there is no magic here: the `FROM` and `WHERE` keywords are not magically added and, like before, notice the exta whitespace to build a valid query. It is just `Add`/`AddIf` on different buffers.

## Installation

As with any Go package, you can just `go get github.com/ibraimgm/query` it and use. However, since this package is tiny (about 120 lines), you might as well just copy the source file (`query.go`) directly inside your project instead of adding another dependency. This project aims to have no dependencies outside stdlib, so you wont add any dependency overhead by adding to your project. Just remember to keep the file header intact and give proper attribution where needed.

## Nexts steps

- Add linting and proper version tags
- Option to add whitespace after each command, to avoid common syntax mistakes
- Support for different database syntax (example: MSSQL uses `@` instead of `$` for parameters)
- A way to customize parameter generation and/or substituion
- Additional helpers to call `database/sql` API directly (may be out of the scope...)

## License

BSD-3. Check the `LICENSE` file for details.
