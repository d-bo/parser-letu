package goldapple

import (
	"time"
)

// A time prefix before collection name
func MakeTimePrefix(coll string) string {
    t := time.Now()
    ti := t.Format("02-01-2006")
    if coll == "" {
        return ti
    }
    fin := ti + "_" + coll
    return fin
}

func MakeTimeMonthlyPrefix(coll string) string {
    t := time.Now()
    ti := t.Format("01-2006")
    fin := ti + "_" + coll
    return fin
}