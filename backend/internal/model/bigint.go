package model

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"math/big"

	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

type BigInt struct {
    Int *big.Int
}

func NewBigInt(i *big.Int) BigInt {
    if i == nil {
        return BigInt{Int: big.NewInt(0)}
    }
    return BigInt{Int: new(big.Int).Set(i)}
}

func (b BigInt) Value() (driver.Value, error) {
    if b.Int == nil {
        return "0", nil
    }
    return b.Int.String(), nil
}

func (b *BigInt) Scan(src any) error {
    if b == nil {
        return fmt.Errorf("BigInt: Scan on nil pointer")
    }
    if src == nil {
        b.Int = big.NewInt(0)
        return nil
    }
    switch v := src.(type) {
    case int64:
        b.Int = big.NewInt(v)
        return nil
    case []byte:
        if len(v) == 0 {
            b.Int = big.NewInt(0)
            return nil
        }
        bi := new(big.Int)
        _, ok := bi.SetString(string(v), 10)
        if !ok {
            return fmt.Errorf("BigInt: failed to parse []byte %q", string(v))
        }
        b.Int = bi
        return nil
    case string:
        bi := new(big.Int)
        _, ok := bi.SetString(v, 10)
        if !ok {
            return fmt.Errorf("BigInt: failed to parse string %q", v)
        }
        b.Int = bi
        return nil
    default:
        return fmt.Errorf("BigInt: unsupported scan type %T", src)
    }
}

func (b BigInt) MarshalJSON() ([]byte, error) {
    if b.Int == nil {
        return []byte(`"0"`), nil
    }
    return json.Marshal(b.Int.String())
}

func (b *BigInt) UnmarshalJSON(data []byte) error {
    if b == nil {
        return fmt.Errorf("BigInt: UnmarshalJSON on nil pointer")
    }
    var s string
    if err := json.Unmarshal(data, &s); err == nil {
        bi := new(big.Int)
        _, ok := bi.SetString(s, 10)
        if !ok {
            return fmt.Errorf("BigInt: invalid number %q", s)
        }
        b.Int = bi
        return nil
    }

    var n json.Number
    if err := json.Unmarshal(data, &n); err == nil {
        bi := new(big.Int)
        _, ok := bi.SetString(n.String(), 10)
        if !ok {
            return fmt.Errorf("BigInt: invalid number %q", n.String())
        }
        b.Int = bi
        return nil
    }
    return fmt.Errorf("BigInt: invalid JSON %s", string(data))
}

func (BigInt) GormDataType() string {
    return "numeric"
}

func (BigInt) GormDBDataType(db *gorm.DB, field *schema.Field) string {
    switch db.Dialector.Name() {
    case "postgres":
        return "numeric(78,0)"
    default:
        return "numeric"
    }
}