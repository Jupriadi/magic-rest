// package readhelper
package magicrest

import (
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Options mengontrol behaviour fungsi ReadPaginated
type Options struct {
	SearchField       string
	OrderBy           string // fallback order if not provided
	PreloadFields     []string
	DefaultFieldTypes map[string]string // e.g. "id":"uuid", "status":"string"
	DefaultPage       int
	DefaultPageSize   int
	AllowGroupBy      bool
}

// Result meta dan data yang dikembalikan
type Result[T any] struct {
	Data []T
	Meta map[string]interface{}
}

// ErrInvalidFilter digunakan bila ada filter tidak valid
var ErrInvalidFilter = errors.New("invalid filter value")

// ReadPaginated: core function yang tidak bergantung gin.
// - query: url.Values (bisa dari request.URL.Query())
// - db: *gorm.DB (sudah di-set model, joins, etc jika perlu dari caller)
// - modelPtr: pointer ke slice/struct model seperti &models.YourModel{} (digunakan untuk scanning)
// Mengembalikan data (slice T), meta (dengan pagination), dan error.
func ReadPaginated[T any](query url.Values, db *gorm.DB, modelPtr *T, opts Options) (Result[T], error) {
	// defaults
	page := opts.DefaultPage
	if page <= 0 {
		page = 1
	}
	pageSize := opts.DefaultPageSize
	if pageSize <= 0 {
		pageSize = 10
	}

	if p := query.Get("page"); p != "" {
		if pi, err := strconv.Atoi(p); err == nil && pi > 0 {
			page = pi
		}
	}
	if ps := query.Get("pageSize"); ps != "" {
		if psi, err := strconv.Atoi(ps); err == nil && psi > 0 {
			pageSize = psi
		}
	}

	search := query.Get("search")
	invalidFilter := false

	// if no custom default provided, use sensible defaults
	defaultAllowed := map[string]string{
		"id":        "uuid",
		"status":    "string",
		"jumlah":    "int",
		"gudang_id": "uuid",
	}
	// merge defaults
	if opts.DefaultFieldTypes == nil {
		opts.DefaultFieldTypes = defaultAllowed
	} else {
		for k, v := range defaultAllowed {
			if _, ok := opts.DefaultFieldTypes[k]; !ok {
				opts.DefaultFieldTypes[k] = v
			}
		}
	}

	// 🔹 Dynamic filters: filter[field]=value
	for key, vals := range query {
		if strings.HasPrefix(key, "filter[") && strings.HasSuffix(key, "]") {
			field := key[7 : len(key)-1]
			if field == "" {
				continue
			}
			value := vals[0]
			fieldType := opts.DefaultFieldTypes[field]

			if strings.Contains(value, ",") {
				parts := strings.Split(value, ",")
				for i := range parts {
					parts[i] = strings.TrimSpace(parts[i])
				}
				switch fieldType {
				case "int":
					var ints []int
					for _, v := range parts {
						iv, err := strconv.Atoi(v)
						if err != nil {
							invalidFilter = true
							continue
						}
						ints = append(ints, iv)
					}
					if len(ints) > 0 {
						db = db.Where(fmt.Sprintf("%s IN ?", field), ints)
					}
				case "uuid":
					var uuids []string
					for _, v := range parts {
						if _, err := uuid.Parse(v); err != nil {
							invalidFilter = true
							continue
						}
						uuids = append(uuids, v)
					}
					if len(uuids) > 0 {
						db = db.Where(fmt.Sprintf("%s IN ?", field), uuids)
					}
				default:
					db = db.Where(fmt.Sprintf("%s IN ?", field), parts)
				}
			} else {
				switch fieldType {
				case "int":
					iv, err := strconv.Atoi(value)
					if err != nil {
						invalidFilter = true
						continue
					}
					db = db.Where(fmt.Sprintf("%s = ?", field), iv)
				case "uuid":
					if _, err := uuid.Parse(value); err != nil {
						invalidFilter = true
						continue
					}
					db = db.Where(fmt.Sprintf("%s = ?", field), value)
				default:
					db = db.Where(fmt.Sprintf("%s = ?", field), value)
				}
			}
		}
	}

	if invalidFilter {
		return Result[T]{Data: []T{}, Meta: map[string]interface{}{}}, ErrInvalidFilter
	}

	// 🔹 Preload (from query ?preload=A,B or from opts)
	if preloadQuery := query.Get("preload"); preloadQuery != "" {
		fields := strings.Split(preloadQuery, ",")
		for _, f := range fields {
			if f = strings.TrimSpace(f); f != "" {
				db = db.Preload(f)
			}
		}
	} else {
		for _, f := range opts.PreloadFields {
			db = db.Preload(f)
		}
	}

	// 🔹 Search
	if opts.SearchField != "" && search != "" {
		if strings.Contains(opts.SearchField, ".") {
			parts := strings.Split(opts.SearchField, ".")
			if len(parts) == 2 {
				relation := parts[0]
				field := parts[1]
				// Note: caller should be responsible untuk JOIN alias yang benar bila perlu.
				db = db.Where(fmt.Sprintf("%s.%s ILIKE ?", relation, field), "%"+search+"%")
			} else {
				db = db.Where(fmt.Sprintf("%s ILIKE ?", opts.SearchField), "%"+search+"%")
			}
		} else {
			db = db.Where(fmt.Sprintf("%s ILIKE ?", opts.SearchField), "%"+search+"%")
		}
	}

	// 🔹 Group by (opsional)
	if opts.AllowGroupBy {
		if gq := query.Get("groupby"); gq != "" {
			fields := strings.Split(gq, ",")
			for i := range fields {
				fields[i] = strings.TrimSpace(fields[i])
			}
			groupExpr := strings.Join(fields, ", ")
			db = db.Select(fmt.Sprintf("%s, MAX(created_at) as created_at", groupExpr)).
				Group(groupExpr).
				Order("MAX(created_at) desc")
		}
	}

	// 🔹 Order by
	orderBy := opts.OrderBy
	if qOrder := query.Get("order"); qOrder != "" {
		orderBy = qOrder
	}
	if orderBy != "" {
		db = db.Order(orderBy)
	} else {
		db = db.Order("created_at desc")
	}

	// 🔹 Paginate (menggunakan helper PaginateGeneric)
	data, pagination, err := PaginateGeneric[T](db, modelPtr, page, pageSize)
	if err != nil {
		return Result[T]{}, err
	}

	return Result[T]{
		Data: data,
		Meta: map[string]interface{}{"pagination": pagination},
	}, nil
}

// ReadPaginatedFromGin: wrapper nyaman untuk pemakai Gin.
// Caller tetap bertanggung jawab mengirim response HTTP.
func ReadPaginatedFromGin[T any](ctxQuery url.Values, db *gorm.DB, modelPtr *T, opts Options) (Result[T], error) {
	return ReadPaginated[T](ctxQuery, db, modelPtr, opts)
}

// PaginateGeneric: contoh implementasi paginate sederhana.
// modelPtr: pointer ke slice model tipe T (e.g. &[]models.User{})
// Mengembalikan []T dan map pagination.
func PaginateGeneric[T any](db *gorm.DB, modelPtr *T, page, pageSize int) ([]T, map[string]interface{}, error) {
	// count total
	var total int64
	countDB := db.Session(&gorm.Session{}) // copy session
	if err := countDB.Count(&total).Error; err != nil {
		return nil, nil, err
	}

	// apply limit offset and find
	offset := (page - 1) * pageSize
	findDB := db.Limit(pageSize).Offset(offset)
	out := new([]T)
	if err := findDB.Find(out).Error; err != nil {
		return nil, nil, err
	}

	// build pagination meta
	totalPages := int((total + int64(pageSize) - 1) / int64(pageSize))
	pagination := map[string]interface{}{
		"page":      page,
		"pageSize":  pageSize,
		"pageCount": totalPages,
		"total":     total,
		"hasNext":   page < totalPages,
		"hasPrev":   page > 1 && totalPages > 0,
	}

	// convert *[]T to []T
	return *out, pagination, nil
}
