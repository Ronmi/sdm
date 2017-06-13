// package sdm is a simple tool to map go struc to/from SQL table
//
// For those who are busy or lazy: see example to get it done quick.
//
// SDM follows go idiom: adding tags to struct field, SDM parses tags
// and decides what to do.
//
// The format of tag is:
//
//     `sdm:"column_name,property,property,..."`
//
// SDM supports 4 properties:
//
//   - ai:    This column is auto increased. SDM will not pass value to
//            DB when inserting.
//   - pri_:  Specify primary key name. SDM does not validate number of
//            primary key for you.
//   - uniq_: Specify unique key name. You can compose multiple columns
//            to single unique key. Same rules applies to indexes. See
//            example code for how to use it.
//   - idx_:  Specify search index.
//
// As you cn see, SDM does not support foreign key, and possibly never
// support it. ORM is suggested if you need foreign key mapping.
//
// SDM should be safe to use in concurrent environment. Read/write to
// internal data are lock-protected.
package sdm
