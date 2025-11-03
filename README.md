# 🧩 MAGIC REST

**MAGIC REST** is a lightweight Go package that simplifies reading data from databases using [GORM](https://gorm.io) — with built-in support for **dynamic filtering**, **pagination**, **search**, **sorting**, **preloading**, and **grouping**, all through query parameters.

It was inspired by the flexible API style of [Strapi](https://strapi.io), and designed to work seamlessly with [Gin](https://gin-gonic.com) or any Go HTTP framework.

---

## 🚀 Features

- 🔍 **Dynamic Filtering** — use `?filter[field]=value` or `?filter[field]=a,b,c`
- 🔎 **Search Support** — easily add search on any column
- 📄 **Pagination** — controlled via `?page=` and `?pageSize=`
- 🔁 **Preload Relations** — via `?preload=RelationA,RelationB`
- 🧮 **Grouping** — via `?groupby=fieldA,fieldB`
- ↕️ **Sorting** — via `?order=column ASC|DESC`
- ⚙️ **Type-safe Filters** — automatic type validation for `int`, `uuid`, and `string`
- 🧠 **Framework Agnostic** — works with or without Gin

---

## 📦 Installation

```bash
go get github.com/Jupriadi/mgaic-rest@latest 

```

# 🧠 Quick Example (using Gin + GORM)

```bash 
package main

import (
    "github.com/gin-gonic/gin"
    "github.com/Jupriadi/magic-rest"
    "gorm.io/driver/postgres"
    "gorm.io/gorm"
)

type Barang struct {
    ID     string
    Name   string
    Status string
}

func main() {
    // Setup GORM
    dsn := "host=localhost user=postgres password=secret dbname=mydb port=5432 sslmode=disable"
    db, _ := gorm.Open(postgres.Open(dsn), &gorm.Config{})

    // Setup Gin
    r := gin.Default()
    r.GET("/barang", func(ctx *gin.Context) {
        opts := magicrest.Options{
            SearchField: "name",
            OrderBy:     "created_at desc",
            AllowGroupBy: true,
            DefaultFieldTypes: map[string]string{
                "id":     "uuid",
                "status": "string",
            },
        }

        result, err := magicrest.ReadPaginatedFromGin[Barang](
            ctx.Request.URL.Query(),
            db,
            &Barang{},
            opts,
        )

        if err != nil {
            ctx.JSON(400, gin.H{"error": err.Error()})
            return
        }

        ctx.JSON(200, gin.H{
            "data": result.Data,
            "meta": result.Meta,
        })
    })

    r.Run(":8080")
}
```

## ✅ Example requests:

GET /barang?page=1&pageSize=10
GET /barang?filter[status]=active
GET /barang?filter[id]=uuid1,uuid2
GET /barang?search=keyboard
GET /barang?order=name asc
GET /barang?preload=TypeBarang,Kategori
GET /barang?groupby=category_id

⚙️ Configuration Options

magicrest behavior is controlled using the Options struct:
```bash
type Options struct {
    SearchField       string              // Field used for search (optional)
    OrderBy           string              // Default order if not specified
    PreloadFields     []string            // Default preloaded relations
    DefaultFieldTypes map[string]string   // Type map: "uuid", "int", or "string"
    DefaultPage       int                 // Default page (fallback)
    DefaultPageSize   int                 // Default page size (fallback)
    AllowGroupBy      bool                // Enable ?groupby= query
}

🧾 Returned Data Structure

Each call to ReadPaginated or ReadPaginatedFromGin returns:

type Result[T any] struct {
    Data []T
    Meta map[string]interface{}
}


Example response:

{
  "data": [
    { "id": "abc123", "name": "Product A" },
    { "id": "def456", "name": "Product B" }
  ],
  "meta": {
    "pagination": {
      "page": 1,
      "pageSize": 10,
      "pageCount": 5,
      "total": 48,
      "hasNext": true,
      "hasPrev": false
    }
  }
}
```


# 🧩 Query Parameters Overview

```bash
Parameter	Description	Example
page	Page number	?page=2
pageSize	Items per page	?pageSize=20
filter[field]	Filter by field	?filter[status]=active
filter[field] (multi)	Multiple values	?filter[id]=uuid1,uuid2
search	Search by keyword	?search=apple
order	Sorting order	?order=name asc
preload	Preload relations	?preload=Category,Brand
groupby	Group by fields (if enabled)	?groupby=category_id
🧰 Advanced Usage (Non-Gin Example)

If you’re not using Gin, you can still call MagicRest directly:

query := url.Values{
    "page": []string{"1"},
    "filter[status]": []string{"active"},
}

result, err := magicrest.ReadPaginated[Barang](query, db, &Barang{}, magicrest.Options{})
if err != nil {
    log.Fatal(err)
}

fmt.Println(result.Data)
fmt.Println(result.Meta)
```

🪪 License

# MIT License © 2025 Jupriadi

❤️ Contributing

Contributions, ideas, and pull requests are welcome!
Feel free to open an issue or fork the repository.