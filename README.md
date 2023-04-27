# rowstructgen

## Usage

```sql
-- schema.sql
CREATE TABLE `users` (
  `id` BIGINT UNSIGNED NOT NULL,
  `name` VARCHAR(255) NOT NULL,
  `deleted_at` DATETIME NULL,
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARACTER SET=utf8mb4;

```

```sh
$ go run github.com/utgwkk/rowstructgen -schema schema.sql -table users -struct User -package row
```

```go
package row

import "time"

type User struct {
	Id        uint64     `db:"id"`
	Name      string     `db:"name"`
	DeletedAt *time.Time `db:"deleted_at"`
	CreatedAt time.Time  `db:"created_at"`
	UpdatedAt time.Time  `db:"updated_at"`
}
```
