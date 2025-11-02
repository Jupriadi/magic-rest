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
go get github.com/jupriadi/mgaicrest@latest
